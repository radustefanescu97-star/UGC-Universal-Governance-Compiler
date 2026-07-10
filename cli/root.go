package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var binaryVersion = "dev"

var rootCmd = &cobra.Command{
	Use:   "ugc",
	Short: "Universal Governance Compiler",
	Long:  `UGC is a local CLI tool that establishes a Single Source of Truth for AI governance.`,
}

// SetVersion wires the build-time binary version into the CLI (including cobra --version).
func SetVersion(version string) {
	if version == "" {
		version = "dev"
	}
	binaryVersion = version
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
