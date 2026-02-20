package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
)

// ---------- --from flag tests ----------

// TestListFromFlagExists verifies that the --from flag is registered on list.
func TestListFromFlagExists(t *testing.T) {
	cmd := newListCommand(testContentFS())
	f := cmd.Flags().Lookup("from")
	if f == nil {
		t.Fatal("expected --from flag on list command")
	}
	if f.DefValue != "" {
		t.Errorf("--from default should be empty, got %q", f.DefValue)
	}
}

// TestSearchFromFlagExists verifies that the --from flag is registered on search.
func TestSearchFromFlagExists(t *testing.T) {
	cmd := newSearchCommand(testContentFS())
	f := cmd.Flags().Lookup("from")
	if f == nil {
		t.Fatal("expected --from flag on search command")
	}
}

// TestShowFromFlagExists verifies that the --from flag is registered on show.
func TestShowFromFlagExists(t *testing.T) {
	cmd := newShowCommand(testContentFS())
	f := cmd.Flags().Lookup("from")
	if f == nil {
		t.Fatal("expected --from flag on show command")
	}
}

// TestInstallFromFlagExists verifies that the --from flag is registered on install.
func TestInstallFromFlagExists(t *testing.T) {
	cmd := newInstallCommand(testContentFS())
	f := cmd.Flags().Lookup("from")
	if f == nil {
		t.Fatal("expected --from flag on install command")
	}
}

// TestInstallFromRequiresPackage verifies that --from without --package errors.
func TestInstallFromRequiresPackage(t *testing.T) {
	cmd := newInstallCommand(testContentFS())
	cmd.SetArgs([]string{"--from", "some-source"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --from used without --package")
	}
	if !strings.Contains(err.Error(), "--from requires --package") {
		t.Errorf("error should mention --from requires --package, got: %v", err)
	}
}

// TestListWithoutFromUsesEmbedded verifies that list without --from
// still works with just embedded packages (backward compatible).
func TestListWithoutFromUsesEmbedded(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newListCommand(testContentFS())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "git-workflow") {
		t.Errorf("expected git-workflow in output, got:\n%s", output)
	}
}

// TestSearchWithoutFromUsesEmbedded verifies that search without --from
// still works with just embedded packages.
func TestSearchWithoutFromUsesEmbedded(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"git"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "git-workflow") {
		t.Errorf("expected git-workflow in output, got:\n%s", output)
	}
}

// TestShowWithoutFromUsesEmbedded verifies that show without --from
// still works with embedded packages.
func TestShowWithoutFromUsesEmbedded(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newShowCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"git-workflow"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "git-workflow") {
		t.Errorf("expected git-workflow in output, got:\n%s", output)
	}
}

// TestListFromInvalidSourceErrors verifies that an invalid --from value
// produces a clear error.
func TestListFromInvalidSourceErrors(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	cmd := newListCommand(testContentFS())
	cmd.SetArgs([]string{"--from", "nonexistent-source"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --from value")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention source not found, got: %v", err)
	}
}

// TestSearchFromInvalidSourceErrors verifies that an invalid --from value
// produces a clear error on search.
func TestSearchFromInvalidSourceErrors(t *testing.T) {
	cmd := newSearchCommand(testContentFS())
	cmd.SetArgs([]string{"query", "--from", "nonexistent-source"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --from value")
	}
}

// TestShowFromInvalidSourceErrors verifies that an invalid --from value
// produces a clear error on show.
func TestShowFromInvalidSourceErrors(t *testing.T) {
	cmd := newShowCommand(testContentFS())
	cmd.SetArgs([]string{"some-pkg", "--from", "nonexistent-source"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --from value")
	}
}

// TestListVerboseShowsSourceName verifies that verbose output shows
// the correct source name when --from is not set.
func TestListVerboseShowsSourceName(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFS())
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "--verbose"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "source: embedded") {
		t.Errorf("verbose output should show 'source: embedded', got:\n%s", output)
	}
}

// TestListJSONIncludesFromSource verifies that the list --json output
// still works correctly when --from is not set.
func TestListJSONIncludesFromSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Packages []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(result.Packages) == 0 {
		t.Fatal("expected at least one package in JSON output")
	}

	found := false
	for _, p := range result.Packages {
		if p.Name == "git-workflow" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected git-workflow in JSON package list")
	}
}

// TestInstallFromWithPersonaErrors verifies that --from with --persona
// without --package is allowed (persona install may use it).
func TestInstallFromWithPersonaNoPackage(t *testing.T) {
	// --from + --persona without --package should not trigger the
	// "--from requires --package" error, because persona installs
	// handle source resolution differently.
	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{"--from", "some-source", "--persona", "test-persona", "--for", "copilot"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// We expect an error (persona probably doesn't exist), but NOT
	// the "--from requires --package" error.
	err := cmd.Execute()
	if err != nil && strings.Contains(err.Error(), "--from requires --package") {
		t.Errorf("--from + --persona should not require --package, got: %v", err)
	}
}
