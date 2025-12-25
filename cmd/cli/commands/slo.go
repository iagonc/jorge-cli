package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/slo"
	"github.com/spf13/cobra"
)

// NewSLOCommand creates the slo command
func NewSLOCommand(usecase *slo.SLOUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slo",
		Short: "Gerenciamento de SLOs",
		Long: `Ferramentas para gerenciamento de Service Level Objectives.

Subcomandos:
  dashboard - Mostra dashboard de SLOs
  calc      - Calcula error budget
  downtime  - Calcula downtime permitido
  init      - Gera config de exemplo`,
		Example: `  # Ver dashboard de SLOs
  jorge slo dashboard slo.yaml

  # Calcular error budget
  jorge slo calc --target 99.9 --current 99.85 --window 30d

  # Calcular downtime permitido
  jorge slo downtime --target 99.9 --window 30d

  # Gerar config de exemplo
  jorge slo init > slo.yaml`,
	}

	// Dashboard subcommand
	dashboardCmd := &cobra.Command{
		Use:   "dashboard <config-file>",
		Short: "Mostra dashboard de SLOs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			dashboard := usecase.GetDashboard(config)
			displaySLODashboard(dashboard)
			return nil
		},
	}

	// Calc subcommand
	calcCmd := &cobra.Command{
		Use:   "calc",
		Short: "Calcula error budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetFloat64("target")
			current, _ := cmd.Flags().GetFloat64("current")
			window, _ := cmd.Flags().GetString("window")

			sloObj := models.SLO{
				Name:   "Manual Calculation",
				Target: target,
				Window: window,
			}

			budget := usecase.CalculateBudget(sloObj, current)
			displayErrorBudget(budget, target, current)
			return nil
		},
	}
	calcCmd.Flags().Float64("target", 99.9, "SLO target (%)")
	calcCmd.Flags().Float64("current", 99.95, "Current SLI (%)")
	calcCmd.Flags().String("window", "30d", "Window (e.g., 30d, 7d)")

	// Downtime subcommand
	downtimeCmd := &cobra.Command{
		Use:   "downtime",
		Short: "Calcula downtime permitido",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetFloat64("target")
			window, _ := cmd.Flags().GetString("window")

			allowance := usecase.CalculateDowntimeAllowance(target, window)
			displayDowntimeAllowance(allowance)
			return nil
		},
	}
	downtimeCmd.Flags().Float64("target", 99.9, "SLO target (%)")
	downtimeCmd.Flags().String("window", "30d", "Window (e.g., 30d, 7d)")

	// Report subcommand
	reportCmd := &cobra.Command{
		Use:   "report <config-file>",
		Short: "Gera relatorio de SLOs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			period, _ := cmd.Flags().GetString("period")

			for _, sloObj := range config.SLOs {
				// Simulated data - in production would query metrics
				totalEvents := int64(100000)
				goodEvents := int64(float64(totalEvents) * (sloObj.Target + 0.2) / 100)

				report := usecase.GenerateReport(sloObj, goodEvents, totalEvents, period)
				displaySLOReport(report)
			}

			return nil
		},
	}
	reportCmd.Flags().String("period", "7d", "Report period")

	// Init subcommand
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Gera config de exemplo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usecase.GenerateConfig())
		},
	}

	cmd.AddCommand(dashboardCmd, calcCmd, downtimeCmd, reportCmd, initCmd)

	return cmd
}

func displaySLODashboard(dashboard *models.SLODashboard) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 SLO Dashboard: " + dashboard.Service))
	fmt.Println()

	// Overall health
	healthStyle := getHealthStyle(dashboard.OverallHealth)
	fmt.Printf("   Overall Health: %s\n", healthStyle.Render(strings.ToUpper(dashboard.OverallHealth)))
	fmt.Printf("   Last Updated: %s\n\n", dashboard.LastUpdated.Format("2006-01-02 15:04:05"))

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("   %-25s %-12s %-12s %-15s %s\n",
		lipgloss.NewStyle().Bold(true).Render("SLO"),
		lipgloss.NewStyle().Bold(true).Render("Target"),
		lipgloss.NewStyle().Bold(true).Render("Current"),
		lipgloss.NewStyle().Bold(true).Render("Budget"),
		lipgloss.NewStyle().Bold(true).Render("Status"))
	fmt.Println(strings.Repeat("─", 80))

	for _, report := range dashboard.SLOs {
		budgetStyle := getBudgetStatusStyle(report.ErrorBudget.Status)
		complianceIcon := "✓"
		if !report.Compliance {
			complianceIcon = "✗"
		}

		fmt.Printf("   %-25s %-12s %-12s %-15s %s\n",
			truncateStr(report.SLO.Name, 23),
			fmt.Sprintf("%.2f%%", report.Target),
			fmt.Sprintf("%.2f%%", report.CurrentSLI),
			fmt.Sprintf("%.1f%% remaining", report.ErrorBudget.BudgetPercent),
			budgetStyle.Render(complianceIcon+" "+report.ErrorBudget.Status))
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
}

func displayErrorBudget(budget models.ErrorBudget, target, current float64) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 Error Budget"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(20)
	statusStyle := getBudgetStatusStyle(budget.Status)

	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("   %s %.2f%%\n", labelStyle.Render("Target SLO:"), target)
	fmt.Printf("   %s %.2f%%\n", labelStyle.Render("Current SLI:"), current)
	fmt.Printf("   %s %s\n", labelStyle.Render("Status:"), statusStyle.Render(strings.ToUpper(budget.Status)))
	fmt.Println(strings.Repeat("─", 50))

	fmt.Println()
	fmt.Printf("   %s %.3f%%\n", labelStyle.Render("Total Budget:"), budget.BudgetTotal)
	fmt.Printf("   %s %.3f%%\n", labelStyle.Render("Used:"), budget.BudgetUsed)
	fmt.Printf("   %s %.3f%% (%.1f%% of total)\n",
		labelStyle.Render("Remaining:"),
		budget.BudgetRemaining,
		budget.BudgetPercent)
	fmt.Println()

	// Visual budget bar
	barWidth := 40
	usedWidth := int(float64(barWidth) * (budget.BudgetUsed / budget.BudgetTotal))
	if usedWidth > barWidth {
		usedWidth = barWidth
	}
	remainingWidth := barWidth - usedWidth

	bar := strings.Repeat("█", usedWidth) + strings.Repeat("░", remainingWidth)
	fmt.Printf("   [%s]\n", bar)
	fmt.Println()
}

func displayDowntimeAllowance(allowance map[string]string) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⏱️  Downtime Allowance"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(22)

	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("   %s %s\n", labelStyle.Render("Target:"), allowance["target"])
	fmt.Printf("   %s %s\n", labelStyle.Render("Window:"), allowance["window"])
	fmt.Println(strings.Repeat("─", 50))

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   Allowed Downtime:"))
	fmt.Printf("   %s %s\n", labelStyle.Render("Total (in window):"), allowance["allowed_downtime"])
	fmt.Printf("   %s %s\n", labelStyle.Render("Per Month:"), allowance["monthly_allowance"])
	fmt.Printf("   %s %s\n", labelStyle.Render("Per Week:"), allowance["weekly_allowance"])
	fmt.Printf("   %s %s\n", labelStyle.Render("Per Day:"), allowance["daily_allowance"])
	fmt.Println()
}

func displaySLOReport(report models.SLOReport) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 SLO Report: " + report.SLO.Name))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(18)

	complianceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	complianceText := "COMPLIANT"
	if !report.Compliance {
		complianceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		complianceText = "NON-COMPLIANT"
	}

	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("   %s %s\n", labelStyle.Render("Period:"), report.Period)
	fmt.Printf("   %s %.2f%%\n", labelStyle.Render("Target:"), report.Target)
	fmt.Printf("   %s %.2f%%\n", labelStyle.Render("Current:"), report.CurrentSLI)
	fmt.Printf("   %s %s\n", labelStyle.Render("Compliance:"), complianceStyle.Render(complianceText))
	fmt.Println(strings.Repeat("─", 50))

	fmt.Println()
	fmt.Printf("   %s %d\n", labelStyle.Render("Total Events:"), report.TotalEvents)
	fmt.Printf("   %s %d\n", labelStyle.Render("Good Events:"), report.GoodEvents)
	fmt.Printf("   %s %d\n", labelStyle.Render("Bad Events:"), report.BadEvents)
	fmt.Println()
}

func getHealthStyle(health string) lipgloss.Style {
	switch health {
	case "healthy":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	case "warning":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case "critical":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}

func getBudgetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "healthy":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case "warning":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "critical":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case "exhausted":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}
