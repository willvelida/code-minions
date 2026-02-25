package cmd

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/installer"
)

// dryRunJSONAction is the JSON representation of a single planned
// filesystem operation during dry-run.
type dryRunJSONAction struct {
	Path       string `json:"path"`
	Action     string `json:"action"` // "create", "modify", "skipped", "unchanged", "remove", "not found"
	Annotation string `json:"annotation,omitempty"`
}

// actionsToJSON converts FileAction slices to their JSON representation.
func actionsToJSON(actions []installer.FileAction) []dryRunJSONAction {
	if len(actions) == 0 {
		return []dryRunJSONAction{}
	}
	out := make([]dryRunJSONAction, len(actions))
	for i, a := range actions {
		out[i] = dryRunJSONAction{
			Path:       a.Path,
			Action:     a.Kind.String(),
			Annotation: a.Annotation,
		}
	}
	return out
}

// dryRunActionGroup pairs a section header with its list of actions.
type dryRunActionGroup struct {
	header string
	kind   installer.ActionKind
	color  *color.Color
}

// printDryRunInstallActions prints grouped dry-run output for install/update.
func printDryRunInstallActions(w io.Writer, actions []installer.FileAction) {
	groups := []dryRunActionGroup{
		{"Would create:", installer.ActionCreate, color.New(color.FgGreen)},
		{"Would modify:", installer.ActionModify, color.New(color.FgYellow)},
		{"Would skip (use --force to overwrite):", installer.ActionSkipped, color.New(color.FgYellow)},
		{"Would not change:", installer.ActionUnchanged, color.New(color.Faint)},
	}
	printGroupedActions(w, actions, groups)

	// Summary
	create, modify, skipped, unchanged := 0, 0, 0, 0
	for _, a := range actions {
		switch a.Kind {
		case installer.ActionCreate:
			create++
		case installer.ActionModify:
			modify++
		case installer.ActionSkipped:
			skipped++
		case installer.ActionUnchanged:
			unchanged++
		}
	}
	bold := color.New(color.Bold)
	_, _ = fmt.Fprintln(w)
	_, _ = bold.Fprintf(w, "Summary: %d to create, %d to modify, %d skipped, %d unchanged\n",
		create, modify, skipped, unchanged)
}

// printDryRunUninstallActions prints grouped dry-run output for uninstall.
func printDryRunUninstallActions(w io.Writer, actions []installer.FileAction) {
	groups := []dryRunActionGroup{
		{"Would remove:", installer.ActionRemove, color.New(color.FgRed)},
		{"Would not find:", installer.ActionNotFound, color.New(color.Faint)},
	}
	printGroupedActions(w, actions, groups)

	// Summary
	remove, notFound := 0, 0
	for _, a := range actions {
		switch a.Kind {
		case installer.ActionRemove:
			remove++
		case installer.ActionNotFound:
			notFound++
		}
	}
	bold := color.New(color.Bold)
	_, _ = fmt.Fprintln(w)
	_, _ = bold.Fprintf(w, "Summary: %d to remove, %d not found\n",
		remove, notFound)
}

// printDryRunUpdateActions prints grouped dry-run output for update.
func printDryRunUpdateActions(w io.Writer, actions []installer.FileAction) {
	groups := []dryRunActionGroup{
		{"Would update:", installer.ActionModify, color.New(color.FgYellow)},
		{"Would create:", installer.ActionCreate, color.New(color.FgGreen)},
		{"Would skip (use --force to overwrite):", installer.ActionSkipped, color.New(color.FgYellow)},
		{"Already up to date:", installer.ActionUnchanged, color.New(color.Faint)},
	}
	printGroupedActions(w, actions, groups)

	// Summary
	update, unchanged := 0, 0
	for _, a := range actions {
		switch a.Kind {
		case installer.ActionModify, installer.ActionCreate:
			update++
		case installer.ActionUnchanged:
			unchanged++
		}
	}
	bold := color.New(color.Bold)
	_, _ = fmt.Fprintln(w)
	_, _ = bold.Fprintf(w, "Summary: %d to update, %d unchanged\n",
		update, unchanged)
}

// printGroupedActions prints actions organised by kind with section headers.
func printGroupedActions(w io.Writer, actions []installer.FileAction, groups []dryRunActionGroup) {
	for _, g := range groups {
		var matching []installer.FileAction
		for _, a := range actions {
			if a.Kind == g.kind {
				matching = append(matching, a)
			}
		}
		if len(matching) == 0 {
			continue
		}

		_, _ = fmt.Fprintln(w, g.header)
		for _, a := range matching {
			annotation := ""
			if a.Annotation != "" {
				annotation = " (" + a.Annotation + ")"
			}
			_, _ = g.color.Fprintf(w, "  %s %s%s\n", a.Kind.Prefix(), a.Path, annotation)
		}
		_, _ = fmt.Fprintln(w)
	}
}

// printDryRunBanner prints the standardised dry-run banner.
func printDryRunBanner(w io.Writer) {
	_, _ = color.New(color.FgYellow, color.Bold).Fprintln(w, "Dry run — no changes will be made")
	_, _ = fmt.Fprintln(w)
}

// mergeActions combines Actions slices from multiple Results.
func mergeActions(dst *[]installer.FileAction, src []installer.FileAction) {
	*dst = append(*dst, src...)
}
