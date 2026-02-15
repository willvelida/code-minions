package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------
// Mutual exclusivity tests (Task 14)
// ---------------------------------------------------------------

func TestVerboseAndQuietAreMutuallyExclusive(t *testing.T) {
	cmd := newJSONTestRootCmd(fstest.MapFS{})
	cmd.SetArgs([]string{"version", "--verbose", "--quiet"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --verbose and --quiet are both set")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("error should mention flag group exclusivity, got: %v", err)
	}
}

func TestVerboseAndJSONAreMutuallyExclusive(t *testing.T) {
	cmd := newJSONTestRootCmd(fstest.MapFS{})
	cmd.SetArgs([]string{"version", "--verbose", "--json"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --verbose and --json are both set")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("error should mention flag group exclusivity, got: %v", err)
	}
}

func TestQuietAndJSONAreMutuallyExclusive(t *testing.T) {
	cmd := newJSONTestRootCmd(fstest.MapFS{})
	cmd.SetArgs([]string{"version", "--quiet", "--json"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --quiet and --json are both set")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("error should mention flag group exclusivity, got: %v", err)
	}
}

// ---------------------------------------------------------------
// --quiet tests (Task 13)
// ---------------------------------------------------------------

func TestInstallQuiet(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stdout should be empty in quiet mode
	if buf.Len() != 0 {
		t.Errorf("expected empty stdout in quiet mode, got: %q", buf.String())
	}

	// Files should still be installed
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected file at %s", agentFile)
	}
}

func TestInstallQuietStillInstalls(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both agent and skill files exist
	for _, rel := range []string{
		filepath.Join("agents", "git-workflow.agent.md"),
		filepath.Join("skills", "git-workflow", "SKILL.md"),
	} {
		path := filepath.Join(target, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file at %s", path)
		}
	}
}

func TestUninstallQuiet(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall with --quiet
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"uninstall", "--package", "git-workflow", "--target", target, "--quiet"})

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty stdout in quiet mode, got: %q", buf.String())
	}
}

func TestUpdateQuiet(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Update with --quiet
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"update", "--package", "git-workflow", "--target", target, "--quiet"})

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty stdout in quiet mode, got: %q", buf.String())
	}
}

func TestUpdateQuietNoPackages(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"update", "--target", target, "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce no output (quiet suppresses "no packages found" message)
	if buf.Len() != 0 {
		t.Errorf("expected empty stdout in quiet mode, got: %q", buf.String())
	}
}

func TestVersionQuietWarns(t *testing.T) {
	originalVersion := Version
	originalBuildInfo := readBuildInfo
	Version = "v1.2.3"
	readBuildInfo = func() (*debug.BuildInfo, bool) { return nil, false }
	t.Cleanup(func() {
		Version = originalVersion
		readBuildInfo = originalBuildInfo
	})

	var stdout, stderr bytes.Buffer
	cmd := newJSONTestRootCmd(fstest.MapFS{})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version", "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stdout should still have the version (quiet is no-op for version)
	if !strings.Contains(stdout.String(), "v1.2.3") {
		t.Errorf("expected version in stdout, got: %q", stdout.String())
	}

	// Stderr should have the warning
	if !strings.Contains(stderr.String(), "--quiet has no effect") {
		t.Errorf("expected quiet warning on stderr, got: %q", stderr.String())
	}
}

func TestListQuietWarns(t *testing.T) {
	content := testContentFS()

	var stdout, stderr bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"list", "--quiet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stderr should have the warning
	if !strings.Contains(stderr.String(), "--quiet has no effect") {
		t.Errorf("expected quiet warning on stderr, got: %q", stderr.String())
	}
}

// ---------------------------------------------------------------
// --verbose tests (Task 12)
// ---------------------------------------------------------------

func TestInstallVerbose(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--verbose"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verbose should include the package list
	if !strings.Contains(output, "packages:") {
		t.Errorf("expected verbose package list in output, got:\n%s", output)
	}

	// Files should still be installed
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected file at %s", agentFile)
	}
}

func TestInstallVerboseShowsSkipHint(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// First install
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Second install with --verbose should show skip hints
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"install", "--package", "git-workflow", "--target", target, "--verbose"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--force") {
		t.Errorf("expected verbose skip hint mentioning --force, got:\n%s", output)
	}
}

func TestUpdateVerbose(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Update with --verbose
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"update", "--package", "git-workflow", "--target", target, "--verbose"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "packages") {
		t.Errorf("expected verbose package info in output, got:\n%s", output)
	}
}

func TestUpdateVerboseAutoDetect(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	cmd1 := newJSONTestRootCmd(content)
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetArgs([]string{"install", "--package", "git-workflow", "--target", target})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Update without --package (auto-detect) with --verbose
	var buf bytes.Buffer
	cmd2 := newJSONTestRootCmd(content)
	cmd2.SetOut(&buf)
	cmd2.SetArgs([]string{"update", "--target", target, "--verbose"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "auto-detected") {
		t.Errorf("expected 'auto-detected' in verbose output, got:\n%s", output)
	}
}

func TestListVerbose(t *testing.T) {
	content := testContentFS()

	var buf bytes.Buffer
	cmd := newJSONTestRootCmd(content)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list", "--verbose"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "scanned:") {
		t.Errorf("expected verbose scan paths in output, got:\n%s", output)
	}
}

// ---------------------------------------------------------------
// getOutputMode unit tests
// ---------------------------------------------------------------

func TestGetOutputMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want OutputMode
	}{
		{
			name: "default is normal",
			args: []string{"version"},
			want: OutputNormal,
		},
		{
			name: "json flag",
			args: []string{"version", "--json"},
			want: OutputJSON,
		},
		{
			name: "verbose flag",
			args: []string{"version", "--verbose"},
			want: OutputVerbose,
		},
		{
			name: "quiet flag",
			args: []string{"version", "--quiet"},
			want: OutputQuiet,
		},
		{
			name: "verbose short flag",
			args: []string{"version", "-v"},
			want: OutputVerbose,
		},
		{
			name: "quiet short flag",
			args: []string{"version", "-q"},
			want: OutputQuiet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMode OutputMode
			root := NewRootCommand(fstest.MapFS{})
			// Replace version command's RunE to capture the mode
			for _, c := range root.Commands() {
				if c.Name() == "version" {
					c.RunE = func(cmd *cobra.Command, args []string) error {
						capturedMode = getOutputMode(cmd)
						return nil
					}
				}
			}
			root.SetOut(&bytes.Buffer{})
			root.SetArgs(tt.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedMode != tt.want {
				t.Errorf("got mode %d, want %d", capturedMode, tt.want)
			}
		})
	}
}
