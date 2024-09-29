package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/iagonc/jorge-cli/cmd/cli/commands"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/config"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/usecase"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"
)

func main() {
    // Initialize the logger
    logger, err := utils.InitializeLogger()
    if err != nil {
        os.Exit(1)
    }
    defer logger.Sync()

    // Load configurations
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Error("Error loading configurations", zap.Error(err))
        os.Exit(1)
    }

    // Initialize the HTTP client
    client := utils.NewHTTPClient(cfg.Timeout)

    // Initialize the resource Usecase
    resourceUsecase := usecase.NewResourceUsecase(client, cfg, logger)

    // Set up the root command
    var rootCmd = &cobra.Command{
        Use:     "cli",
        Short:   "CLI to interact with the API",
        Long:    "A command-line tool to interact with the API for managing resources.",
        Version: cfg.Version,
    }

    // Add commands, passing the resourceUsecase
    rootCmd.AddCommand(commands.NewListCommand(resourceUsecase))
    rootCmd.AddCommand(commands.NewCreateCommand(resourceUsecase))
    rootCmd.AddCommand(commands.NewDeleteCommand(resourceUsecase))
    rootCmd.AddCommand(commands.NewUpdateCommand(resourceUsecase))

    // Handle system signals for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go utils.HandleSignals(cancel, logger)

    // Execute the root command with context
    if err := rootCmd.ExecuteContext(ctx); err != nil {
        logger.Error("Error executing command", zap.Error(err))
        os.Exit(1)
    }
}
