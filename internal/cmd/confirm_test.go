package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmPromptAcceptsY(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("y\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for 'y' input")
	}
	if !strings.Contains(out.String(), "Continue?") {
		t.Errorf("expected prompt in output, got: %q", out.String())
	}
}

func TestConfirmPromptAcceptsYes(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("yes\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for 'yes' input")
	}
}

func TestConfirmPromptAcceptsUppercaseY(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("Y\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for 'Y' input")
	}
}

func TestConfirmPromptAcceptsYES(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("YES\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for 'YES' input")
	}
}

func TestConfirmPromptDeclinesN(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("n\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for 'n' input")
	}
}

func TestConfirmPromptDeclinesEmpty(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for empty input (default No)")
	}
}

func TestConfirmPromptDeclinesArbitrary(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader("maybe\n"), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for arbitrary input")
	}
}

func TestConfirmPromptDeclinesEOF(t *testing.T) {
	var out bytes.Buffer
	ok, err := confirmPrompt(strings.NewReader(""), &out, "Continue? [y/N]: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for EOF (no input)")
	}
}
