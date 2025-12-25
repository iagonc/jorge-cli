package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/status"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command
func NewStatusCommand(usecase *status.StatusUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status page e monitoramento",
		Long: `Ferramentas para criar e verificar status pages.

Subcomandos:
  check    - Verifica status de componentes
  init     - Gera config de exemplo
  single   - Verifica um unico endpoint`,
		Example: `  # Verificar status de todos os componentes
  jorge status check status.yaml

  # Gerar config de exemplo
  jorge status init > status.yaml

  # Verificar um unico endpoint
  jorge status single https://api.example.com --type http`,
	}

	// Check subcommand
	checkCmd := &cobra.Command{
		Use:   "check <config-file>",
		Short: "Verifica status de componentes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			result, err := usecase.CheckAll(ctx, config)
			if err != nil {
				return err
			}

			displayStatusPageResult(result)
			return nil
		},
	}

	// Init subcommand
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Gera config de exemplo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usecase.GenerateConfig())
		},
	}

	// Single subcommand
	singleCmd := &cobra.Command{
		Use:   "single <target>",
		Short: "Verifica um unico endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			checkType, _ := cmd.Flags().GetString("type")
			result := usecase.CheckSingle(ctx, args[0], checkType)

			displaySingleStatus(args[0], checkType, result)
			return nil
		},
	}
	singleCmd.Flags().String("type", "http", "Tipo de check (http, tcp, dns)")

	// Watch subcommand - continuous monitoring
	watchCmd := &cobra.Command{
		Use:   "watch <config-file>",
		Short: "Monitora continuamente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			interval, _ := cmd.Flags().GetInt("interval")

			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			// Initial check
			ctx := context.Background()
			result, err := usecase.CheckAll(ctx, config)
			if err != nil {
				return err
			}
			clearScreen()
			displayStatusPageResult(result)

			// Continuous monitoring
			for range ticker.C {
				result, err := usecase.CheckAll(ctx, config)
				if err != nil {
					continue
				}
				clearScreen()
				displayStatusPageResult(result)
			}

			return nil
		},
	}
	watchCmd.Flags().Int("interval", 30, "Intervalo em segundos")

	cmd.AddCommand(checkCmd, initCmd, singleCmd, watchCmd)

	return cmd
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func displayStatusPageResult(result *models.StatusPageResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 " + result.Title))
	if result.Description != "" {
		fmt.Printf("   %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(result.Description))
	}
	fmt.Println()

	// Overall status
	overallStyle := getStatusStyle(result.OverallStatus)
	fmt.Printf("   Overall: %s\n", overallStyle.Render(string(result.OverallStatus)))
	fmt.Printf("   Last Updated: %s\n", result.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Group components
	groups := make(map[string][]models.StatusComponent)
	for _, comp := range result.Components {
		group := comp.Group
		if group == "" {
			group = "Services"
		}
		groups[group] = append(groups[group], comp)
	}

	// Display by group
	for group, components := range groups {
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("   %s\n", lipgloss.NewStyle().Bold(true).Render(group))
		fmt.Println(strings.Repeat("─", 60))

		for _, comp := range components {
			statusStyle := getStatusStyle(comp.Status.State)
			statusIcon := getStatusIcon(comp.Status.State)

			fmt.Printf("   %s %-25s %s",
				statusIcon,
				comp.Name,
				statusStyle.Render(string(comp.Status.State)))

			if comp.Status.Latency != "" {
				fmt.Printf("  (%s)", comp.Status.Latency)
			}
			fmt.Println()

			if comp.Status.Message != "" && comp.Status.State != models.StatusOperational {
				fmt.Printf("      %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(comp.Status.Message))
			}
		}
		fmt.Println()
	}

	// Active incidents
	if len(result.ActiveIncidents) > 0 {
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("   %s Active Incidents\n",
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("⚠"))
		fmt.Println(strings.Repeat("─", 60))

		for _, incident := range result.ActiveIncidents {
			impactStyle := getImpactStyle(incident.Impact)
			fmt.Printf("   %s [%s] %s\n",
				impactStyle.Render("●"),
				incident.Status,
				incident.Title)
			fmt.Printf("      Started: %s\n",
				incident.CreatedAt.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	}

	// Write to file if specified
	if os.Getenv("STATUS_OUTPUT") != "" {
		// Could export to JSON/HTML for static hosting
	}
}

func displaySingleStatus(target, checkType string, status models.ComponentStatus) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 Status Check"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(12)
	statusStyle := getStatusStyle(status.State)
	statusIcon := getStatusIcon(status.State)

	fmt.Printf("   %s %s\n", labelStyle.Render("Target:"), target)
	fmt.Printf("   %s %s\n", labelStyle.Render("Type:"), checkType)
	fmt.Printf("   %s %s %s\n", labelStyle.Render("Status:"),
		statusIcon, statusStyle.Render(string(status.State)))
	fmt.Printf("   %s %s\n", labelStyle.Render("Latency:"), status.Latency)
	if status.Message != "" {
		fmt.Printf("   %s %s\n", labelStyle.Render("Message:"), status.Message)
	}
	fmt.Printf("   %s %s\n", labelStyle.Render("Checked:"),
		status.LastChecked.Format("2006-01-02 15:04:05"))
	fmt.Println()
}

func getStatusStyle(state models.StatusState) lipgloss.Style {
	switch state {
	case models.StatusOperational:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	case models.StatusDegraded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case models.StatusPartialOutage:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	case models.StatusMajorOutage:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	case models.StatusMaintenance:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}

func getStatusIcon(state models.StatusState) string {
	switch state {
	case models.StatusOperational:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	case models.StatusDegraded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("●")
	case models.StatusPartialOutage:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("●")
	case models.StatusMajorOutage:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("●")
	case models.StatusMaintenance:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("○")
	}
}

func getImpactStyle(impact string) lipgloss.Style {
	switch impact {
	case "critical":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case "major":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case "minor":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}
