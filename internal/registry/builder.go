package registry

import (
	"fmt"
	"io"
	"io/fs"
)

// BuildRegistry constructs a Registry from the global source config
// and the embedded filesystem. Sources are ordered:
// 1. Named sources from config (in config order)
// 2. Embedded source (always last — fallback)
//
// If the config file does not exist, a registry with only the
// embedded source is returned (no error).
//
// Failing configured sources are silently skipped — use
// BuildRegistryWithWarnings to surface warnings to the user.
func BuildRegistry(embeddedContent fs.FS) (*Registry, error) {
	return BuildRegistryFromConfig(embeddedContent, nil)
}

// BuildRegistryWithWarnings is like BuildRegistry but writes warning
// messages to w when a configured source fails to initialise. This is
// the preferred entry point for CLI commands so that the user sees
// which sources were skipped on stderr.
func BuildRegistryWithWarnings(embeddedContent fs.FS, w io.Writer) (*Registry, error) {
	return BuildRegistryFromConfigWithWarnings(embeddedContent, nil, w)
}

// BuildRegistryFromConfig constructs a Registry using a specific
// config. If cfg is nil, LoadConfig() is called to read the global
// config. This variant is useful for testing.
//
// Failing configured sources are skipped (warn-and-skip). Use
// BuildRegistryFromConfigWithWarnings to surface those warnings.
func BuildRegistryFromConfig(embeddedContent fs.FS, cfg *Config) (*Registry, error) {
	return BuildRegistryFromConfigWithWarnings(embeddedContent, cfg, nil)
}

// BuildRegistryFromConfigWithWarnings constructs a Registry from a
// config, writing warnings for any sources that fail to initialise.
// If w is nil, warnings are discarded.
func BuildRegistryFromConfigWithWarnings(embeddedContent fs.FS, cfg *Config, w io.Writer) (*Registry, error) {
	if cfg == nil {
		var err error
		cfg, err = LoadConfig()
		if err != nil {
			// Config load failure is not fatal — fall back to embedded only,
			// but warn the user so they can investigate.
			if w != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to load source config: %v; using embedded sources only\n", err)
			}
			return NewRegistry(NewEmbeddedSource(embeddedContent)), nil
		}
	}

	var sources []Source
	for _, sc := range cfg.Sources {
		switch sc.Type {
		case "git":
			src, err := NewGitSource(sc)
			if err != nil {
				// Warn-and-skip: auto-loaded sources should not block the command
				if w != nil {
					_, _ = fmt.Fprintf(w, "Warning: skipping source %q: %v\n", sc.Name, err)
				}
				continue
			}
			// Eagerly validate the source so failures are caught now
			// (warn-and-skip) rather than surfacing as hard errors
			// later inside Registry.ListPackages / Search / etc.
			if err := src.Ensure(); err != nil {
				if w != nil {
					_, _ = fmt.Fprintf(w, "Warning: skipping source %q: %v\n", sc.Name, err)
				}
				continue
			}
			sources = append(sources, src)
		default:
			if w != nil {
				_, _ = fmt.Fprintf(w, "Warning: skipping source %q: unsupported type %q\n", sc.Name, sc.Type)
			}
			continue
		}
	}

	// Embedded source is always last in priority
	sources = append(sources, NewEmbeddedSource(embeddedContent))

	return NewRegistry(sources...), nil
}

// BuildRegistryWithFrom constructs a Registry scoped to a single
// source resolved from the --from flag. When --from is specified,
// the registry contains only the resolved source (no configured
// sources, no embedded).
//
// The from parameter is resolved as follows:
//  1. If it matches a configured source name → use that source
//  2. If it looks like a Git URL → create an ad-hoc GitSource
//  3. Otherwise → error
func BuildRegistryWithFrom(embeddedContent fs.FS, from string) (*Registry, error) {
	return BuildRegistryWithFromAndConfig(embeddedContent, from, nil)
}

// BuildRegistryWithFromAndConfig is the testable variant of
// BuildRegistryWithFrom that accepts an explicit config.
func BuildRegistryWithFromAndConfig(embeddedContent fs.FS, from string, cfg *Config) (*Registry, error) {
	if cfg == nil {
		var err error
		cfg, err = LoadConfig()
		if err != nil {
			cfg = &Config{} // Continue with empty config
		}
	}

	// Try to resolve --from as a named source first
	if sc := cfg.FindSource(from); sc != nil {
		if err := ValidateSourceType(sc.Type); err != nil {
			return nil, fmt.Errorf("source %q: %w", from, err)
		}
		// Build a registry with only this source (--from restricts scope)
		src, err := NewGitSource(*sc)
		if err != nil {
			return nil, fmt.Errorf("failed to create source %q: %w", from, err)
		}
		return NewRegistry(src), nil
	}

	// Try as a Git URL
	if IsGitURL(from) {
		src, err := NewGitSourceFromURL(from)
		if err != nil {
			return nil, fmt.Errorf("failed to create source from URL %q: %w", from, err)
		}
		return NewRegistry(src), nil
	}

	// Neither a named source nor a URL
	return nil, fmt.Errorf("source %q not found. Run 'code-minions source list' to see configured sources, or provide a Git URL", from)
}

// ResolveFrom resolves a --from value into a Source, using the given
// config for named source lookup. Returns the source and nil error
// on success.
func ResolveFrom(from string, cfg *Config) (Source, error) {
	if cfg == nil {
		var err error
		cfg, err = LoadConfig()
		if err != nil {
			cfg = &Config{}
		}
	}

	// Named source
	if sc := cfg.FindSource(from); sc != nil {
		if err := ValidateSourceType(sc.Type); err != nil {
			return nil, fmt.Errorf("source %q: %w", from, err)
		}
		return NewGitSource(*sc)
	}

	// Git URL
	if IsGitURL(from) {
		return NewGitSourceFromURL(from)
	}

	return nil, fmt.Errorf("source %q not found. Run 'code-minions source list' to see configured sources, or provide a Git URL", from)
}
