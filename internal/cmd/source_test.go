package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/registry"
)

// withTestConfig sets up a temporary config dir and overrides the
// config path for the duration of the test. We do this by creating
// a config file at a known location and pointing LoadConfig/SaveConfig
// to it via environment variable manipulation and direct path usage.
//
// Since the source commands use registry.LoadConfig() and registry.SaveConfig()
// which use os.UserConfigDir(), we instead test the commands by calling
// them with a config file pre-seeded, and verify the file changes.
// For unit tests of the source command, we test the registry config
// functions directly and the command output.

func TestSourceAddCommand(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Redirect config path to a temp dir so we never touch the real
	// global config. SetConfigRoot is thread-safe and restored by Cleanup.
	configDir := t.TempDir()
	registry.SetConfigRoot(configDir)
	t.Cleanup(func() { registry.SetConfigRoot("") })

	var buf bytes.Buffer
	cmd := newSourceAddCommand()
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"my-team", "https://github.com/org/repo.git"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSourceAddCommandRequiresArgs(t *testing.T) {
	cmd := newSourceAddCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestSourceAddCommandRequiresTwoArgs(t *testing.T) {
	cmd := newSourceAddCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"only-name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}

func TestSourceListCommand(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	registry.SetConfigRoot(t.TempDir())
	t.Cleanup(func() { registry.SetConfigRoot("") })

	var buf bytes.Buffer
	cmd := newSourceListCommand()
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should always contain the embedded source
	if !strings.Contains(output, "embedded") {
		t.Error("expected output to contain 'embedded'")
	}
}

func TestSourceListCommandJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	registry.SetConfigRoot(t.TempDir())
	t.Cleanup(func() { registry.SetConfigRoot("") })

	var buf bytes.Buffer
	cmd := newSourceListCommand()
	// Need to register persistent flags since --json is on root
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("quiet", false, "")
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Sources []struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			BuiltIn bool   `json:"built_in"`
		} `json:"sources"`
		Count int `json:"count"`
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, buf.String())
	}

	if result.Count < 1 {
		t.Error("expected at least 1 source (embedded)")
	}

	foundEmbedded := false
	for _, src := range result.Sources {
		if src.Name == "embedded" && src.BuiltIn {
			foundEmbedded = true
		}
	}
	if !foundEmbedded {
		t.Error("expected embedded source in JSON output")
	}
}

func TestSourceRemoveCommandRequiresArg(t *testing.T) {
	cmd := newSourceRemoveCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestSourceRemoveEmbeddedFails(t *testing.T) {
	cmd := newSourceRemoveCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"embedded"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when trying to remove embedded source")
	}
}

func TestSourceCommandSubcommands(t *testing.T) {
	cmd := newSourceCommand()

	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = true
	}

	expected := []string{"add", "list", "remove"}
	for _, name := range expected {
		if !subcommands[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

// TestSourceRoundTrip tests adding, listing, and removing a source
// using the config functions directly (since commands use os.UserConfigDir).
func TestSourceRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sources.yaml")

	// Start with empty config
	cfg := &registry.Config{}
	if err := registry.SaveConfigTo(configPath, cfg); err != nil {
		t.Fatalf("save empty config: %v", err)
	}

	// Add a source
	cfg, err := registry.LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if err := cfg.AddSource(registry.SourceConfig{
		Name: "my-team",
		Type: "git",
		URL:  "https://github.com/org/repo.git",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := registry.SaveConfigTo(configPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Load and verify
	cfg, err = registry.LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("load after add: %v", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}

	// Remove the source
	if err := cfg.RemoveSource("my-team"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := registry.SaveConfigTo(configPath, cfg); err != nil {
		t.Fatalf("save after remove: %v", err)
	}

	// Verify empty
	cfg, err = registry.LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("load after remove: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("expected 0 sources after remove, got %d", len(cfg.Sources))
	}
}

// TestSourceConfigFileContent verifies the YAML file format.
func TestSourceConfigFileContent(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sources.yaml")

	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "my-team", Type: "git", URL: "https://github.com/org/repo.git"},
		},
	}

	if err := registry.SaveConfigTo(configPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: my-team") {
		t.Errorf("expected YAML to contain source name, got:\n%s", content)
	}
	if !strings.Contains(content, "type: git") {
		t.Errorf("expected YAML to contain source type, got:\n%s", content)
	}
	if !strings.Contains(content, "url: https://github.com/org/repo.git") {
		t.Errorf("expected YAML to contain source URL, got:\n%s", content)
	}
}
