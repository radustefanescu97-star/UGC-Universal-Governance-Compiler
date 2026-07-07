package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

var dryRun bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local Standard Corpus",
	Long:  `Synchronize local .universal-governance/ directories with the compiler version while preserving local edits through drift detection.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := engine.UpdateCorpus(dryRun)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Update failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate update without modifying files")
	rootCmd.AddCommand(updateCmd)
}
