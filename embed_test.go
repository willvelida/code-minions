package codeminions

import (
	"io/fs"
	"testing"
)

func TestEmbeddedContentContainsPackages(t *testing.T) {
	entries, err := fs.ReadDir(Content, "packages")
	if err != nil {
		t.Fatalf("failed to read packages directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("packages directory is empty")
	}

	// Verify a known package exists and has the expected structure
	expectedPackages := []string{
		"threat-modelling",
		"developer-mentor",
		"git-workflow",
	}

	for _, pkg := range expectedPackages {
		// Check the package directory exists
		_, err := fs.Stat(Content, "packages/"+pkg)
		if err != nil {
			t.Errorf("expected package %q not found: %v", pkg, err)
			continue
		}

		// Check it has an agents subdirectory
		_, err = fs.ReadDir(Content, "packages/"+pkg+"/agents")
		if err != nil {
			t.Errorf("package %q missing agents/ directory: %v", pkg, err)
		}

		// Check it has a skills subdirectory
		_, err = fs.ReadDir(Content, "packages/"+pkg+"/skills")
		if err != nil {
			t.Errorf("package %q missing skills/ directory: %v", pkg, err)
		}

		// Check it has an instructions subdirectory
		entries, err := fs.ReadDir(Content, "packages/"+pkg+"/instructions")
		if err != nil {
			t.Errorf("package %q missing instructions/ directory: %v", pkg, err)
		} else if len(entries) == 0 {
			t.Errorf("package %q has empty instructions/ directory", pkg)
		}

		// Check it has a prompts subdirectory
		promptEntries, err := fs.ReadDir(Content, "packages/"+pkg+"/prompts")
		if err != nil {
			t.Errorf("package %q missing prompts/ directory: %v", pkg, err)
		} else if len(promptEntries) == 0 {
			t.Errorf("package %q has empty prompts/ directory", pkg)
		}
	}
}
