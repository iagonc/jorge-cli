package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/tunnel"
	"github.com/spf13/cobra"
)

// NewTunnelCommand creates the tunnel command
func NewTunnelCommand(usecase *tunnel.TunnelUsecase) *cobra.Command {
	var proxyURL string
	var proxyType string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "tunnel <target>",
		Short: "Testa conectividade através de proxy/tunnel",
		Long: `Testa se um destino é acessível diretamente e/ou através de um proxy.

Útil para:
  - Verificar se precisa de proxy para acessar recursos
  - Comparar latência direta vs proxy
  - Testar configuração de proxy corporativo
  - Debug de problemas de conectividade`,
		Example: `  # Testar acesso direto
  jorge tunnel api.exemplo.com

  # Testar via proxy HTTP
  jorge tunnel api.interno.com --proxy http://proxy.corp:8080

  # Testar via SOCKS5
  jorge tunnel db.interno.com:5432 --proxy socks5://proxy:1080 --type socks5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
			defer cancel()

			pType := models.ProxyTypeDirect
			if proxyType != "" {
				pType = models.ProxyType(proxyType)
			} else if proxyURL != "" {
				// Auto-detect proxy type
				if strings.HasPrefix(proxyURL, "socks5://") {
					pType = models.ProxyTypeSOCKS5
				} else if strings.HasPrefix(proxyURL, "socks4://") {
					pType = models.ProxyTypeSOCKS4
				} else {
					pType = models.ProxyTypeHTTP
				}
			}

			config := models.TunnelConfig{
				Target:    target,
				ProxyURL:  proxyURL,
				ProxyType: pType,
				Timeout:   timeout,
			}

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🚇 Tunnel/Proxy Test"))
			fmt.Printf("   Target: %s\n", target)
			if proxyURL != "" {
				fmt.Printf("   Proxy: %s (%s)\n", proxyURL, pType)
			}
			fmt.Println()

			result, err := usecase.Test(ctx, config)
			if err != nil {
				return err
			}

			displayTunnelResult(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&proxyURL, "proxy", "", "URL do proxy (http://host:port ou socks5://host:port)")
	cmd.Flags().StringVar(&proxyType, "type", "", "Tipo de proxy (http, https, socks4, socks5)")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Timeout por teste")

	return cmd
}

func displayTunnelResult(result *models.TunnelResult) {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" TESTES DE CONECTIVIDADE")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	for _, test := range result.Tests {
		var icon string
		var statusStyle lipgloss.Style

		if test.Success {
			icon = "✓"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		} else {
			icon = "✗"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}

		fmt.Printf("   %s %s\n",
			statusStyle.Render(icon),
			lipgloss.NewStyle().Bold(true).Render(test.Name))

		if test.Success {
			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(test.Details))
		} else {
			fmt.Printf("      %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(test.Error))
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" RESUMO")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Direct connection status
	directIcon := "✗"
	directStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	if result.Summary.DirectWorking {
		directIcon = "✓"
		directStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	}
	fmt.Printf("   %s Conexão Direta: %s\n",
		directStyle.Render(directIcon),
		directStyle.Render(func() string {
			if result.Summary.DirectWorking {
				return "OK"
			}
			return "FALHOU"
		}()))

	// Proxy status (if tested)
	if result.ProxyUsed != "" {
		proxyIcon := "✗"
		proxyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		if result.Summary.ProxyWorking {
			proxyIcon = "✓"
			proxyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		}
		fmt.Printf("   %s Conexão via Proxy: %s\n",
			proxyStyle.Render(proxyIcon),
			proxyStyle.Render(func() string {
				if result.Summary.ProxyWorking {
					return "OK"
				}
				return "FALHOU"
			}()))

		// Latency comparison
		if result.Summary.DirectWorking && result.Summary.ProxyWorking {
			fmt.Println()
			if result.Summary.ProxyFaster {
				fmt.Printf("   ⚡ Proxy é mais rápido por %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render((-result.Summary.LatencyDiff).Round(time.Millisecond).String()))
			} else if result.Summary.LatencyDiff > 0 {
				fmt.Printf("   📊 Conexão direta é mais rápida por %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(result.Summary.LatencyDiff.Round(time.Millisecond).String()))
			}
		}
	}

	// Recommendation
	fmt.Println()
	fmt.Printf("   💡 %s\n", result.Summary.Recommendation)
	fmt.Printf("\n   Tempo total: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Println()
}
