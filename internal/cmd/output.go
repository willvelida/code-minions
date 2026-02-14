package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// OutputMode represents the mutually exclusive output modes for the CLI.
type OutputMode int

const (
	OutputNormal  OutputMode = iota // Default human-readable output
	OutputVerbose                   // Extra debug details
	OutputQuiet                     // Suppress all stdout; errors to stderr only
	OutputJSON                      // Machine-readable JSON
)

// getOutputMode reads --verbose, --quiet, and --json from the command's
// persistent flags and returns the active mode. Mutual exclusivity is
// enforced by Cobra (MarkFlagsMutuallyExclusive), so at most one flag
// is true.
func getOutputMode(cmd *cobra.Command) OutputMode {
	if v, _ := cmd.Flags().GetBool("json"); v {
		return OutputJSON
	}
	if v, _ := cmd.Flags().GetBool("verbose"); v {
		return OutputVerbose
	}
	if v, _ := cmd.Flags().GetBool("quiet"); v {
		return OutputQuiet
	}
	return OutputNormal
}

// verbosePrintf prints a dim/faint line to cmd.OutOrStdout() only when
// mode is OutputVerbose. It is a no-op for all other modes.
func verbosePrintf(cmd *cobra.Command, mode OutputMode, format string, args ...any) {
	if mode != OutputVerbose {
		return
	}
	dim := color.New(color.Faint)
	_, _ = dim.Fprintf(cmd.OutOrStdout(), format, args...)
}

// quietWarning prints a one-line warning to stderr when --quiet is used
// with an informational command (list, version) where suppressing output
// would be a no-op.
func quietWarning(cmd *cobra.Command, mode OutputMode) {
	if mode != OutputQuiet {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "warning: --quiet has no effect on '%s' (the command exists to display information)\n", cmd.Name())
}
