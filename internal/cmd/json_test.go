package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"
)

// newJSONTestRootCmd builds a root command wired to the given filesystem.
// This ensures the persistent --json flag is registered on the root.
func newJSONTestRootCmd(content fstest.MapFS) *cobra.Command {
	return NewRootCommand(content)
}

func TestVersionJSON(t *testing.T) {
	// Pin the version so the output is deterministic.
	originalVersion := Version
	originalBuildInfo := readBuildInfo
	Version = "v1.2.3"
	readBuildInfo = func() (*debug.BuildInfo, bool) { return nil, false }
	t.Cleanup(func() {
		Version = originalVersion
		readBuildInfo = originalBuildInfo
	})

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(fstest.MapFS{})
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}
	if result.Version != "v1.2.3" {
		t.Errorf("version: got %q, want %q", result.Version, "v1.2.3")
	}
}

func TestListJSON(t *testing.T) {
	content := testContentFSWithDescriptions()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Packages []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"packages"`
		Assistants []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"assistants"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if len(result.Packages) == 0 {
		t.Fatal("expected at least one package in JSON output")
	}

	// Verify package names are present
	names := make(map[string]bool)
	for _, p := range result.Packages {
		names[p.Name] = true
	}
	for _, want := range []string{"git-workflow", "developer-mentor"} {
		if !names[want] {
			t.Errorf("expected package %q in output", want)
		}
	}

	if len(result.Assistants) == 0 {
		t.Fatal("expected at least one assistant in JSON output")
	}

	// Verify all assistants are present
	aNames := make(map[string]bool)
	for _, a := range result.Assistants {
		aNames[a.Name] = true
	}
	for _, want := range []string{"copilot", "claude", "opencode"} {
		if !aNames[want] {
			t.Errorf("expected assistant %q in output", want)
		}
	}
}

func TestInstallJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Copied  []string `json:"copied"`
		Skipped []string `json:"skipped"`
		Errors  []string `json:"errors"`
		Summary struct {
			Copied  int `json:"copied"`
			Skipped int `json:"skipped"`
			Errors  int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if result.Summary.Copied == 0 {
		t.Error("expected at least one copied file")
	}
	if result.Summary.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", result.Summary.Errors)
	}

	// Verify files were actually written to disk
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected file at %s", agentFile)
	}
}

func TestInstallJSONSkipsExisting(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// First install
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Second install with --json should show skipped files
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--json"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Skipped []string `json:"skipped"`
		Summary struct {
			Skipped int `json:"skipped"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if result.Summary.Skipped == 0 {
		t.Error("expected skipped files on second install")
	}
}

func TestUpdateJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first so there's something to update
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Update with --json
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"update", "--package", "git-workflow", "--target", target, "--json"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Updated []string `json:"updated"`
		Errors  []string `json:"errors"`
		Summary struct {
			Updated int `json:"updated"`
			Errors  int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if result.Summary.Updated == 0 {
		t.Error("expected at least one updated file")
	}
	if result.Summary.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", result.Summary.Errors)
	}
}

func TestUninstallJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first so there's something to uninstall
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall with --json
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"uninstall", "--package", "git-workflow", "--target", target, "--json"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Removed     []string `json:"removed"`
		NotFound    []string `json:"not_found"`
		Errors      []string `json:"errors"`
		DirsCleaned []string `json:"dirs_cleaned"`
		Summary     struct {
			Removed  int `json:"removed"`
			NotFound int `json:"not_found"`
			Errors   int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if result.Summary.Removed == 0 {
		t.Error("expected at least one removed file")
	}
	if result.Summary.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", result.Summary.Errors)
	}
}
