package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/deps"
	"github.com/spf13/cobra"
)

// NewDepsCommand creates the deps command
func NewDepsCommand(usecase *deps.DepsUsecase) *cobra.Command {
	var configPath string
	var urls []string
	var tcpAddrs []string

	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Testa dependências de um serviço",
		Long: `Verifica a saúde de todas as dependências de um serviço.

Pode usar um arquivo de configuração YAML ou especificar dependências diretamente.`,
		Example: `  # Via arquivo de configuração
  jorge deps --config deps.yaml

  # Via linha de comando
  jorge deps --url https://api.exemplo.com/health --url https://auth.exemplo.com
  jorge deps --tcp db.exemplo.com:5432 --tcp redis.exemplo.com:6379`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			var config *models.DepsConfig
			var err error

			if configPath != "" {
				config, err = usecase.LoadConfig(configPath)
				if err != nil {
					return err
				}
			} else if len(urls) > 0 || len(tcpAddrs) > 0 {
				config = usecase.CreateQuickDeps(urls, tcpAddrs)
			} else {
				return fmt.Errorf("especifique --config ou --url/--tcp")
			}

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🔗 Dependency Check"))
			if config.Name != "" {
				fmt.Printf("   Config: %s\n", config.Name)
			}
			fmt.Printf("   Dependências: %d\n", len(config.Dependencies))
			fmt.Println()

			result, err := usecase.CheckAll(ctx, config)
			if err != nil {
				return err
			}

			displayDepsResult(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Arquivo de configuração YAML")
	cmd.Flags().StringArrayVar(&urls, "url", nil, "URL HTTP para verificar (pode repetir)")
	cmd.Flags().StringArrayVar(&tcpAddrs, "tcp", nil, "Endereço TCP host:porta (pode repetir)")

	return cmd
}

func displayDepsResult(result *models.DepsResult) {
	fmt.Println(strings.Repeat("─", 60))

	for _, dep := range result.Results {
		var icon string
		var statusStyle lipgloss.Style

		if dep.Healthy {
			icon = "✓"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		} else {
			icon = "✗"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}

		requiredTag := ""
		if dep.Required {
			requiredTag = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(" [required]")
		}

		typeTag := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("[%s]", dep.Type))

		fmt.Printf("   %s %s %s%s\n",
			statusStyle.Render(icon),
			dep.Name,
			typeTag,
			requiredTag)

		if dep.Healthy {
			fmt.Printf("      %s (%s)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(dep.Details),
				dep.ResponseTime.Round(time.Millisecond))
		} else {
			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(dep.Error))
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))

	var summaryIcon string
	var summaryStyle lipgloss.Style
	if result.AllHealthy {
		summaryIcon = "✅"
		summaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
		fmt.Printf("   %s %s\n", summaryIcon, summaryStyle.Render("Todas as dependências OK"))
	} else {
		summaryIcon = "❌"
		summaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		fmt.Printf("   %s %s\n", summaryIcon, summaryStyle.Render(fmt.Sprintf("%d/%d dependências falharam", result.UnhealthyDeps, result.TotalDeps)))
	}

	fmt.Printf("   Tempo total: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Println()
}
