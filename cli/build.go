package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
	"github.com/universal-governance/ugc/engine/parser"
)

const buildJSONSchemaVersion = 1

var buildDryRun bool
var buildRestorePath string
var buildJSON bool

type buildPlanItemJSON struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type buildDryRunJSONOutput struct {
	SchemaVersion int                 `json:"schema_version"`
	DryRun        bool                `json:"dry_run"`
	HasBlockers   bool                `json:"has_blockers"`
	Items         []buildPlanItemJSON `json:"items"`
	Summary       map[string]int      `json:"summary"`
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile governance rules into agent targets",
	Long:  `Reads the agnostic governance files and transpiles them into agent-specific configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if buildJSON && buildRestorePath != "" {
			return fmt.Errorf("cannot combine --json with --restore")
		}

		if !buildJSON {
			fmt.Fprintln(out, "Building governance targets...")
		}

		gov, err := parser.Parse(".universal-governance")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing governance files:", err)
			return err
		}

		targetDir := "."
		generatedDir, err := os.MkdirTemp("", "ugc-build-*")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating build temp dir:", err)
			return err
		}
		defer os.RemoveAll(generatedDir)

		for _, e := range engine.V1Emitters() {
			if err := e.Emit(gov, generatedDir); err != nil {
				fmt.Fprintln(os.Stderr, "Error emitting target to temp dir:", err)
				return err
			}
		}

		if buildRestorePath != "" {
			if buildDryRun {
				fmt.Fprintln(os.Stderr, "Cannot combine --restore with --dry-run.")
				os.Exit(1)
			}
			if err := engine.RestoreGeneratedArtifact(targetDir, generatedDir, gov.SourceHash, buildRestorePath); err != nil {
				if errors.Is(err, engine.ErrRestoreRefused) {
					fmt.Fprintln(os.Stderr, "Restore refused:", err)
				} else {
					fmt.Fprintln(os.Stderr, "Restore failed:", err)
				}
				return err
			}
			fmt.Fprintf(out, "Restored generated artifact: %s\n", buildRestorePath)
			return nil
		}

		plan, err := engine.PlanGeneratedBuild(targetDir, generatedDir, gov.SourceHash)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error planning build:", err)
			return err
		}

		if buildJSON && buildDryRun {
			if err := printBuildPlanJSON(out, plan, true); err != nil {
				return err
			}
			if plan.HasBlockers() {
				return fmt.Errorf("build dry-run blocked")
			}
			return nil
		}

		if !buildJSON {
			printBuildPlan(out, plan)
		}

		if buildDryRun {
			if !buildJSON {
				fmt.Fprintln(out, "Dry run complete. No files written.")
			}
			return nil
		}

		if plan.HasBlockers() {
			fmt.Fprintln(os.Stderr, "Build blocked: unmanaged generated artifacts would be overwritten.")
			for _, item := range plan.BlockedItems() {
				fmt.Fprintf(os.Stderr, "- %s: %s\n", item.Path, item.Reason)
			}
			os.Exit(1)
		}

		if err := engine.ApplyBuildPlan(targetDir, generatedDir, plan); err != nil {
			if errors.Is(err, engine.ErrBuildPlanBlocked) {
				fmt.Fprintln(os.Stderr, "Build blocked: unmanaged generated artifacts would be overwritten.")
			} else if errors.Is(err, engine.ErrBuildRollbackFailed) {
				fmt.Fprintln(os.Stderr, "Build failed during apply and rollback was incomplete. Run ugc audit.")
				fmt.Fprintln(os.Stderr, err)
			} else if errors.Is(err, engine.ErrBuildApplyFailed) {
				fmt.Fprintln(os.Stderr, "Build failed during apply; rollback completed. No clean manifest was written.")
				fmt.Fprintln(os.Stderr, err)
			} else {
				fmt.Fprintln(os.Stderr, "Error applying build plan:", err)
			}
			return err
		}

		if buildJSON {
			if err := printBuildPlanJSON(out, plan, false); err != nil {
				return err
			}
			return nil
		}

		fmt.Fprintln(out, "Build complete.")
		return nil
	},
}

func printBuildPlan(out io.Writer, plan engine.BuildPlan) {
	fmt.Fprintln(out, "Build plan:")
	for _, item := range plan.Items {
		if item.Reason == "" {
			fmt.Fprintf(out, "- %s: %s\n", item.Status, item.Path)
			continue
		}
		fmt.Fprintf(out, "- %s: %s (%s)\n", item.Status, item.Path, item.Reason)
	}
}

func buildPlanJSON(plan engine.BuildPlan, dryRun bool) buildDryRunJSONOutput {
	items := make([]buildPlanItemJSON, 0, len(plan.Items))
	summary := stableBuildSummary(plan)
	for _, item := range plan.Items {
		status := string(item.Status)
		items = append(items, buildPlanItemJSON{
			Path:   item.Path,
			Status: status,
			Reason: item.Reason,
		})
	}
	return buildDryRunJSONOutput{
		SchemaVersion: buildJSONSchemaVersion,
		DryRun:        dryRun,
		HasBlockers:   plan.HasBlockers(),
		Items:         items,
		Summary:       summary,
	}
}

func stableBuildSummary(plan engine.BuildPlan) map[string]int {
	summary := map[string]int{
		string(engine.BuildStatusCreate):           0,
		string(engine.BuildStatusUnchanged):       0,
		string(engine.BuildStatusManagedOverwrite): 0,
		string(engine.BuildStatusBlockedUnmanaged): 0,
	}
	for _, item := range plan.Items {
		summary[string(item.Status)]++
	}
	return summary
}

func printBuildPlanJSON(out io.Writer, plan engine.BuildPlan, dryRun bool) error {
	payload := buildPlanJSON(plan, dryRun)
	if !dryRun {
		payload.HasBlockers = false
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func init() {
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "Preview generated target writes without modifying files")
	buildCmd.Flags().StringVar(&buildRestorePath, "restore", "", "Restore one manifest-owned generated artifact path")
	buildCmd.Flags().BoolVar(&buildJSON, "json", false, "Print machine-readable build plan or apply result")
	rootCmd.AddCommand(buildCmd)
}
