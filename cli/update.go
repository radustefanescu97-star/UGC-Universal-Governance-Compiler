package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

const updateJSONSchemaVersion = 1

var dryRun bool
var updateJSON bool

type updateSummaryCountsJSON struct {
	Created                 int `json:"created"`
	Updated                 int `json:"updated"`
	Unchanged               int `json:"unchanged"`
	SkippedLocalEdits       int `json:"skipped_local_edits"`
	SkippedUnverifiedLegacy int `json:"skipped_unverified_legacy"`
	Failed                  int `json:"failed"`
}

type updateDryRunJSONOutput struct {
	SchemaVersion int                     `json:"schema_version"`
	DryRun        bool                    `json:"dry_run"`
	StateWarning  string                  `json:"state_warning,omitempty"`
	Summary       updateSummaryCountsJSON `json:"summary"`
	Created       []string                `json:"created"`
	Updated       []string                `json:"updated"`
	Unchanged     []string                `json:"unchanged"`
	SkippedLocalEdits       []string      `json:"skipped_local_edits"`
	SkippedUnverifiedLegacy []string      `json:"skipped_unverified_legacy"`
	Failed        []string                `json:"failed"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local Standard Corpus",
	Long:  `Synchronize local .universal-governance/ directories with the compiler version while preserving local edits through drift detection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateJSON && !dryRun {
			return fmt.Errorf("update --json requires --dry-run")
		}

		if updateJSON {
			summary, err := engine.PlanCorpusUpdateDryRun()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Update failed:", err)
				return err
			}
			if err := printUpdateDryRunJSON(cmd.OutOrStdout(), summary); err != nil {
				return err
			}
			return nil
		}

		if err := engine.UpdateCorpus(dryRun); err != nil {
			fmt.Fprintln(os.Stderr, "Update failed:", err)
			return err
		}
		return nil
	},
}

func buildUpdateDryRunJSON(summary engine.UpdateSummary) updateDryRunJSONOutput {
	return updateDryRunJSONOutput{
		SchemaVersion: updateJSONSchemaVersion,
		DryRun:        true,
		StateWarning:  summary.StateWarning,
		Summary: updateSummaryCountsJSON{
			Created:                 len(summary.Created),
			Updated:                 len(summary.Updated),
			Unchanged:               len(summary.Unchanged),
			SkippedLocalEdits:       len(summary.SkippedLocalEdits),
			SkippedUnverifiedLegacy: len(summary.SkippedUnverifiedLegacy),
			Failed:                  len(summary.Failed),
		},
		Created:                 nonNilStrings(summary.Created),
		Updated:                 nonNilStrings(summary.Updated),
		Unchanged:               nonNilStrings(summary.Unchanged),
		SkippedLocalEdits:       nonNilStrings(summary.SkippedLocalEdits),
		SkippedUnverifiedLegacy: nonNilStrings(summary.SkippedUnverifiedLegacy),
		Failed:                  nonNilStrings(summary.Failed),
	}
}

func nonNilStrings(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

func printUpdateDryRunJSON(out io.Writer, summary engine.UpdateSummary) error {
	payload := buildUpdateDryRunJSON(summary)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate update without modifying files")
	updateCmd.Flags().BoolVar(&updateJSON, "json", false, "Print machine-readable dry-run update plan (requires --dry-run)")
	rootCmd.AddCommand(updateCmd)
}
