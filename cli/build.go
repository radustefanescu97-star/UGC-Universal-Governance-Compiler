package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
	"github.com/universal-governance/ugc/engine/parser"
)

var buildDryRun bool
var buildRestorePath string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile governance rules into agent targets",
	Long:  `Reads the agnostic governance files and transpiles them into agent-specific configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building governance targets...")
		gov, err := parser.Parse(".universal-governance")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing governance files:", err)
			os.Exit(1)
		}

		targetDir := "."
		generatedDir, err := os.MkdirTemp("", "ugc-build-*")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating build temp dir:", err)
			os.Exit(1)
		}
		defer os.RemoveAll(generatedDir)

		for _, e := range engine.V1Emitters() {
			if err := e.Emit(gov, generatedDir); err != nil {
				fmt.Fprintln(os.Stderr, "Error emitting target to temp dir:", err)
				os.Exit(1)
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
				os.Exit(1)
			}
			fmt.Printf("Restored generated artifact: %s\n", buildRestorePath)
			return
		}

		plan, err := engine.PlanGeneratedBuild(targetDir, generatedDir, gov.SourceHash)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error planning build:", err)
			os.Exit(1)
		}

		printBuildPlan(plan)

		if buildDryRun {
			fmt.Println("Dry run complete. No files written.")
			return
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
			os.Exit(1)
		}

		fmt.Println("Build complete.")
	},
}

func init() {
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "Preview generated target writes without modifying files")
	buildCmd.Flags().StringVar(&buildRestorePath, "restore", "", "Restore one manifest-owned generated artifact path")
	rootCmd.AddCommand(buildCmd)
}

func printBuildPlan(plan engine.BuildPlan) {
	fmt.Println("Build plan:")
	for _, item := range plan.Items {
		if item.Reason == "" {
			fmt.Printf("- %s: %s\n", item.Status, item.Path)
			continue
		}
		fmt.Printf("- %s: %s (%s)\n", item.Status, item.Path, item.Reason)
	}
}
