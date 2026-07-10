package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

const packetVerifyJSONSchemaVersion = 1

var packetVerifyJSON bool

type packetVerifyJSONOutput struct {
	SchemaVersion int      `json:"schema_version"`
	OK            bool     `json:"ok"`
	TaskID        string   `json:"task_id,omitempty"`
	PacketPath    string   `json:"packet_path,omitempty"`
	SHA256        string   `json:"sha256,omitempty"`
	Reasons       []string `json:"reasons"`
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if packetVerifyPath == "" || packetVerifyApproval == "" {
			return fmt.Errorf("packet verify requires --packet and --approval")
		}

		result := engine.VerifyApprovalPacket(packetVerifyPath, packetVerifyApproval)
		out := cmd.OutOrStdout()

		if packetVerifyJSON {
			if err := printPacketVerifyJSON(out, result); err != nil {
				return err
			}
			if !result.OK {
				return fmt.Errorf("approval verification failed")
			}
			return nil
		}

		if !result.OK {
			fmt.Fprintln(os.Stderr, "Approval verification failed:")
			for _, reason := range result.Reasons {
				fmt.Fprintf(os.Stderr, "- %s\n", reason)
			}
			os.Exit(1)
		}

		fmt.Fprintln(out, "Approval verification passed.")
		fmt.Fprintf(out, "Task ID: %s\n", result.TaskID)
		fmt.Fprintf(out, "Packet Path: %s\n", result.PacketPath)
		fmt.Fprintf(out, "Packet SHA256: %s\n", result.SHA256)
		return nil
	},
}

func buildPacketVerifyJSON(result engine.ApprovalVerification) packetVerifyJSONOutput {
	reasons := result.Reasons
	if reasons == nil {
		reasons = []string{}
	}
	payload := packetVerifyJSONOutput{
		SchemaVersion: packetVerifyJSONSchemaVersion,
		OK:            result.OK,
		Reasons:       reasons,
	}
	if result.OK {
		payload.TaskID = result.TaskID
		payload.PacketPath = result.PacketPath
		payload.SHA256 = result.SHA256
	}
	return payload
}

func printPacketVerifyJSON(out io.Writer, result engine.ApprovalVerification) error {
	payload := buildPacketVerifyJSON(result)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func init() {
	packetNewCmd.Flags().StringVar(&packetNewTaskID, "task-id", "", "Approval packet task id")
	packetNewCmd.Flags().StringVar(&packetNewPath, "path", "", "Approval packet path")
	packetNewCmd.Flags().StringVar(&packetNewSource, "source", "", "Connected source-truth path")
	packetNewCmd.Flags().StringVar(&packetNewMasterplan, "masterplan", "", "Connected masterplan path")
	packetNewCmd.Flags().BoolVar(&packetNewDryRun, "dry-run", false, "Print the packet template without writing it")
	packetVerifyCmd.Flags().StringVar(&packetVerifyPath, "packet", "", "Approval packet path")
	packetVerifyCmd.Flags().StringVar(&packetVerifyApproval, "approval", "", "Approval sentence to verify")
	packetVerifyCmd.Flags().BoolVar(&packetVerifyJSON, "json", false, "Print machine-readable verification result")

	packetCmd.AddCommand(packetNewCmd)
	packetCmd.AddCommand(packetHashCmd)
	packetCmd.AddCommand(packetVerifyCmd)
	rootCmd.AddCommand(packetCmd)
}
