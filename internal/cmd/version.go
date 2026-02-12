package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of code-minions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("code-minions %s\n", Version)
		},
	}
}
