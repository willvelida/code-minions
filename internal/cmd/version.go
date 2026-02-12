package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var readBuildInfo = debug.ReadBuildInfo

func getVersion() string {
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
