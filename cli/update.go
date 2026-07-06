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
	Short: "Actualizează Standard Corpus local",
	Long:  `Sincronizează folderele locale .universal-governance/ cu versiunea din compilator, respectând modificările locale (Drift Detection).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := engine.UpdateCorpus(dryRun)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Eroare la actualizare:", err)
			os.Exit(1)
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate update without modifying files")
	rootCmd.AddCommand(updateCmd)
}
