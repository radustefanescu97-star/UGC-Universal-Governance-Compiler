package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/universal-governance/ugc/engine"
)

const versionJSONSchemaVersion = 1

var releaseLatestAPIURL = "https://api.github.com/repos/radustefanescu97-star/UGC-Universal-Governance-Compiler/releases/latest"

var versionNoCheck bool
var versionJSON bool

type versionOutput struct {
	SchemaVersion   int     `json:"schema_version"`
	BinaryVersion   string  `json:"binary_version"`
	CorpusVersion   string  `json:"corpus_version"`
	GoVersion       string  `json:"go_version"`
	Platform        string  `json:"platform"`
	LatestRelease   *string `json:"latest_release"`
	UpdateAvailable *bool   `json:"update_available"`
}

type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print UGC binary and corpus version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := buildLocalVersionInfo()
		var release releaseCheckResult
		if !versionNoCheck {
			release = fetchLatestRelease()
		}

		out := cmd.OutOrStdout()
		if versionJSON {
			return printVersionJSON(out, info, release, versionNoCheck)
		}
		printVersionHuman(out, info, release, versionNoCheck)
		return nil
	},
}

type localVersionInfo struct {
	BinaryVersion string
	CorpusVersion string
	GoVersion     string
	Platform      string
}

type releaseCheckResult struct {
	Checked         bool
	OK              bool
	TagName         string
	HTMLURL         string
	UpdateAvailable bool
	UpToDate        bool
}

func buildLocalVersionInfo() localVersionInfo {
	return localVersionInfo{
		BinaryVersion: binaryVersion,
		CorpusVersion: engine.EmbeddedCorpusVersion,
		GoVersion:     runtime.Version(),
		Platform:      runtime.GOOS + "/" + runtime.GOARCH,
	}
}

func fetchLatestRelease() releaseCheckResult {
	result := releaseCheckResult{Checked: true}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(releaseLatestAPIURL)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	var payload githubReleaseResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return result
	}
	if payload.TagName == "" {
		return result
	}

	result.OK = true
	result.TagName = payload.TagName
	result.HTMLURL = payload.HTMLURL

	cmp := compareVersions(binaryVersion, payload.TagName)
	if cmp < 0 {
		result.UpdateAvailable = true
	} else {
		result.UpToDate = true
	}
	return result
}

func printVersionHuman(out io.Writer, info localVersionInfo, release releaseCheckResult, noCheck bool) {
	fmt.Fprintf(out, "Binary version: %s\n", info.BinaryVersion)
	fmt.Fprintf(out, "Corpus version: %s\n", info.CorpusVersion)
	fmt.Fprintf(out, "Go: %s %s\n", info.GoVersion, info.Platform)

	if noCheck {
		return
	}
	if !release.OK {
		fmt.Fprintln(out, "Could not check for updates.")
		return
	}
	if release.UpdateAvailable {
		fmt.Fprintf(out, "A newer release is available: %s\n", release.TagName)
		if release.HTMLURL != "" {
			fmt.Fprintf(out, "Release URL: %s\n", release.HTMLURL)
		}
		fmt.Fprintln(out, "Download the new binary, then run `ugc update` in each governed repository to refresh the governance corpus.")
		return
	}
	fmt.Fprintln(out, "Binary is up to date.")
}

func printVersionJSON(w io.Writer, info localVersionInfo, release releaseCheckResult, noCheck bool) error {
	payload := versionOutput{
		SchemaVersion: versionJSONSchemaVersion,
		BinaryVersion: info.BinaryVersion,
		CorpusVersion: info.CorpusVersion,
		GoVersion:     info.GoVersion,
		Platform:      info.Platform,
	}

	if !noCheck && release.OK {
		tag := release.TagName
		payload.LatestRelease = &tag
		available := release.UpdateAvailable
		payload.UpdateAvailable = &available
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func compareVersions(local, remote string) int {
	localParts := versionParts(local)
	remoteParts := versionParts(remote)
	maxLen := len(localParts)
	if len(remoteParts) > maxLen {
		maxLen = len(remoteParts)
	}
	for i := 0; i < maxLen; i++ {
		lv := 0
		rv := 0
		if i < len(localParts) {
			lv = localParts[i]
		}
		if i < len(remoteParts) {
			rv = remoteParts[i]
		}
		if lv < rv {
			return -1
		}
		if lv > rv {
			return 1
		}
	}
	return 0
}

func versionParts(raw string) []int {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(strings.TrimPrefix(raw, "v"), "V")
	if raw == "" || raw == "dev" {
		return []int{0}
	}

	segments := strings.Split(raw, ".")
	parts := make([]int, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			parts = append(parts, 0)
			continue
		}
		digits := segment
		if idx := strings.IndexFunc(segment, func(r rune) bool {
			return r < '0' || r > '9'
		}); idx >= 0 {
			digits = segment[:idx]
		}
		if digits == "" {
			parts = append(parts, 0)
			continue
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			parts = append(parts, 0)
			continue
		}
		parts = append(parts, n)
	}
	return parts
}

func init() {
	versionCmd.Flags().BoolVar(&versionNoCheck, "no-check", false, "Skip checking GitHub for a newer release")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Print machine-readable version information")
	rootCmd.AddCommand(versionCmd)
}
