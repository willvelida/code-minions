package main

import (
	"fmt"
	"os"

	codeminions "github.com/willvelida/code-minions"
	"github.com/willvelida/code-minions/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand(codeminions.Content)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
