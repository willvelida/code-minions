package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	-X github.com/willvelida/code-minions/internal/cmd.Version=v1.2.3
var Version string

var readBuildInfo = debug.ReadBuildInfo

func getVersion() string {
	// Prefer the version injected by ldflags (GoReleaser builds).
	if Version != "" {
		return Version
	}

	// Fall back to debug.BuildInfo (go install builds).
	info, ok := readBuildInfo()
	if !ok {
		return "dev"
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of code-minions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("code-minions %s\n", getVersion())
		},
	}
}
