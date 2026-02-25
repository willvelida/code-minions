package installer

import "fmt"

// ActionKind describes what a dry-run operation would do to a file.
type ActionKind int

const (
	ActionCreate    ActionKind = iota // File does not exist — would be created
	ActionModify                      // File exists with different content — would be overwritten
	ActionUnchanged                   // File exists with identical content — no-op
	ActionRemove                      // File exists — would be deleted (uninstall)
	ActionNotFound                    // File does not exist — nothing to delete (uninstall)
)

// String returns a human-readable label for the action kind.
func (k ActionKind) String() string {
	switch k {
	case ActionCreate:
		return "create"
	case ActionModify:
		return "modify"
	case ActionUnchanged:
		return "unchanged"
	case ActionRemove:
		return "remove"
	case ActionNotFound:
		return "not found"
	default:
		return fmt.Sprintf("ActionKind(%d)", int(k))
	}
}

// Prefix returns a single-character symbol for display: +, ~, =, -, ?
func (k ActionKind) Prefix() string {
	switch k {
	case ActionCreate:
		return "+"
	case ActionModify:
		return "~"
	case ActionUnchanged:
		return "="
	case ActionRemove:
		return "-"
	case ActionNotFound:
		return "?"
	default:
		return "?"
	}
}

// FileAction describes a single planned filesystem operation during dry-run.
type FileAction struct {
	Path       string     // Relative path of the file
	Kind       ActionKind // What would happen
	Annotation string     // Optional context, e.g. "(add reference to threat-modelling)"
}
