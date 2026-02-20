package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// SourceConfig represents a configured external package source.
type SourceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // "git"
	URL  string `yaml:"url"`
}

// Config holds the global source configuration.
type Config struct {
	Sources []SourceConfig `yaml:"sources"`
}

// validSourceName matches lowercase alphanumeric strings with hyphens,
// 1-64 characters long.
var validSourceName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// reservedSourceNames are names that cannot be used for user-defined sources.
var reservedSourceNames = map[string]bool{
	"embedded": true,
}

// validSourceTypes lists the supported source types.
var validSourceTypes = map[string]bool{
	"git": true,
}

// ValidateSourceName checks that a source name is valid.
func ValidateSourceName(name string) error {
	if !validSourceName.MatchString(name) {
		return fmt.Errorf("source name %q is invalid: must be 1-64 lowercase alphanumeric characters or hyphens, starting with a letter or digit", name)
	}
	if reservedSourceNames[name] {
		return fmt.Errorf("source name %q is reserved", name)
	}
	return nil
}

// ValidateSourceType checks that a source type is supported.
func ValidateSourceType(t string) error {
	if !validSourceTypes[t] {
		supported := make([]string, 0, len(validSourceTypes))
		for k := range validSourceTypes {
			supported = append(supported, k)
		}
		return fmt.Errorf("unsupported source type %q (supported: %s)", t, strings.Join(supported, ", "))
	}
	return nil
}

// configRootOverride, when non-empty, replaces the OS user config
// directory for LoadConfig / SaveConfig. Used by tests to isolate
// config operations from the real global config.
var (
	configRootOverride string
	configRootMu       sync.RWMutex
)

// SetConfigRoot overrides the config root directory used by
// ConfigPath. Pass an empty string to restore the default behaviour.
// Intended for tests.
func SetConfigRoot(root string) {
	configRootMu.Lock()
	configRootOverride = root
	configRootMu.Unlock()
}

// ConfigPath returns the platform-appropriate path to sources.yaml.
// Uses os.UserConfigDir() which returns:
//   - Linux: $XDG_CONFIG_HOME or ~/.config
//   - macOS: ~/Library/Application Support
//   - Windows: %AppData%
func ConfigPath() (string, error) {
	configRootMu.RLock()
	root := configRootOverride
	configRootMu.RUnlock()

	if root != "" {
		return filepath.Join(root, "code-minions", "sources.yaml"), nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine config directory: %w", err)
	}
	return filepath.Join(configDir, "code-minions", "sources.yaml"), nil
}

// LoadConfig reads sources.yaml from the platform-appropriate config
// directory. If the file does not exist, an empty Config is returned
// (not an error — it means no external sources are configured).
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadConfigFrom(path)
}

// LoadConfigFrom reads a Config from the given file path.
// If the file does not exist, an empty Config is returned.
func LoadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	return &cfg, nil
}

// SaveConfig writes the config to the platform-appropriate path.
func SaveConfig(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return SaveConfigTo(path, cfg)
}

// SaveConfigTo writes a Config to the given file path, creating
// parent directories as needed.
func SaveConfigTo(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialise config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config %s: %w", path, err)
	}

	return nil
}

// AddSource adds a source to the config. Returns an error if a source
// with the same name already exists, or if the name/type is invalid.
func (c *Config) AddSource(src SourceConfig) error {
	if err := ValidateSourceName(src.Name); err != nil {
		return err
	}
	if err := ValidateSourceType(src.Type); err != nil {
		return err
	}
	if src.URL == "" {
		return fmt.Errorf("source URL is required")
	}

	for _, existing := range c.Sources {
		if existing.Name == src.Name {
			return fmt.Errorf("source %q already exists. Remove it first with 'code-minions source remove %s'", src.Name, src.Name)
		}
	}

	c.Sources = append(c.Sources, src)
	return nil
}

// RemoveSource removes a source by name. Returns an error if the
// source is not found.
func (c *Config) RemoveSource(name string) error {
	for i, src := range c.Sources {
		if src.Name == name {
			c.Sources = append(c.Sources[:i], c.Sources[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("source %q not found", name)
}

// FindSource returns a source config by name, or nil if not found.
func (c *Config) FindSource(name string) *SourceConfig {
	for i := range c.Sources {
		if c.Sources[i].Name == name {
			return &c.Sources[i]
		}
	}
	return nil
}
