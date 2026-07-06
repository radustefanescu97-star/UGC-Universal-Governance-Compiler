package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const sourceHashMarker = "UGC-Source-Hash:"
const markerScanBytes = 4096

var ErrBuildPlanBlocked = errors.New("build blocked by unmanaged generated artifacts")
var ErrBuildApplyFailed = errors.New("build apply failed and rollback succeeded")
var ErrBuildRollbackFailed = errors.New("build apply failed and rollback was incomplete")
var ErrRestoreRefused = errors.New("restore refused")

type BuildPlanStatus string

const (
	BuildStatusCreate           BuildPlanStatus = "create"
	BuildStatusUnchanged        BuildPlanStatus = "unchanged"
	BuildStatusManagedOverwrite BuildPlanStatus = "managed-overwrite"
	BuildStatusBlockedUnmanaged BuildPlanStatus = "blocked-unmanaged"
)

type BuildPlanItem struct {
	Path   string
	Status BuildPlanStatus
	Reason string
}

type BuildPlan struct {
	Items    []BuildPlanItem
	Manifest BuildManifest
}

type BuildApplyError struct {
	ApplyErr       error
	RollbackErrors []error
}

func (e *BuildApplyError) Error() string {
	if len(e.RollbackErrors) > 0 {
		return fmt.Sprintf("%v: %v; rollback errors: %v", ErrBuildRollbackFailed, e.ApplyErr, e.RollbackErrors)
	}
	return fmt.Sprintf("%v: %v", ErrBuildApplyFailed, e.ApplyErr)
}

func (e *BuildApplyError) Unwrap() error {
	if len(e.RollbackErrors) > 0 {
		return ErrBuildRollbackFailed
	}
	return ErrBuildApplyFailed
}

func (p BuildPlan) HasBlockers() bool {
	for _, item := range p.Items {
		if item.Status == BuildStatusBlockedUnmanaged {
			return true
		}
	}
	return false
}

func (p BuildPlan) BlockedItems() []BuildPlanItem {
	var blocked []BuildPlanItem
	for _, item := range p.Items {
		if item.Status == BuildStatusBlockedUnmanaged {
			blocked = append(blocked, item)
		}
	}
	return blocked
}

func PlanGeneratedBuild(rootDir, generatedDir, sourceHash string) (BuildPlan, error) {
	manifest, err := BuildManifestForOutputs(generatedDir, sourceHash)
	if err != nil {
		return BuildPlan{}, err
	}

	priorManifest, hasPriorManifest := readPriorBuildManifest(rootDir)
	plan := BuildPlan{Manifest: manifest}

	for _, artifact := range manifest.Artifacts {
		generatedPath := filepath.Join(generatedDir, filepath.FromSlash(artifact.Path))
		generatedData, err := os.ReadFile(generatedPath)
		if err != nil {
			return BuildPlan{}, err
		}

		actualPath := filepath.Join(rootDir, filepath.FromSlash(artifact.Path))
		actualData, err := os.ReadFile(actualPath)
		if os.IsNotExist(err) {
			plan.Items = append(plan.Items, BuildPlanItem{
				Path:   artifact.Path,
				Status: BuildStatusCreate,
				Reason: "target file does not exist",
			})
			continue
		}
		if err != nil {
			return BuildPlan{}, err
		}

		switch {
		case bytes.Equal(actualData, generatedData):
			plan.Items = append(plan.Items, BuildPlanItem{
				Path:   artifact.Path,
				Status: BuildStatusUnchanged,
				Reason: "target file already matches generated output",
			})
		case hasSourceHashMarker(actualData):
			plan.Items = append(plan.Items, BuildPlanItem{
				Path:   artifact.Path,
				Status: BuildStatusManagedOverwrite,
				Reason: "existing file contains UGC source-hash marker",
			})
		case hasPriorManifestMatch(priorManifest, hasPriorManifest, artifact.Path, actualPath):
			plan.Items = append(plan.Items, BuildPlanItem{
				Path:   artifact.Path,
				Status: BuildStatusManagedOverwrite,
				Reason: "existing file matches prior build manifest",
			})
		default:
			plan.Items = append(plan.Items, BuildPlanItem{
				Path:   artifact.Path,
				Status: BuildStatusBlockedUnmanaged,
				Reason: "existing file is not UGC-managed",
			})
		}
	}

	return plan, nil
}

func ApplyBuildPlan(rootDir, generatedDir string, plan BuildPlan) error {
	return applyBuildPlan(rootDir, generatedDir, plan, buildPlanIO{})
}

func RestoreGeneratedArtifact(rootDir, generatedDir, sourceHash, requestedPath string) error {
	return restoreGeneratedArtifact(rootDir, generatedDir, sourceHash, requestedPath, buildPlanIO{})
}

type buildPlanIO struct {
	mkdirTemp          func(dir, pattern string) (string, error)
	removeAll          func(path string) error
	stat               func(name string) (os.FileInfo, error)
	readFile           func(name string) ([]byte, error)
	mkdirAll           func(path string, perm os.FileMode) error
	writeFileAtomic    func(filename string, data []byte, perm os.FileMode) error
	remove             func(name string) error
	writeBuildManifest func(rootDir string, manifest BuildManifest) error
}

type buildOperation struct {
	item       BuildPlanItem
	actualPath string
	data       []byte
	backup     buildBackup
}

type buildBackup struct {
	existed    bool
	backupPath string
	mode       os.FileMode
}

func applyBuildPlan(rootDir, generatedDir string, plan BuildPlan, io buildPlanIO) error {
	io = io.withDefaults()

	if plan.HasBlockers() {
		return ErrBuildPlanBlocked
	}

	transactionDir, err := io.mkdirTemp("", "ugc-build-transaction-*")
	if err != nil {
		return err
	}
	defer io.removeAll(transactionDir)

	operations, err := prepareBuildOperations(rootDir, generatedDir, transactionDir, plan, io)
	if err != nil {
		return err
	}

	var applied []buildOperation
	for _, operation := range operations {
		if err := io.mkdirAll(filepath.Dir(operation.actualPath), 0755); err != nil {
			return rollbackBuildOperations(applied, err, io)
		}
		if err := io.writeFileAtomic(operation.actualPath, operation.data, 0644); err != nil {
			return rollbackBuildOperations(applied, err, io)
		}
		applied = append(applied, operation)
	}

	if err := io.writeBuildManifest(rootDir, plan.Manifest); err != nil {
		return rollbackBuildOperations(applied, err, io)
	}

	return nil
}

func restoreGeneratedArtifact(rootDir, generatedDir, sourceHash, requestedPath string, io buildPlanIO) error {
	io = io.withDefaults()

	relPath, err := normalizeRestorePath(requestedPath)
	if err != nil {
		return err
	}

	priorManifest, err := ReadBuildManifest(rootDir)
	if err != nil {
		return fmt.Errorf("%w: prior build manifest is required: %v", ErrRestoreRefused, err)
	}
	if priorManifest.SourceHash != sourceHash {
		return fmt.Errorf("%w: source hash changed since prior build; run normal build", ErrRestoreRefused)
	}
	if !manifestContainsPath(priorManifest, relPath) {
		return fmt.Errorf("%w: path is not owned by the prior build manifest: %s", ErrRestoreRefused, relPath)
	}

	expectedManifest, err := BuildManifestForOutputs(generatedDir, sourceHash)
	if err != nil {
		return err
	}
	if !manifestContainsPath(expectedManifest, relPath) {
		return fmt.Errorf("%w: path is not in the current expected artifact set: %s", ErrRestoreRefused, relPath)
	}

	generatedPath := filepath.Join(generatedDir, filepath.FromSlash(relPath))
	generatedInfo, err := io.stat(generatedPath)
	if err != nil {
		return err
	}
	if generatedInfo.IsDir() {
		return fmt.Errorf("%w: generated path is a directory: %s", ErrRestoreRefused, relPath)
	}
	data, err := io.readFile(generatedPath)
	if err != nil {
		return err
	}

	actualPath := filepath.Join(rootDir, filepath.FromSlash(relPath))
	if info, err := io.stat(actualPath); err == nil && info.IsDir() {
		return fmt.Errorf("%w: target path is a directory: %s", ErrRestoreRefused, relPath)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := io.mkdirAll(filepath.Dir(actualPath), 0755); err != nil {
		return err
	}
	if err := io.writeFileAtomic(actualPath, data, 0644); err != nil {
		return err
	}
	return nil
}

func normalizeRestorePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: restore path is required", ErrRestoreRefused)
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("%w: absolute paths are not allowed", ErrRestoreRefused)
	}
	if strings.ContainsAny(path, "*?[") {
		return "", fmt.Errorf("%w: glob patterns are not allowed", ErrRestoreRefused)
	}
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: directory paths are not allowed", ErrRestoreRefused)
	}

	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("%w: path traversal is not allowed", ErrRestoreRefused)
	}
	for _, segment := range strings.Split(clean, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("%w: invalid restore path", ErrRestoreRefused)
		}
	}
	return clean, nil
}

func manifestContainsPath(manifest BuildManifest, path string) bool {
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == path {
			return true
		}
	}
	return false
}

func prepareBuildOperations(rootDir, generatedDir, transactionDir string, plan BuildPlan, io buildPlanIO) ([]buildOperation, error) {
	var operations []buildOperation
	for _, item := range plan.Items {
		switch item.Status {
		case BuildStatusUnchanged:
			continue
		case BuildStatusCreate, BuildStatusManagedOverwrite:
			generatedPath := filepath.Join(generatedDir, filepath.FromSlash(item.Path))
			data, err := io.readFile(generatedPath)
			if err != nil {
				return nil, err
			}
			actualPath := filepath.Join(rootDir, filepath.FromSlash(item.Path))

			backup := buildBackup{mode: 0644}
			info, err := io.stat(actualPath)
			switch {
			case err == nil:
				if info.IsDir() {
					return nil, fmt.Errorf("target path is a directory: %s", item.Path)
				}
				backup.existed = true
				backup.mode = info.Mode().Perm()
				backup.backupPath = filepath.Join(transactionDir, fmt.Sprintf("%06d.backup", len(operations)))
				existingData, err := io.readFile(actualPath)
				if err != nil {
					return nil, err
				}
				if err := io.writeFileAtomic(backup.backupPath, existingData, backup.mode); err != nil {
					return nil, err
				}
			case os.IsNotExist(err):
				backup.existed = false
			default:
				return nil, err
			}

			operations = append(operations, buildOperation{
				item:       item,
				actualPath: actualPath,
				data:       data,
				backup:     backup,
			})
		case BuildStatusBlockedUnmanaged:
			return nil, ErrBuildPlanBlocked
		default:
			return nil, fmt.Errorf("unknown build plan status %q for %s", item.Status, item.Path)
		}
	}

	return operations, nil
}

func rollbackBuildOperations(applied []buildOperation, applyErr error, io buildPlanIO) error {
	var rollbackErrors []error

	for i := len(applied) - 1; i >= 0; i-- {
		operation := applied[i]
		if operation.backup.existed {
			data, err := io.readFile(operation.backup.backupPath)
			if err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("%s backup read: %w", operation.item.Path, err))
				continue
			}
			if err := io.mkdirAll(filepath.Dir(operation.actualPath), 0755); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("%s restore mkdir: %w", operation.item.Path, err))
				continue
			}
			if err := io.writeFileAtomic(operation.actualPath, data, operation.backup.mode); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("%s restore write: %w", operation.item.Path, err))
			}
			continue
		}

		if err := io.remove(operation.actualPath); err != nil && !os.IsNotExist(err) {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("%s remove created file: %w", operation.item.Path, err))
		}
	}

	return &BuildApplyError{ApplyErr: applyErr, RollbackErrors: rollbackErrors}
}

func (io buildPlanIO) withDefaults() buildPlanIO {
	if io.mkdirTemp == nil {
		io.mkdirTemp = os.MkdirTemp
	}
	if io.removeAll == nil {
		io.removeAll = os.RemoveAll
	}
	if io.stat == nil {
		io.stat = os.Stat
	}
	if io.readFile == nil {
		io.readFile = os.ReadFile
	}
	if io.mkdirAll == nil {
		io.mkdirAll = os.MkdirAll
	}
	if io.writeFileAtomic == nil {
		io.writeFileAtomic = atomicWriteFile
	}
	if io.remove == nil {
		io.remove = os.Remove
	}
	if io.writeBuildManifest == nil {
		io.writeBuildManifest = WriteBuildManifest
	}
	return io
}

func readPriorBuildManifest(rootDir string) (BuildManifest, bool) {
	manifest, err := ReadBuildManifest(rootDir)
	if err != nil {
		return BuildManifest{}, false
	}
	return manifest, true
}

func hasSourceHashMarker(data []byte) bool {
	if len(data) > markerScanBytes {
		data = data[:markerScanBytes]
	}
	return bytes.Contains(data, []byte(sourceHashMarker))
}

func hasPriorManifestMatch(manifest BuildManifest, ok bool, relPath, actualPath string) bool {
	if !ok {
		return false
	}
	actualHash, err := hashPath(actualPath)
	if err != nil {
		return false
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == relPath && artifact.SHA256 == actualHash {
			return true
		}
	}
	return false
}
