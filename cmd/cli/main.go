package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/iagonc/jorge-cli/cmd/cli/commands"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/config"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/network"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
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
	resourceUsecase := resource.NewResourceUsecase(client, cfg, logger)
	networkUsecase := network.NewNetworkDebugUsecase(logger)

	// Set up the root command
	var rootCmd = &cobra.Command{
		Use:     "jorge-cli",
		Short:   "Jorge CLI - A friendly network diagnostic and resource management tool",
		Long:    "A command-line tool to perform network diagnostics and manage resources via API.",
		Version: cfg.Version,
	}

	// Add commands, passing the usecases
	rootCmd.AddCommand(commands.NewListCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewCreateCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewDeleteCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewUpdateCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewNetworkDebugCommand(networkUsecase))

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
