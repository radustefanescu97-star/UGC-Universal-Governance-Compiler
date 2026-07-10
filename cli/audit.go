package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

const auditJSONSchemaVersion = 1

var auditJSON bool

type auditDriftJSON struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type auditJSONOutput struct {
	SchemaVersion       int                          `json:"schema_version"`
	AuditPassed         bool                         `json:"audit_passed"`
	SourceValid         bool                         `json:"source_valid"`
	SourceErrors        []string                     `json:"source_errors"`
	SourceHash          string                       `json:"source_hash,omitempty"`
	Drift               []auditDriftJSON             `json:"drift"`
	UnexpectedArtifacts []string                     `json:"unexpected_artifacts"`
	ManifestFindings    []string                     `json:"manifest_findings"`
	CorpusState         string                       `json:"corpus_state"`
	CorpusStateMessage  string                       `json:"corpus_state_message"`
	CapabilityCoverage  map[string]map[string]string `json:"capability_coverage"`
	ExpectedArtifacts   []string                     `json:"expected_artifacts"`
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Verify structural validity and file drift",
	Long:  `Ensures structural validity, verifies file drift, checks target capability coverage, and reports on corpus state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if !auditJSON {
			fmt.Fprintln(out, "Auditing governance configurations...")
		}

		result, err := engine.AuditProject(".")
		if err != nil {
			if auditJSON {
				if writeErr := printAuditJSON(cmd.OutOrStdout(), result, true); writeErr != nil {
					return writeErr
				}
			}
			fmt.Fprintln(os.Stderr, "Audit error:", err)
			return err
		}

		if auditJSON {
			if err := printAuditJSON(out, result, false); err != nil {
				return err
			}
			if result.Failed() {
				return fmt.Errorf("audit failed")
			}
			return nil
		}

		printAuditHuman(out, result)
		if result.Failed() {
			fmt.Fprintln(os.Stderr, "Audit failed: Drift detected in generated files.")
			os.Exit(1)
		}
		return nil
	},
}

func printAuditHuman(out io.Writer, result engine.AuditResult) {
	if len(result.SourceErrors) > 0 {
		for _, sourceErr := range result.SourceErrors {
			fmt.Fprintf(os.Stderr, "Source validity failed: %s\n", sourceErr)
		}
	} else {
		fmt.Fprintln(out, "Source validity: ok")
	}
	if result.SourceHash != "" {
		fmt.Fprintf(out, "Source Hash: %s\n", result.SourceHash)
	}

	for _, drift := range result.Drift {
		fmt.Fprintf(out, "Drift detected: %s\n", drift.Message)
	}
	for _, path := range result.UnexpectedArtifacts {
		fmt.Fprintf(out, "Drift detected: Unexpected generated artifact %s\n", path)
	}
	for _, finding := range result.ManifestFindings {
		fmt.Fprintf(out, "Build manifest failed: %s\n", finding)
	}

	switch result.CorpusState {
	case "ok":
		fmt.Fprintln(out, "Corpus state: ok")
	case "missing", "legacy":
		fmt.Fprintf(out, "Corpus state: %s\n", result.CorpusStateMessage)
	case "failed":
		fmt.Fprintln(os.Stderr, "Corpus state failed:", result.CorpusStateMessage)
	}

	if result.Failed() {
		return
	}

	fmt.Fprintln(out, "Target capability coverage:")
	fmt.Fprintln(out, result.CapabilitySummary)
	fmt.Fprintln(out, "Audit complete. No drift detected.")
}

func buildAuditJSON(result engine.AuditResult) auditJSONOutput {
	drift := make([]auditDriftJSON, 0, len(result.Drift))
	for _, item := range result.Drift {
		drift = append(drift, auditDriftJSON{
			Path:    item.Path,
			Message: item.Message,
		})
	}

	unexpected := result.UnexpectedArtifacts
	if unexpected == nil {
		unexpected = []string{}
	}
	manifestFindings := result.ManifestFindings
	if manifestFindings == nil {
		manifestFindings = []string{}
	}
	sourceErrors := result.SourceErrors
	if sourceErrors == nil {
		sourceErrors = []string{}
	}
	expectedArtifacts := result.ExpectedArtifacts
	if expectedArtifacts == nil {
		expectedArtifacts = []string{}
	}

	return auditJSONOutput{
		SchemaVersion:       auditJSONSchemaVersion,
		AuditPassed:         !result.Failed(),
		SourceValid:         len(result.SourceErrors) == 0,
		SourceErrors:        sourceErrors,
		SourceHash:          result.SourceHash,
		Drift:               drift,
		UnexpectedArtifacts: unexpected,
		ManifestFindings:    manifestFindings,
		CorpusState:         normalizeCorpusState(result.CorpusState),
		CorpusStateMessage:  result.CorpusStateMessage,
		CapabilityCoverage:  engine.TargetCapabilityMatrix(),
		ExpectedArtifacts:   expectedArtifacts,
	}
}

func normalizeCorpusState(state string) string {
	switch state {
	case "ok", "missing", "legacy", "failed", "unknown":
		return state
	case "":
		return "unknown"
	default:
		return "unknown"
	}
}

func printAuditJSON(out io.Writer, result engine.AuditResult, auditErrored bool) error {
	payload := buildAuditJSON(result)
	if auditErrored {
		payload.AuditPassed = false
		payload.SourceValid = false
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func init() {
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "Print machine-readable audit results")
	rootCmd.AddCommand(auditCmd)
}
