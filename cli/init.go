package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine/ingestion"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize UGC in the current directory",
	Long:  `Bootstraps a project with the .universal-governance/ standard templates and infers constraints.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ingestion.InitTemplates()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error initializing:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
