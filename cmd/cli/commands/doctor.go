package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/diagnose"
	"github.com/spf13/cobra"
)

// NewDoctorCommand creates the doctor command
func NewDoctorCommand(usecase *diagnose.DoctorUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Verifica a saúde da sua rede local",
		Long: `Verifica a saúde da sua conexão de rede local.

Executa automaticamente:
  - Verificação de interfaces de rede
  - Teste de configuração DNS
  - Teste de conectividade com gateway
  - Teste de acesso à Internet

Útil para diagnosticar problemas de conexão antes de testar serviços externos.`,
		Example: `  # Verificar saúde da rede local
  jorge-cli doctor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fmt.Println()
			fmt.Println(diagTitleStyle.Render("🏥 Doctor - Verificação de Rede Local"))
			fmt.Println()
			fmt.Println("   Verificando sua conexão de rede...")
			fmt.Println()

			result, err := usecase.RunDoctor(ctx)
			if err != nil {
				return fmt.Errorf("erro no diagnóstico: %w", err)
			}

			displayDoctorResult(result)
			return nil
		},
	}

	return cmd
}

func displayDoctorResult(result *models.DoctorResult) {
	// Connectivity
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" CONECTIVIDADE LOCAL")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if result.Connectivity.HasIPv4 || result.Connectivity.HasIPv6 {
		fmt.Printf("   %s Interfaces ativas: %s\n",
			diagPassedStyle.Render("✓"),
			strings.Join(result.Connectivity.Interfaces, ", "))
		if len(result.Connectivity.LocalIPs) > 0 {
			fmt.Printf("   %s IPs locais: %s\n",
				diagPassedStyle.Render("✓"),
				strings.Join(result.Connectivity.LocalIPs, ", "))
		}
		if result.Connectivity.HasIPv4 {
			fmt.Printf("   %s IPv4: Configurado\n", diagPassedStyle.Render("✓"))
		}
		if result.Connectivity.HasIPv6 {
			fmt.Printf("   %s IPv6: Configurado\n", diagPassedStyle.Render("✓"))
		}
	} else {
		fmt.Printf("   %s Sem endereço IP configurado\n", diagFailedStyle.Render("✗"))
	}

	// DNS
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" DNS")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if len(result.DNS.Servers) > 0 {
		fmt.Printf("   Servidores: %s\n", strings.Join(result.DNS.Servers, ", "))
	}

	if result.DNS.CanResolve {
		latencyStyle := diagPassedStyle
		if result.DNS.ResponseTime > 500*time.Millisecond {
			latencyStyle = diagWarningStyle
		}
		fmt.Printf("   %s Resolução: Funcionando (%s)\n",
			diagPassedStyle.Render("✓"),
			latencyStyle.Render(result.DNS.ResponseTime.Round(time.Millisecond).String()))
	} else {
		fmt.Printf("   %s Resolução: Falhou\n", diagFailedStyle.Render("✗"))
	}

	// Gateway
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" GATEWAY")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if result.Gateway.IP != "" {
		fmt.Printf("   Gateway: %s\n", result.Gateway.IP)
		if result.Gateway.Reachable {
			fmt.Printf("   %s Status: Acessível (%s)\n",
				diagPassedStyle.Render("✓"),
				result.Gateway.ResponseTime.Round(time.Millisecond))
		} else {
			fmt.Printf("   %s Status: Inacessível\n", diagFailedStyle.Render("✗"))
		}
	} else {
		fmt.Printf("   %s Gateway não encontrado\n", diagWarningStyle.Render("!"))
	}

	// Internet
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" INTERNET")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if result.Internet.Connected {
		latencyStyle := diagPassedStyle
		if result.Internet.ResponseTime > 1*time.Second {
			latencyStyle = diagWarningStyle
		}
		fmt.Printf("   %s Acesso: Funcionando (%s)\n",
			diagPassedStyle.Render("✓"),
			latencyStyle.Render(result.Internet.ResponseTime.Round(time.Millisecond).String()))
		if result.Internet.PublicIP != "" {
			fmt.Printf("   IP Público: %s\n", result.Internet.PublicIP)
		}
	} else {
		fmt.Printf("   %s Sem acesso à Internet\n", diagFailedStyle.Render("✗"))
	}

	// Overall Result
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" RESULTADO")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	var statusIcon, statusText string
	var statusStyle lipgloss.Style
	switch result.Overall {
	case models.StatusHealthy:
		statusIcon = "✅"
		statusText = "SUA REDE ESTÁ OK!"
		statusStyle = diagPassedStyle
	case models.StatusUnhealthy:
		statusIcon = "⚠️"
		statusText = "ATENÇÃO NECESSÁRIA"
		statusStyle = diagWarningStyle
	case models.StatusError:
		statusIcon = "❌"
		statusText = "PROBLEMAS DETECTADOS"
		statusStyle = diagFailedStyle
	}

	fmt.Printf("   %s %s\n", statusIcon, statusStyle.Render(statusText))
	fmt.Printf("   Tempo de verificação: %v\n", result.Duration.Round(time.Millisecond))

	// Problems
	if len(result.Problems) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" PROBLEMAS")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()

		for _, problem := range result.Problems {
			var icon string
			var style lipgloss.Style
			switch problem.Severity {
			case models.SeverityCritical:
				icon = "🔴"
				style = diagFailedStyle
			case models.SeverityWarning:
				icon = "🟡"
				style = diagWarningStyle
			default:
				icon = "🔵"
				style = diagCategoryStyle
			}

			fmt.Printf("   %s %s\n", icon, style.Render(problem.Title))
			fmt.Printf("      %s\n", problem.Description)
		}
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(" 💡 COMO RESOLVER")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()

		for i, suggestion := range result.Suggestions {
			fmt.Printf("   %d. %s\n", i+1, lipgloss.NewStyle().Bold(true).Render(suggestion.Title))
			fmt.Printf("      %s\n", suggestion.Description)
			if suggestion.Command != "" {
				fmt.Printf("      $ %s\n", diagCommandStyle.Render(suggestion.Command))
			}
		}
	}

	fmt.Println()
}
