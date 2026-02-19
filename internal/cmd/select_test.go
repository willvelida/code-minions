package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/willvelida/code-minions/internal/model"
)

func testPackages() []model.Package {
	return []model.Package{
		{Name: "threat-modelling", Description: "STRIDE-based threat modelling"},
		{Name: "git-workflow", Description: "Git workflow conventions"},
		{Name: "developer-mentor", Description: "Developer mentoring skill"},
	}
}

func TestSelectPackagesSpecificNumbers(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("1,3\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 2 {
		t.Fatalf("expected 2 selected, got %d: %v", len(selected), selected)
	}
	if selected[0] != "threat-modelling" {
		t.Errorf("selected[0]: got %q, want %q", selected[0], "threat-modelling")
	}
	if selected[1] != "developer-mentor" {
		t.Errorf("selected[1]: got %q, want %q", selected[1], "developer-mentor")
	}
}

func TestSelectPackagesAll(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("all\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 3 {
		t.Fatalf("expected 3 selected, got %d: %v", len(selected), selected)
	}
	for i, pkg := range packages {
		if selected[i] != pkg.Name {
			t.Errorf("selected[%d]: got %q, want %q", i, selected[i], pkg.Name)
		}
	}
}

func TestSelectPackagesAllCaseInsensitive(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("ALL\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 3 {
		t.Fatalf("expected 3 selected, got %d", len(selected))
	}
}

func TestSelectPackagesEmptyInputReturnsAll(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 3 {
		t.Fatalf("expected 3 selected (default all), got %d: %v", len(selected), selected)
	}
}

func TestSelectPackagesInvalidInput(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("abc\n")
	var stdout bytes.Buffer

	_, err := selectPackages(stdin, &stdout, packages)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
	if !strings.Contains(err.Error(), "invalid selection") {
		t.Errorf("expected 'invalid selection' in error, got: %v", err)
	}
}

func TestSelectPackagesOutOfRange(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("5\n")
	var stdout bytes.Buffer

	_, err := selectPackages(stdin, &stdout, packages)
	if err == nil {
		t.Fatal("expected error for out-of-range input, got nil")
	}
	if !strings.Contains(err.Error(), "must be between 1 and") {
		t.Errorf("expected range error, got: %v", err)
	}
}

func TestSelectPackagesZeroIndex(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("0\n")
	var stdout bytes.Buffer

	_, err := selectPackages(stdin, &stdout, packages)
	if err == nil {
		t.Fatal("expected error for index 0, got nil")
	}
}

func TestSelectPackagesDeduplicates(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("1,1,2\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 2 {
		t.Fatalf("expected 2 (deduplicated), got %d: %v", len(selected), selected)
	}
	if selected[0] != "threat-modelling" || selected[1] != "git-workflow" {
		t.Errorf("unexpected selection: %v", selected)
	}
}

func TestSelectPackagesWithSpaces(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader(" 1 , 2 \n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(selected), selected)
	}
}

func TestSelectPackagesEmptyList(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	selected, err := selectPackages(stdin, &stdout, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 0 {
		t.Errorf("expected empty result for empty package list, got %v", selected)
	}
}

func TestSelectPackagesOutputFormat(t *testing.T) {
	packages := testPackages()
	stdin := strings.NewReader("all\n")
	var stdout bytes.Buffer

	_, err := selectPackages(stdin, &stdout, packages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Available packages:") {
		t.Error("output should contain 'Available packages:' header")
	}
	if !strings.Contains(output, "[1] threat-modelling") {
		t.Error("output should contain numbered package list")
	}
	if !strings.Contains(output, "STRIDE-based threat modelling") {
		t.Error("output should contain package descriptions")
	}
}

// --- selectAssistant tests ---

func TestSelectAssistantDefaultsToDetected(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	result, err := selectAssistant(stdin, &stdout, []string{"claude", "copilot"}, []string{"claude", "codex", "copilot", "cursor", "gemini", "opencode"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "claude" {
		t.Errorf("expected default to first detected 'claude', got %q", result)
	}
}

func TestSelectAssistantDefaultsToCopilotWhenNoneDetected(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	result, err := selectAssistant(stdin, &stdout, nil, []string{"claude", "copilot"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "copilot" {
		t.Errorf("expected default 'copilot', got %q", result)
	}
}

func TestSelectAssistantUserChoice(t *testing.T) {
	stdin := strings.NewReader("gemini\n")
	var stdout bytes.Buffer

	result, err := selectAssistant(stdin, &stdout, []string{"copilot"}, []string{"copilot", "gemini"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "gemini" {
		t.Errorf("expected 'gemini', got %q", result)
	}
}

func TestSelectAssistantShowsDetected(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	_, err := selectAssistant(stdin, &stdout, []string{"copilot"}, []string{"copilot", "claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Detected assistants: copilot") {
		t.Errorf("output should show detected assistants, got:\n%s", output)
	}
}
