package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSourceName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"my-team", false},
		{"a", false},
		{"abc-123", false},
		{"a-b-c-d", false},
		// Invalid names
		{"", true},
		{"-starts-with-hyphen", true},
		{"HAS_UPPER", true},
		{"has spaces", true},
		{"has.dots", true},
		{"embedded", true}, // reserved
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourceName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourceName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSourceType(t *testing.T) {
	tests := []struct {
		typ     string
		wantErr bool
	}{
		{"git", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			err := ValidateSourceType(tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourceType(%q) error = %v, wantErr %v", tt.typ, err, tt.wantErr)
			}
		})
	}
}

func TestConfigAddSource(t *testing.T) {
	cfg := &Config{}

	err := cfg.AddSource(SourceConfig{
		Name: "my-team",
		Type: "git",
		URL:  "https://github.com/org/repo.git",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "my-team" {
		t.Errorf("name: got %q", cfg.Sources[0].Name)
	}
}

func TestConfigAddSourceDuplicate(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-team", Type: "git", URL: "https://example.com/repo.git"},
		},
	}

	err := cfg.AddSource(SourceConfig{
		Name: "my-team",
		Type: "git",
		URL:  "https://example.com/other.git",
	})
	if err == nil {
		t.Fatal("expected error for duplicate source name")
	}
}

func TestConfigAddSourceInvalidName(t *testing.T) {
	cfg := &Config{}

	err := cfg.AddSource(SourceConfig{
		Name: "HAS_UPPER",
		Type: "git",
		URL:  "https://example.com/repo.git",
	})
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestConfigAddSourceInvalidType(t *testing.T) {
	cfg := &Config{}

	err := cfg.AddSource(SourceConfig{
		Name: "my-source",
		Type: "ftp",
		URL:  "ftp://example.com/repo",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestConfigAddSourceEmptyURL(t *testing.T) {
	cfg := &Config{}

	err := cfg.AddSource(SourceConfig{
		Name: "my-source",
		Type: "git",
		URL:  "",
	})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestConfigRemoveSource(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "first", Type: "git", URL: "https://example.com/a.git"},
			{Name: "second", Type: "git", URL: "https://example.com/b.git"},
		},
	}

	if err := cfg.RemoveSource("first"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source after removal, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "second" {
		t.Errorf("remaining source: got %q, want %q", cfg.Sources[0].Name, "second")
	}
}

func TestConfigRemoveSourceNotFound(t *testing.T) {
	cfg := &Config{}

	err := cfg.RemoveSource("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent source")
	}
}

func TestConfigFindSource(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-team", Type: "git", URL: "https://example.com/repo.git"},
		},
	}

	found := cfg.FindSource("my-team")
	if found == nil {
		t.Fatal("expected to find source")
	}
	if found.URL != "https://example.com/repo.git" {
		t.Errorf("URL: got %q", found.URL)
	}

	notFound := cfg.FindSource("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent source")
	}
}

func TestLoadConfigFromNonexistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("expected empty config, got %d sources", len(cfg.Sources))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code-minions", "sources.yaml")

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-team", Type: "git", URL: "https://github.com/org/repo.git"},
			{Name: "community", Type: "git", URL: "https://github.com/community/skills.git"},
		},
	}

	if err := SaveConfigTo(path, cfg); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}

	// Load it back
	loaded, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if len(loaded.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(loaded.Sources))
	}
	if loaded.Sources[0].Name != "my-team" {
		t.Errorf("first source name: got %q", loaded.Sources[0].Name)
	}
	if loaded.Sources[1].Name != "community" {
		t.Errorf("second source name: got %q", loaded.Sources[1].Name)
	}
	if loaded.Sources[0].URL != "https://github.com/org/repo.git" {
		t.Errorf("first source URL: got %q", loaded.Sources[0].URL)
	}
}

func TestLoadConfigFromInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(path, []byte("{{{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfigFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should end with the expected suffix
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}

	base := filepath.Base(path)
	if base != "sources.yaml" {
		t.Errorf("expected filename sources.yaml, got %q", base)
	}

	dir := filepath.Base(filepath.Dir(path))
	if dir != "code-minions" {
		t.Errorf("expected parent dir code-minions, got %q", dir)
	}
}

func TestConfigAddSourceReservedName(t *testing.T) {
	cfg := &Config{}

	err := cfg.AddSource(SourceConfig{
		Name: "embedded",
		Type: "git",
		URL:  "https://example.com/repo.git",
	})
	if err == nil {
		t.Fatal("expected error for reserved name 'embedded'")
	}
}

func TestConfigAddMultipleSources(t *testing.T) {
	cfg := &Config{}

	sources := []SourceConfig{
		{Name: "first", Type: "git", URL: "https://example.com/a.git"},
		{Name: "second", Type: "git", URL: "https://example.com/b.git"},
		{Name: "third", Type: "git", URL: "https://example.com/c.git"},
	}

	for _, src := range sources {
		if err := cfg.AddSource(src); err != nil {
			t.Fatalf("adding %q: %v", src.Name, err)
		}
	}

	if len(cfg.Sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(cfg.Sources))
	}

	// Verify ordering is preserved
	for i, src := range sources {
		if cfg.Sources[i].Name != src.Name {
			t.Errorf("source %d: got %q, want %q", i, cfg.Sources[i].Name, src.Name)
		}
	}
}

func TestConfigRemoveMiddleSource(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "first", Type: "git", URL: "https://example.com/a.git"},
			{Name: "second", Type: "git", URL: "https://example.com/b.git"},
			{Name: "third", Type: "git", URL: "https://example.com/c.git"},
		},
	}

	if err := cfg.RemoveSource("second"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "first" || cfg.Sources[1].Name != "third" {
		t.Errorf("remaining sources: %q, %q", cfg.Sources[0].Name, cfg.Sources[1].Name)
	}
}

// TestLoadConfig_EmptyConfigWithTempRoot verifies that LoadConfig
// returns an empty config when a valid config root is set but no
// config file exists yet.
func TestLoadConfig_EmptyConfigWithTempRoot(t *testing.T) {
	root := t.TempDir()
	SetConfigRoot(root)
	t.Cleanup(func() { SetConfigRoot("") })

	// LoadConfig should succeed with empty config (no file yet)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("expected 0 sources, got %d", len(cfg.Sources))
	}
}

// TestLoadConfig_InvalidYAMLPropagatesError verifies that LoadConfig
// surfaces YAML parse errors rather than masking them.
func TestLoadConfig_InvalidYAMLPropagatesError(t *testing.T) {
	root := t.TempDir()
	SetConfigRoot(root)
	t.Cleanup(func() { SetConfigRoot("") })

	// Write invalid YAML to the config path
	configDir := filepath.Join(root, "code-minions")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "sources.yaml"), []byte(": invalid: yaml: [{{"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// TestSetConfigRoot verifies that SetConfigRoot redirects ConfigPath.
func TestSetConfigRoot(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigRoot(tmpDir)
	t.Cleanup(func() { SetConfigRoot("") })

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if !strings.HasPrefix(path, tmpDir) {
		t.Errorf("expected path under %q, got %q", tmpDir, path)
	}
}
