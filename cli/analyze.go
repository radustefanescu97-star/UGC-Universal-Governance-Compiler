package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine/ingestion"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Scan for framework indicators",
	Long:  `Read-only reporting tool to scan for framework indicators (Node, Go, Python).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ingestion.AnalyzeProject()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error analyzing:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
