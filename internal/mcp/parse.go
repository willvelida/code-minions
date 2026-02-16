package mcp

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
)

// mcpConfigFile is the canonical MCP config filename within a package.
const mcpConfigFile = "mcp.yaml"

// LoadConfig reads and parses the mcp.yaml from a package directory inside
// the given filesystem. Returns (nil, nil) when the package has no mcp.yaml
// — this is the normal case for packages that don't declare MCP servers.
func LoadConfig(content fs.FS, packageDir string) (*Config, error) {
	filePath := path.Join(packageDir, mcpConfigFile)

	data, err := fs.ReadFile(content, filePath)
	if err != nil {
		// fs.ErrNotExist means the package simply doesn't have MCP config —
		// that's fine, not an error. We use errors.Is so that wrapped
		// errors (e.g. *fs.PathError from fstest.MapFS) are also caught.
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	cfg, err := ParseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	return cfg, nil
}
