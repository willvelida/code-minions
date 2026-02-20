package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptTeamName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid name",
			input: "platform-engineering\n",
			want:  "platform-engineering",
		},
		{
			name:  "name with spaces trimmed",
			input: "  my-team  \n",
			want:  "my-team",
		},
		{
			name:    "empty input",
			input:   "\n",
			wantErr: true,
			errMsg:  "required",
		},
		{
			name:    "EOF with no input",
			input:   "",
			wantErr: true,
			errMsg:  "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := promptTeamName(strings.NewReader(tt.input), &buf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			// Verify prompt was displayed
			if !strings.Contains(buf.String(), "Team name") {
				t.Error("expected prompt to contain 'Team name'")
			}
		})
	}
}

func TestPromptTeamDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid description",
			input: "Standard setup for the platform team\n",
			want:  "Standard setup for the platform team",
		},
		{
			name:    "empty input",
			input:   "\n",
			wantErr: true,
			errMsg:  "required",
		},
		{
			name:    "EOF with no input",
			input:   "",
			wantErr: true,
			errMsg:  "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := promptTeamDescription(strings.NewReader(tt.input), &buf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptPersonaName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid name",
			input: "senior-dev\n",
			want:  "senior-dev",
		},
		{
			name:  "name with spaces trimmed",
			input: "  backend-dev  \n",
			want:  "backend-dev",
		},
		{
			name:    "empty input",
			input:   "\n",
			wantErr: true,
			errMsg:  "required",
		},
		{
			name:    "EOF with no input",
			input:   "",
			wantErr: true,
			errMsg:  "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := promptPersonaName(strings.NewReader(tt.input), &buf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptAssistant(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "explicit choice",
			input: "claude\n",
			want:  "claude",
		},
		{
			name:  "empty input uses default",
			input: "\n",
			want:  "copilot",
		},
		{
			name:  "EOF uses default",
			input: "",
			want:  "copilot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := promptAssistant(strings.NewReader(tt.input), &buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			// Verify prompt shows default
			if !strings.Contains(buf.String(), "copilot") {
				t.Error("expected prompt to show default 'copilot'")
			}
		})
	}
}

func TestSanitizeTeamName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-project", "my-project"},
		{"my project", "my-project"},
		{"my.project", "my-project"},
		{"My_Project", "My_Project"},
		{"--leading", "t--leading"},
		{"_leading", "t_leading"},
		{"valid123", "valid123"},
		{"", "my-team"},
		{"!!!??", "my-team"},
		{"hello world!", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeTeamName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeTeamName(%q): got %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
