package main

import (
	"fmt"
	"os"

	"github.com/iagonc/jorge-cli/cmd/cli/commands"
	"github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "cli",
        Short: "CLI to interact with the API",
        Long:  "A command line tool to interact with the API for managing resources.",
    }

    rootCmd.AddCommand(commands.NewListCommand())
    rootCmd.AddCommand(commands.NewCreateCommand())
    rootCmd.AddCommand(commands.NewDeleteCommand())
	rootCmd.AddCommand(commands.NewUpdateCommand())

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
