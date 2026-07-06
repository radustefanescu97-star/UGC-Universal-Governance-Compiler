package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Verify structural validity and file drift",
	Long:  `Ensures structural validity, verifies file drift, checks target capability coverage, and reports on corpus state.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Auditing governance configurations...")

		result, err := engine.AuditProject(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Audit error:", err)
			os.Exit(1)
		}

		if len(result.SourceErrors) > 0 {
			for _, sourceErr := range result.SourceErrors {
				fmt.Fprintf(os.Stderr, "Source validity failed: %s\n", sourceErr)
			}
		} else {
			fmt.Println("Source validity: ok")
		}
		if result.SourceHash != "" {
			fmt.Printf("Source Hash: %s\n", result.SourceHash)
		}

		for _, drift := range result.Drift {
			fmt.Printf("Drift detected: %s\n", drift.Message)
		}
		for _, path := range result.UnexpectedArtifacts {
			fmt.Printf("Drift detected: Unexpected generated artifact %s\n", path)
		}
		for _, finding := range result.ManifestFindings {
			fmt.Printf("Build manifest failed: %s\n", finding)
		}

		switch result.CorpusState {
		case "ok":
			fmt.Println("Corpus state: ok")
		case "missing", "legacy":
			fmt.Printf("Corpus state: %s\n", result.CorpusStateMessage)
		case "failed":
			fmt.Fprintln(os.Stderr, "Corpus state failed:", result.CorpusStateMessage)
		}

		if result.Failed() {
			fmt.Fprintln(os.Stderr, "Audit failed: Drift detected in generated files.")
			os.Exit(1)
		}

		fmt.Println("Target capability coverage:")
		fmt.Println(result.CapabilitySummary)
		fmt.Println("Audit complete. No drift detected.")
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
}
