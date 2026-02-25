package installer

import "testing"

func TestActionKindString(t *testing.T) {
	tests := []struct {
		kind ActionKind
		want string
	}{
		{ActionCreate, "create"},
		{ActionModify, "modify"},
		{ActionUnchanged, "unchanged"},
		{ActionRemove, "remove"},
		{ActionNotFound, "not found"},
		{ActionKind(99), "ActionKind(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("ActionKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestActionKindPrefix(t *testing.T) {
	tests := []struct {
		kind ActionKind
		want string
	}{
		{ActionCreate, "+"},
		{ActionModify, "~"},
		{ActionUnchanged, "="},
		{ActionRemove, "-"},
		{ActionNotFound, "?"},
		{ActionKind(99), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := tt.kind.Prefix(); got != tt.want {
				t.Errorf("ActionKind.Prefix() = %q, want %q", got, tt.want)
			}
		})
	}
}
