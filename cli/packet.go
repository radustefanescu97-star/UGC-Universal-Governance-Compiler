package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

var packetNewTaskID string
var packetNewPath string
var packetNewSource string
var packetNewMasterplan string
var packetNewDryRun bool
var packetVerifyPath string
var packetVerifyApproval string

var packetCmd = &cobra.Command{
	Use:   "packet",
	Short: "Create, hash, and verify approval packets",
}

var packetNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new approval packet template",
	Run: func(cmd *cobra.Command, args []string) {
		if packetNewTaskID == "" || packetNewPath == "" || packetNewSource == "" || packetNewMasterplan == "" {
			fmt.Fprintln(os.Stderr, "packet new requires --task-id, --path, --source, and --masterplan")
			os.Exit(1)
		}

		template := engine.ApprovalPacketTemplate(engine.ApprovalPacketOptions{
			TaskID:         packetNewTaskID,
			Path:           packetNewPath,
			SourcePath:     packetNewSource,
			MasterplanPath: packetNewMasterplan,
		})

		if packetNewDryRun {
			fmt.Print(template)
			return
		}

		if _, err := os.Stat(packetNewPath); err == nil {
			fmt.Fprintf(os.Stderr, "packet new refuses to overwrite existing path: %s\n", packetNewPath)
			os.Exit(1)
		} else if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "packet new failed to inspect path: %v\n", err)
			os.Exit(1)
		}

		if err := os.MkdirAll(filepath.Dir(packetNewPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "packet new failed to create parent directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(packetNewPath, []byte(template), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "packet new failed to write packet: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Approval packet created: %s\n", packetNewPath)
	},
}

var packetHashCmd = &cobra.Command{
	Use:   "hash <PATH>",
	Short: "Print SHA256 for an approval packet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := engine.PacketSHA256(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "packet hash failed:", err)
			os.Exit(1)
		}
		fmt.Println(hash)
	},
}

var packetVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a hash-bound approval sentence",
	Run: func(cmd *cobra.Command, args []string) {
		if packetVerifyPath == "" || packetVerifyApproval == "" {
			fmt.Fprintln(os.Stderr, "packet verify requires --packet and --approval")
			os.Exit(1)
		}

		result := engine.VerifyApprovalPacket(packetVerifyPath, packetVerifyApproval)
		if !result.OK {
			fmt.Fprintln(os.Stderr, "Approval verification failed:")
			for _, reason := range result.Reasons {
				fmt.Fprintf(os.Stderr, "- %s\n", reason)
			}
			os.Exit(1)
		}

		fmt.Println("Approval verification passed.")
		fmt.Printf("Task ID: %s\n", result.TaskID)
		fmt.Printf("Packet Path: %s\n", result.PacketPath)
		fmt.Printf("Packet SHA256: %s\n", result.SHA256)
	},
}

func init() {
	packetNewCmd.Flags().StringVar(&packetNewTaskID, "task-id", "", "Approval packet task id")
	packetNewCmd.Flags().StringVar(&packetNewPath, "path", "", "Approval packet path")
	packetNewCmd.Flags().StringVar(&packetNewSource, "source", "", "Connected source-truth path")
	packetNewCmd.Flags().StringVar(&packetNewMasterplan, "masterplan", "", "Connected masterplan path")
	packetNewCmd.Flags().BoolVar(&packetNewDryRun, "dry-run", false, "Print the packet template without writing it")
	packetVerifyCmd.Flags().StringVar(&packetVerifyPath, "packet", "", "Approval packet path")
	packetVerifyCmd.Flags().StringVar(&packetVerifyApproval, "approval", "", "Approval sentence to verify")

	packetCmd.AddCommand(packetNewCmd)
	packetCmd.AddCommand(packetHashCmd)
	packetCmd.AddCommand(packetVerifyCmd)
	rootCmd.AddCommand(packetCmd)
}
