package registry

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/willvelida/code-minions/internal/model"
)

// GitSource wraps a cloned Git repository as a Source.
// The repository must contain packages under a packages/ directory,
// following the same layout as code-minions' own packages/ directory.
type GitSource struct {
	mu       sync.Mutex
	name     string
	url      string
	cacheDir string
	embedded *EmbeddedSource // delegates all FS operations
	headSHA  string          // commit SHA after Ensure(); used by lockfile
}

// NewGitSource creates a Git-based source from a config entry.
// The repository is cloned (or pulled) lazily when Ensure() is called.
func NewGitSource(cfg SourceConfig) (*GitSource, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("git source %q: URL is required", cfg.Name)
	}

	cacheDir, err := gitCacheDir(cfg.Name)
	if err != nil {
		return nil, err
	}

	return &GitSource{
		name:     cfg.Name,
		url:      cfg.URL,
		cacheDir: cacheDir,
	}, nil
}

// NewGitSourceFromURL creates an ad-hoc Git source from a URL.
// Used when --from is a URL rather than a named source.
// The cache directory is derived from a hash of the URL.
//
// Bare domain/path forms like "github.com/org/repo" are normalised
// to "https://github.com/org/repo" so that git clone succeeds.
func NewGitSourceFromURL(url string) (*GitSource, error) {
	if url == "" {
		return nil, fmt.Errorf("git source URL is required")
	}

	// Normalise bare domain/path forms (e.g. "github.com/org/repo")
	// to https:// URLs so git clone works reliably.
	if !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "git@") &&
		!strings.HasPrefix(url, "ssh://") {
		url = "https://" + url
	}

	// Generate a deterministic name from the URL hash
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))[:12]
	name := "url-" + hash

	cacheDir, err := gitCacheDir(name)
	if err != nil {
		return nil, err
	}

	return &GitSource{
		name:     name,
		url:      url,
		cacheDir: cacheDir,
	}, nil
}

// NewGitSourceWithDir creates a GitSource pointing to an existing
// local directory. This is primarily for testing, allowing tests to
// set up a local Git repo and wrap it directly.
func NewGitSourceWithDir(name, url, dir string) *GitSource {
	return &GitSource{
		name:     name,
		url:      url,
		cacheDir: dir,
	}
}

// Name returns the source name.
func (s *GitSource) Name() string { return s.name }

// Type returns "git".
func (s *GitSource) Type() string { return "git" }

// URL returns the source URL.
func (s *GitSource) URL() string { return s.url }

// HeadSHA returns the commit SHA of the cloned/pulled repository.
// Returns "" if Ensure() has not been called yet.
func (s *GitSource) HeadSHA() string { return s.headSHA }

// Ensure clones the repo if not cached, or pulls if already cached.
// After this call, s.embedded is ready for use.
// Ensure is safe for concurrent use.
func (s *GitSource) Ensure() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.embedded != nil {
		return nil // Already initialised
	}

	if err := checkGitAvailable(); err != nil {
		return err
	}

	if dirExists(s.cacheDir) {
		// Pull latest changes
		if err := gitPull(s.cacheDir); err != nil {
			// If pull fails (e.g. force push), try a fresh clone
			if removeErr := os.RemoveAll(s.cacheDir); removeErr != nil {
				return fmt.Errorf("pull failed and cleanup failed: pull: %w, cleanup: %v", err, removeErr)
			}
			if err := gitClone(s.url, s.cacheDir); err != nil {
				return fmt.Errorf("failed to clone source %q from %s: %w", s.name, s.url, err)
			}
		}
	} else {
		if err := gitClone(s.url, s.cacheDir); err != nil {
			return fmt.Errorf("failed to clone source %q from %s: %w", s.name, s.url, err)
		}
	}

	// Capture the commit SHA for lockfile pinning.
	if sha, err := gitRevParseHEAD(s.cacheDir); err == nil {
		s.headSHA = sha
	}

	// Verify the cloned repo has a packages/ directory
	packagesDir := filepath.Join(s.cacheDir, "packages")
	if !dirExists(packagesDir) {
		return fmt.Errorf("source %q does not contain a packages/ directory", s.name)
	}

	s.embedded = NewEmbeddedSource(os.DirFS(s.cacheDir))
	return nil
}

// ListPackages returns all packages from the cloned Git repo.
func (s *GitSource) ListPackages() ([]model.Package, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	pkgs, err := s.embedded.ListPackages()
	if err != nil {
		return nil, err
	}
	// Note: source stamping (setting Package.Source) is handled at the
	// registry level by Registry.ListPackages, not here.
	return pkgs, nil
}

// GetPackage returns metadata for a specific package.
func (s *GitSource) GetPackage(name string) (*model.Package, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.GetPackage(name)
}

// DownloadPackage returns an fs.FS rooted at the package directory.
func (s *GitSource) DownloadPackage(name string, version string) (fs.FS, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.DownloadPackage(name, version)
}

// ListPersonas returns all personas from the cloned repo.
func (s *GitSource) ListPersonas() ([]model.Persona, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.ListPersonas()
}

// GetPersona returns a specific persona.
func (s *GitSource) GetPersona(name string) (*model.Persona, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.GetPersona(name)
}

// ListTeams returns all teams from the cloned repo.
func (s *GitSource) ListTeams() ([]model.Team, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.ListTeams()
}

// GetTeam returns a specific team.
func (s *GitSource) GetTeam(name string) (*model.Team, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return s.embedded.GetTeam(name)
}

// Search performs a text search across the cloned repo's packages.
func (s *GitSource) Search(query string) ([]model.SearchResult, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	results, err := s.embedded.Search(query)
	if err != nil {
		return nil, err
	}
	// Override the source name so results reflect this git source
	for i := range results {
		results[i].Source = s.name
	}
	return results, nil
}

// --- Git helpers ---

// checkGitAvailable verifies that git is on the PATH.
func checkGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is required but not found in PATH. Install Git from https://git-scm.com")
	}
	return nil
}

// gitClone performs a shallow clone of a Git repo.
func gitClone(url, dest string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(output))
		if strings.Contains(outStr, "Authentication") || strings.Contains(outStr, "authentication") ||
			strings.Contains(outStr, "Permission denied") || strings.Contains(outStr, "could not read") {
			return fmt.Errorf("authentication failed for %s. Ensure your Git credentials are configured: %s", url, outStr)
		}
		return fmt.Errorf("git clone failed: %s", outStr)
	}
	return nil
}

// gitRevParseHEAD returns the full commit SHA of HEAD in the given repo.
func gitRevParseHEAD(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gitPull performs a fast-forward pull in an existing repo.
func gitPull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// cacheRootOverride, when non-empty, replaces the OS user cache
// directory as the root for git source caches. This is intended for
// testing so that tests don't pollute (or collide with) the real
// user cache.
var (
	cacheRootOverride string
	cacheRootMu       sync.RWMutex
)

// SetCacheRoot overrides the cache root directory used by GitSource.
// Pass an empty string to restore the default (os.UserCacheDir).
// This is exported for use in integration tests.
func SetCacheRoot(root string) {
	cacheRootMu.Lock()
	cacheRootOverride = root
	cacheRootMu.Unlock()
}

// gitCacheDir returns the cache directory for a named source.
func gitCacheDir(name string) (string, error) {
	cacheRootMu.RLock()
	root := cacheRootOverride
	cacheRootMu.RUnlock()
	if root == "" {
		var err error
		root, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine cache directory: %w", err)
		}
		root = filepath.Join(root, "code-minions", "sources")
	}
	return filepath.Join(root, name), nil
}

// dirExists returns true if the path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsGitURL returns true if the string looks like a Git URL.
// It matches explicit URL schemes (https://, http://, git@, ssh://)
// and bare domain/path forms like "github.com/org/repo" (must contain
// both a "." and a "/").
func IsGitURL(s string) bool {
	if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
		return true
	}
	if strings.HasPrefix(s, "git@") {
		return true
	}
	if strings.HasPrefix(s, "ssh://") {
		return true
	}
	// github.com/user/repo style
	if strings.Contains(s, "/") && strings.Contains(s, ".") {
		return true
	}
	return false
}

// ShortNameFromURL derives a human-readable short name from a Git URL.
// For example:
//
//	"https://github.com/org/repo.git" → "org/repo"
//	"git@github.com:org/repo.git"     → "org/repo"
//
// If the URL cannot be parsed, a hash-based fallback is returned.
func ShortNameFromURL(url string) string {
	// Strip common prefixes and suffixes
	s := url
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "ssh://")
	s = strings.TrimPrefix(s, "git@")
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")

	// For git@host:org/repo style, replace ":" with "/"
	s = strings.Replace(s, ":", "/", 1)

	// Try to extract org/repo (last two path segments)
	parts := strings.Split(s, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	// Fallback: hash-based name
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))[:12]
	return "url-" + hash
}
