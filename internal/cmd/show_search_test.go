package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/model"
)

func testContentFSWithManifest() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte(`name: git-workflow
version: 0.1.0
description: Git workflow automation
author: testauthor
license: MIT
tags:
  - git
  - workflow
  - automation
compatibility:
  - copilot
  - claude
contents:
  agents:
    - agents/git-workflow.agent.md
  skills:
    - skills/git-workflow/SKILL.md
  actions:
    - skills/git-workflow/actions/commit.md
  standards:
    - skills/git-workflow/standards/branch.md
`),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("# Git Skill"),
		},
		"packages/git-workflow/skills/git-workflow/actions/commit.md": &fstest.MapFile{
			Data: []byte("# Commit"),
		},
		"packages/git-workflow/skills/git-workflow/standards/branch.md": &fstest.MapFile{
			Data: []byte("# Branch"),
		},
		"packages/developer-mentor/package.yaml": &fstest.MapFile{
			Data: []byte(`name: developer-mentor
version: 0.1.0
description: Mentoring
author: testauthor
tags:
  - mentoring
  - learning
`),
		},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{
			Data: []byte("# Mentor Agent"),
		},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: developer-mentor\ndescription: 'Mentoring'\n---\n# Mentor"),
		},
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte(`name: threat-modelling
version: 0.1.0
description: Threat modelling for your codebase
author: testauthor
tags:
  - security
  - architecture
  - threat
compatibility:
  - copilot
  - claude
contents:
  agents:
    - agents/threat-modelling.agent.md
  skills:
    - skills/threat-modelling/SKILL.md
`),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Agent"),
		},
		"packages/threat-modelling/skills/threat-modelling/SKILL.md": &fstest.MapFile{
			Data: []byte("# Threat Skill"),
		},
	}
}

// ---------- show command tests ----------

func TestShowCommandDisplaysPackageDetails(t *testing.T) {
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

	expected := []string{
		"git-workflow",
		"0.1.0",
		"Git workflow automation",
		"testauthor",
		"MIT",
		"copilot, claude",
		"Agents (1)",
		"Skills (1)",
		"Actions (1)",
		"Standards (1)",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q, got:\n%s", want, output)
		}
	}
}

func TestShowCommandJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"show", "git-workflow", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if result["name"] != "git-workflow" {
		t.Errorf("expected name=git-workflow, got %v", result["name"])
	}
	if result["version"] != "0.1.0" {
		t.Errorf("expected version=0.1.0, got %v", result["version"])
	}
}

func TestShowCommandNotFound(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	cmd := newShowCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

// ---------- search command tests ----------

func TestSearchCommandFindsPackage(t *testing.T) {
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
		t.Errorf("search for 'git' should find git-workflow, got:\n%s", output)
	}
	if !strings.Contains(output, "1 found") {
		t.Errorf("should show result count, got:\n%s", output)
	}
}

func TestSearchCommandNoResults(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"zzzznotfound"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No results found") {
		t.Errorf("should show 'No results found', got:\n%s", output)
	}
}

func TestSearchCommandJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"search", "workflow", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Query   string `json:"query"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if result.Query != "workflow" {
		t.Errorf("expected query=workflow, got %v", result.Query)
	}
	if result.Count != 1 {
		t.Errorf("expected 1 result, got %d", result.Count)
	}
}

// ---------- list --detail tests ----------

func TestListDetailShowsContents(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "--detail"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "→") {
		t.Errorf("--detail should show content summary with →, got:\n%s", output)
	}
}

func TestListJSONNotTruncated(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Create a package with a very long description
	longDesc := strings.Repeat("A", 150)
	fs := fstest.MapFS{
		"packages/long-desc/package.yaml": &fstest.MapFile{
			Data: []byte("name: long-desc\nversion: 1.0.0\ndescription: '" + longDesc + "'\n"),
		},
		"packages/long-desc/agents/long-desc.agent.md": &fstest.MapFile{
			Data: []byte("# Agent"),
		},
	}

	var buf bytes.Buffer
	root := newJSONTestRootCmd(fs)
	root.SetOut(&buf)
	root.SetArgs([]string{"list", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Packages []struct {
			Description string `json:"description"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(result.Packages) == 0 {
		t.Fatal("expected at least one package")
	}
	if len(result.Packages[0].Description) != 150 {
		t.Errorf("JSON description should not be truncated, got length %d", len(result.Packages[0].Description))
	}
	if strings.HasSuffix(result.Packages[0].Description, "...") {
		t.Error("JSON description should not end with '...'")
	}
}

// ---------- search --tag tests ----------

func TestSearchTagFlag(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"--tag", "security"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "threat-modelling") {
		t.Errorf("--tag security should find threat-modelling, got:\n%s", output)
	}
	if strings.Contains(output, "git-workflow") {
		t.Errorf("--tag security should NOT find git-workflow, got:\n%s", output)
	}
}

func TestSearchTagFlagWithQuery(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"threat", "--tag", "security"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "threat-modelling") {
		t.Errorf("search 'threat' --tag security should find threat-modelling, got:\n%s", output)
	}
}

func TestSearchTagFlagCaseInsensitive(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"--tag", "Security"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "threat-modelling") {
		t.Errorf("--tag Security (uppercase) should find threat-modelling, got:\n%s", output)
	}
}

func TestSearchTagFlagNoMatch(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"--tag", "nonexistent-tag"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No results found") {
		t.Errorf("--tag nonexistent-tag should show no results, got:\n%s", output)
	}
}

func TestSearchOutputShowsTags(t *testing.T) {
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
	if !strings.Contains(output, "[git, workflow, automation]") {
		t.Errorf("output should show tags in square brackets, got:\n%s", output)
	}
}

func TestSearchMatchesTags(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// "security" doesn't appear in threat-modelling's name but is a tag
	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"security"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "threat-modelling") {
		t.Errorf("search for 'security' should find threat-modelling via tag match, got:\n%s", output)
	}
}

func TestSearchNoQueryNoTagErrors(t *testing.T) {
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no query and no --tag")
	}
	if !strings.Contains(err.Error(), "provide a search query") {
		t.Errorf("error should mention 'provide a search query', got: %v", err)
	}
}

// ---------- search --info tests ----------

func TestSearchInfoFlag(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"--info", "git-workflow"})
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expected := []string{
		"git-workflow",
		"0.1.0",
		"Git workflow automation",
		"testauthor",
		"MIT",
		"git, workflow, automation",
		"Source:",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("--info output should contain %q, got:\n%s", want, output)
		}
	}
}

func TestSearchInfoFlagJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"search", "--info", "git-workflow", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if result["name"] != "git-workflow" {
		t.Errorf("expected name=git-workflow, got %v", result["name"])
	}
	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected tags to be an array, got %T", result["tags"])
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestSearchInfoFlagNotFound(t *testing.T) {
	cmd := newSearchCommand(testContentFSWithManifest())
	cmd.SetArgs([]string{"--info", "nonexistent"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

// ---------- search JSON with tags ----------

func TestSearchJSONIncludesTags(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Override loadManifest to return an empty manifest during tests.
	origLoad := loadManifest
	loadManifest = func(string) (*model.InstallManifest, error) {
		return &model.InstallManifest{Packages: []model.InstalledPackage{}}, nil
	}
	t.Cleanup(func() { loadManifest = origLoad })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"search", "git", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Results []struct {
			Name      string   `json:"name"`
			Tags      []string `json:"tags"`
			Installed bool     `json:"installed"`
		} `json:"results"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(result.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if len(result.Results[0].Tags) != 3 {
		t.Errorf("expected 3 tags on git-workflow, got %d", len(result.Results[0].Tags))
	}
	if result.Results[0].Installed {
		t.Error("expected installed=false in test environment")
	}
}

// ---------- search --tag with JSON ----------

func TestSearchTagFlagJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Override loadManifest for test isolation
	origLoad := loadManifest
	loadManifest = func(string) (*model.InstallManifest, error) {
		return &model.InstallManifest{Packages: []model.InstalledPackage{}}, nil
	}
	t.Cleanup(func() { loadManifest = origLoad })

	var buf bytes.Buffer
	root := newJSONTestRootCmd(testContentFSWithManifest())
	root.SetOut(&buf)
	root.SetArgs([]string{"search", "--tag", "security", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Tag     string `json:"tag"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if result.Tag != "security" {
		t.Errorf("expected tag=security, got %v", result.Tag)
	}
	if result.Count != 1 {
		t.Errorf("expected 1 result, got %d", result.Count)
	}
	if len(result.Results) > 0 && result.Results[0].Name != "threat-modelling" {
		t.Errorf("expected threat-modelling, got %v", result.Results[0].Name)
	}
}
