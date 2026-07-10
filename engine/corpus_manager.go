package engine

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/universal-governance/ugc/core_embed"
)

const GovernanceDir = ".universal-governance"
const StateFile = ".universal-governance/.state.json"
const StateSchemaVersion = 1

var ErrStateMissing = errors.New("state file missing")
var ErrStateLegacy = errors.New("state file uses legacy schema")

type CorpusState struct {
	SchemaVersion int               `json:"schema_version"`
	CorpusVersion string            `json:"corpus_version"`
	Hashes        map[string]string `json:"hashes"`
}

type UpdateSummary struct {
	StateWarning            string
	Created                 []string
	Updated                 []string
	Unchanged               []string
	SkippedLocalEdits       []string
	SkippedUnverifiedLegacy []string
	Failed                  []string
}

type corpusUpdateIO struct {
	writeFileAtomic func(filename string, data []byte, perm os.FileMode) error
	saveState       func(CorpusState) error
}

// LoadState reads and validates the official hashes from the state file.
func LoadState() (CorpusState, error) {
	return loadStateAt(StateFile)
}

func LoadStateForRoot(rootDir string) (CorpusState, error) {
	return loadStateAt(filepath.Join(rootDir, StateFile))
}

func loadStateAt(path string) (CorpusState, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return CorpusState{SchemaVersion: StateSchemaVersion, CorpusVersion: "1.0.0", Hashes: map[string]string{}}, ErrStateMissing
	}
	if err != nil {
		return CorpusState{}, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return CorpusState{}, fmt.Errorf("invalid state file JSON: %w", err)
	}
	if _, ok := raw["schema_version"]; !ok {
		return CorpusState{SchemaVersion: 0, CorpusVersion: "legacy", Hashes: map[string]string{}}, ErrStateLegacy
	}

	var state CorpusState
	if err := json.Unmarshal(data, &state); err != nil {
		return CorpusState{}, fmt.Errorf("invalid state file schema: %w", err)
	}
	if state.SchemaVersion != StateSchemaVersion {
		return CorpusState{}, fmt.Errorf("unsupported state schema version %d", state.SchemaVersion)
	}
	if state.Hashes == nil {
		state.Hashes = map[string]string{}
	}
	if state.CorpusVersion == "" {
		state.CorpusVersion = "1.0.0"
	}
	return state, nil
}

// SaveState writes the official hashes to the state file.
func SaveState(s CorpusState) error {
	return saveStateAt(StateFile, s)
}

func saveStateAt(path string, s CorpusState) error {
	if s.SchemaVersion == 0 {
		s.SchemaVersion = StateSchemaVersion
	}
	if s.CorpusVersion == "" {
		s.CorpusVersion = "1.0.0"
	}
	if s.Hashes == nil {
		s.Hashes = map[string]string{}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0644)
}

// hashBytes computes the SHA256 of a byte slice.
func hashBytes(data []byte) string {
	return fmt.Sprintf("%x", sha256Sum(data))
}

func sha256Sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile)
		return err
	}
	return nil
}

// UpdateCorpus performs drift detection and updates the local governance folder.
func UpdateCorpus(dryRun bool) error {
	_, err := updateCorpus(dryRun, false, corpusUpdateIO{})
	return err
}

// PlanCorpusUpdateDryRun previews corpus synchronization without writing files or human output.
func PlanCorpusUpdateDryRun() (UpdateSummary, error) {
	return updateCorpus(true, true, corpusUpdateIO{})
}

func updateCorpus(dryRun bool, quiet bool, io corpusUpdateIO) (UpdateSummary, error) {
	summary := UpdateSummary{}
	if io.writeFileAtomic == nil {
		io.writeFileAtomic = atomicWriteFile
	}
	if io.saveState == nil {
		io.saveState = SaveState
	}

	if !dryRun {
		if err := os.MkdirAll(GovernanceDir, 0755); err != nil {
			return summary, err
		}
	}

	state, stateErr := LoadState()
	stateTrusted := true
	switch {
	case stateErr == nil:
	case errors.Is(stateErr, ErrStateMissing):
		stateTrusted = false
		summary.StateWarning = "State: missing .state.json; existing changed files will be treated as unverified legacy."
		if !quiet {
			fmt.Println(summary.StateWarning)
		}
	case errors.Is(stateErr, ErrStateLegacy):
		stateTrusted = false
		summary.StateWarning = "State: legacy .state.json schema; existing changed files will be treated as unverified legacy."
		if !quiet {
			fmt.Println(summary.StateWarning)
		}
	default:
		return summary, stateErr
	}

	err := fs.WalkDir(core_embed.StandardCorpus, "standard_corpus", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel("standard_corpus", path)
		localPath := filepath.Join(GovernanceDir, relPath)

		embedData, err := fs.ReadFile(core_embed.StandardCorpus, path)
		if err != nil {
			return err
		}
		embedHash := hashBytes(embedData)

		// Check if local file exists
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			if !dryRun {
				if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
					summary.Failed = append(summary.Failed, relPath)
					return err
				}
				if err := io.writeFileAtomic(localPath, embedData, 0644); err != nil {
					summary.Failed = append(summary.Failed, relPath)
					return err
				}
				state.Hashes[relPath] = embedHash
			}
			summary.Created = append(summary.Created, relPath)
			if !quiet {
				fmt.Printf("[NEW] %s\n", localPath)
			}
			return nil
		}

		// Local file exists, check drift
		localHash, err := hashPath(localPath)
		if err != nil {
			summary.Failed = append(summary.Failed, relPath)
			return err
		}

		// If it's perfectly matching the new official version already, just update state.
		if localHash == embedHash {
			if !dryRun {
				state.Hashes[relPath] = embedHash
			}
			summary.Unchanged = append(summary.Unchanged, relPath)
			return nil
		}

		originalHash, hasOriginal := state.Hashes[relPath]
		if !stateTrusted || !hasOriginal {
			summary.SkippedUnverifiedLegacy = append(summary.SkippedUnverifiedLegacy, relPath)
			return nil
		}

		if localHash == originalHash {
			if !dryRun {
				if err := io.writeFileAtomic(localPath, embedData, 0644); err != nil {
					summary.Failed = append(summary.Failed, relPath)
					return err
				}
				state.Hashes[relPath] = embedHash
			}
			summary.Updated = append(summary.Updated, relPath)
			if !quiet {
				fmt.Printf("[UPDATED] %s\n", localPath)
			}
		} else {
			summary.SkippedLocalEdits = append(summary.SkippedLocalEdits, relPath)
		}

		return nil
	})

	if err != nil {
		return summary, err
	}

	if !dryRun {
		if err := io.saveState(state); err != nil {
			return summary, err
		}
	}

	if !quiet {
		printUpdateCorpusHuman(summary, dryRun)
	}

	return summary, nil
}

func printUpdateCorpusHuman(summary UpdateSummary, dryRun bool) {
	if dryRun {
		fmt.Printf("Dry run complete: %d created, %d updated, %d unchanged, %d skipped-local-edits, %d skipped-unverified-legacy, %d failed.\n",
			len(summary.Created), len(summary.Updated), len(summary.Unchanged), len(summary.SkippedLocalEdits), len(summary.SkippedUnverifiedLegacy), len(summary.Failed))
	} else {
		fmt.Printf("Update complete: %d created, %d updated, %d unchanged, %d skipped-local-edits, %d skipped-unverified-legacy, %d failed.\n",
			len(summary.Created), len(summary.Updated), len(summary.Unchanged), len(summary.SkippedLocalEdits), len(summary.SkippedUnverifiedLegacy), len(summary.Failed))
	}

	if len(summary.SkippedLocalEdits) > 0 {
		fmt.Println("\nWARNING: The following files have local edits and were not overwritten:")
		for _, f := range summary.SkippedLocalEdits {
			fmt.Printf("- %s\n", f)
		}
	}
	if len(summary.SkippedUnverifiedLegacy) > 0 {
		fmt.Println("\nWARNING: The following files cannot be verified against versioned state and were not overwritten:")
		for _, f := range summary.SkippedUnverifiedLegacy {
			fmt.Printf("- %s\n", f)
		}
	}
}
