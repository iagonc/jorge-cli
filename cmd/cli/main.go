package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/iagonc/jorge-cli/cmd/cli/commands"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/config"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"
)

func main() {
    // Carrega as configurações
    cfg, err := config.LoadConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erro ao carregar configurações: %v\n", err)
        os.Exit(1)
    }

    // Inicializa o logger
    logger, err := zap.NewProduction()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erro ao inicializar o logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Sync()

    // Inicializa o cliente HTTP
    client := utils.NewHTTPClient(cfg.Timeout)

    // Inicializa o serviço de recursos
    resourceService := services.NewResourceService(client, cfg, logger)

    // Configura o comando root
    var rootCmd = &cobra.Command{
        Use:     "cli",
        Short:   "CLI to interact with the API",
        Long:    "A command line tool to interact with the API for managing resources.",
        Version: cfg.Version,
    }

    // Adiciona os comandos, passando o resourceService
    rootCmd.AddCommand(commands.NewListCommand(resourceService))
    rootCmd.AddCommand(commands.NewCreateCommand(resourceService))
    rootCmd.AddCommand(commands.NewDeleteCommand(resourceService))
    rootCmd.AddCommand(commands.NewUpdateCommand(resourceService))

    // Gerencia sinais do sistema para cancelamento
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        <-c
        cancel()
    }()

    // Executa o comando com o contexto
    if err := rootCmd.ExecuteContext(ctx); err != nil {
        logger.Sugar().Errorf("Erro ao executar comando: %v", err)
        os.Exit(1)
    }
}
