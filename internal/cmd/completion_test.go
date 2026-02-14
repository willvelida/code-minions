package cmd

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"
)

// newTestRootCmd builds a minimal root command with the completion subcommand
// registered, matching the structure in NewRootCommand.
func newTestRootCmd() *cobra.Command {
	content := fstest.MapFS{}
	return NewRootCommand(content)
}

func TestCompletionBash(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "bash"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "code-minions") {
		t.Errorf("expected bash output to reference code-minions, got:\n%s", output)
	}
}

func TestCompletionZsh(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "zsh"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#compdef") {
		t.Errorf("expected zsh output to contain #compdef, got:\n%s", output)
	}
}

func TestCompletionFish(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "fish"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "complete") {
		t.Errorf("expected fish output to contain 'complete', got:\n%s", output)
	}
}

func TestCompletionPowershell(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "powershell"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("expected PowerShell output to contain Register-ArgumentCompleter, got:\n%s", output)
	}
}

func TestCompletionNoArgsShowsHelp(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "bash") || !strings.Contains(output, "powershell") {
		t.Errorf("expected help output to list available shells, got:\n%s", output)
	}
}
