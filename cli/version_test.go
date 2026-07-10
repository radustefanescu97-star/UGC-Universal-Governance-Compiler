package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		local  string
		remote string
		want   int
	}{
		{"1.0.7", "v1.0.8", -1},
		{"1.0.8", "v1.0.8", 0},
		{"1.0.9", "v1.0.8", 1},
		{"dev", "v1.0.8", -1},
		{"1.0.8-test", "v1.0.8", 0},
	}
	for _, tc := range cases {
		if got := compareVersions(tc.local, tc.remote); got != tc.want {
			t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.local, tc.remote, got, tc.want)
		}
	}
}

func TestVersionJSONNoCheckGoldenShape(t *testing.T) {
	oldVersion := binaryVersion
	binaryVersion = "dev"
	defer func() { binaryVersion = oldVersion }()

	buf := &bytes.Buffer{}
	oldOut := versionCmd.OutOrStdout()
	versionCmd.SetOut(buf)
	defer versionCmd.SetOut(oldOut)

	versionNoCheck = true
	versionJSON = true
	defer func() {
		versionNoCheck = false
		versionJSON = false
	}()

	if err := versionCmd.RunE(versionCmd, nil); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var got versionOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}

	if got.SchemaVersion != versionJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", got.SchemaVersion, versionJSONSchemaVersion)
	}
	if got.BinaryVersion != "dev" {
		t.Fatalf("binary_version = %q, want dev", got.BinaryVersion)
	}
	if got.CorpusVersion != "1.0.0" {
		t.Fatalf("corpus_version = %q, want 1.0.0", got.CorpusVersion)
	}
	if got.LatestRelease != nil {
		t.Fatalf("latest_release = %v, want null", got.LatestRelease)
	}
	if got.UpdateAvailable != nil {
		t.Fatalf("update_available = %v, want null", got.UpdateAvailable)
	}
	if got.GoVersion == "" || got.Platform == "" {
		t.Fatalf("expected go_version and platform to be populated: %+v", got)
	}
}

func TestFetchLatestReleaseNewer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(githubReleaseResponse{
			TagName: "v9.9.9",
			HTMLURL: "https://example.com/release",
		})
	}))
	defer server.Close()

	oldURL := releaseLatestAPIURL
	oldVersion := binaryVersion
	releaseLatestAPIURL = server.URL
	binaryVersion = "1.0.0"
	defer func() {
		releaseLatestAPIURL = oldURL
		binaryVersion = oldVersion
	}()

	result := fetchLatestRelease()
	if !result.OK {
		t.Fatal("expected successful release check")
	}
	if result.TagName != "v9.9.9" {
		t.Fatalf("tag_name = %q, want v9.9.9", result.TagName)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update_available true for newer remote tag")
	}
}

func TestFetchLatestReleaseSame(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(githubReleaseResponse{
			TagName: "v1.0.8",
			HTMLURL: "https://example.com/release",
		})
	}))
	defer server.Close()

	oldURL := releaseLatestAPIURL
	oldVersion := binaryVersion
	releaseLatestAPIURL = server.URL
	binaryVersion = "1.0.8"
	defer func() {
		releaseLatestAPIURL = oldURL
		binaryVersion = oldVersion
	}()

	result := fetchLatestRelease()
	if !result.OK {
		t.Fatal("expected successful release check")
	}
	if result.UpdateAvailable {
		t.Fatal("expected update_available false for same version")
	}
	if !result.UpToDate {
		t.Fatal("expected up-to-date result")
	}
}

func TestFetchLatestReleaseUnreachable(t *testing.T) {
	oldURL := releaseLatestAPIURL
	releaseLatestAPIURL = "http://127.0.0.1:1"
	defer func() { releaseLatestAPIURL = oldURL }()

	result := fetchLatestRelease()
	if result.OK {
		t.Fatal("expected unreachable release check to fail gracefully")
	}
}

func TestVersionNoCheckSkipsHTTP(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldURL := releaseLatestAPIURL
	releaseLatestAPIURL = server.URL
	defer func() { releaseLatestAPIURL = oldURL }()

	buf := &bytes.Buffer{}
	oldOut := versionCmd.OutOrStdout()
	versionCmd.SetOut(buf)
	defer versionCmd.SetOut(oldOut)

	versionNoCheck = true
	versionJSON = false
	defer func() {
		versionNoCheck = false
		versionJSON = false
	}()

	if err := versionCmd.RunE(versionCmd, nil); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if called {
		t.Fatal("expected --no-check to skip release HTTP request")
	}
	if !strings.Contains(buf.String(), "Binary version:") {
		t.Fatalf("expected human version output, got: %q", buf.String())
	}
}

func TestPrintVersionHumanUnreachable(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	printVersionHuman(os.Stdout, buildLocalVersionInfo(), releaseCheckResult{Checked: true}, false)
	w.Close()
	os.Stdout = old
	out := <-outCh

	if !strings.Contains(out, "Could not check for updates.") {
		t.Fatalf("expected graceful offline note, got: %q", out)
	}
}
