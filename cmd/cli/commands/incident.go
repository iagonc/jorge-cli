package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/incident"
	"github.com/spf13/cobra"
)

// NewIncidentCommand creates the incident command
func NewIncidentCommand(usecase *incident.IncidentUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incident",
		Short: "Gerenciamento de incidentes",
		Long: `Ferramentas para gerenciamento de incidentes.

Subcomandos:
  create  - Cria novo incidente
  list    - Lista incidentes
  get     - Mostra detalhes de um incidente
  update  - Atualiza status
  note    - Adiciona nota
  summary - Mostra estatisticas`,
		Example: `  # Criar incidente
  jorge incident create "API Down" --severity critical

  # Listar incidentes abertos
  jorge incident list --status open

  # Atualizar status
  jorge incident update INC-123 --status investigating

  # Adicionar nota
  jorge incident note INC-123 "Identificado problema no DB"

  # Ver estatisticas
  jorge incident summary`,
	}

	// Create subcommand
	createCmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Cria novo incidente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description, _ := cmd.Flags().GetString("description")
			severityStr, _ := cmd.Flags().GetString("severity")

			severity := models.IncidentSeverity(severityStr)

			inc, err := usecase.Create(args[0], description, severity)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Incidente criado: %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				lipgloss.NewStyle().Bold(true).Render(inc.ID))

			displayIncident(inc)
			return nil
		},
	}
	createCmd.Flags().StringP("description", "d", "", "Descricao do incidente")
	createCmd.Flags().StringP("severity", "s", "medium", "Severidade (critical, high, medium, low)")

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista incidentes",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			limit, _ := cmd.Flags().GetInt("limit")

			list, err := usecase.List(status, limit)
			if err != nil {
				return err
			}

			displayIncidentList(list)
			return nil
		},
	}
	listCmd.Flags().String("status", "", "Filtrar por status")
	listCmd.Flags().Int("limit", 20, "Limite de resultados")

	// Get subcommand
	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Mostra detalhes de um incidente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inc, err := usecase.Get(args[0])
			if err != nil {
				return err
			}

			displayIncident(inc)
			return nil
		},
	}

	// Update subcommand
	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Atualiza status do incidente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			statusStr, _ := cmd.Flags().GetString("status")
			note, _ := cmd.Flags().GetString("note")

			if statusStr == "" {
				return fmt.Errorf("--status is required")
			}

			status := models.IncidentStatus(statusStr)
			inc, err := usecase.UpdateStatus(args[0], status, note)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Incidente atualizado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))

			displayIncident(inc)
			return nil
		},
	}
	updateCmd.Flags().String("status", "", "Novo status (investigating, identified, monitoring, resolved)")
	updateCmd.Flags().String("note", "", "Nota adicional")

	// Note subcommand
	noteCmd := &cobra.Command{
		Use:   "note <id> <message>",
		Short: "Adiciona nota ao incidente",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			author, _ := cmd.Flags().GetString("author")

			_, err := usecase.AddNote(args[0], args[1], author)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Nota adicionada\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			return nil
		},
	}
	noteCmd.Flags().String("author", "", "Autor da nota")

	// Commander subcommand
	commanderCmd := &cobra.Command{
		Use:   "commander <id> <name>",
		Short: "Define o incident commander",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := usecase.SetCommander(args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Incident commander definido: %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				args[1])
			return nil
		},
	}

	// Summary subcommand
	summaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "Mostra estatisticas de incidentes",
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := usecase.GetSummary()
			if err != nil {
				return err
			}

			displayIncidentSummary(summary)
			return nil
		},
	}

	// Export subcommand
	exportCmd := &cobra.Command{
		Use:   "export <id>",
		Short: "Exporta incidente",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			output, err := usecase.Export(args[0], format)
			if err != nil {
				return err
			}

			fmt.Println(output)
			return nil
		},
	}
	exportCmd.Flags().String("format", "yaml", "Formato (json, yaml)")

	cmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, noteCmd, commanderCmd, summaryCmd, exportCmd)

	return cmd
}

func displayIncident(inc *models.Incident) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🚨 Incident: " + inc.ID))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(15)
	severityStyle := getSeverityStyle(string(inc.Severity))
	statusStyle := getIncidentStatusStyle(string(inc.Status))

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   %s %s\n", labelStyle.Render("Title:"),
		lipgloss.NewStyle().Bold(true).Render(inc.Title))
	fmt.Printf("   %s %s\n", labelStyle.Render("Severity:"),
		severityStyle.Render(strings.ToUpper(string(inc.Severity))))
	fmt.Printf("   %s %s\n", labelStyle.Render("Status:"),
		statusStyle.Render(string(inc.Status)))

	if inc.Commander != "" {
		fmt.Printf("   %s %s\n", labelStyle.Render("Commander:"), inc.Commander)
	}

	fmt.Printf("   %s %s\n", labelStyle.Render("Created:"),
		inc.CreatedAt.Format("2006-01-02 15:04:05"))

	if inc.ResolvedAt != nil {
		fmt.Printf("   %s %s\n", labelStyle.Render("Resolved:"),
			inc.ResolvedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   %s %s\n", labelStyle.Render("Duration:"), inc.Duration)
	}
	fmt.Println(strings.Repeat("─", 60))

	if inc.Description != "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📝 Description"))
		fmt.Printf("   %s\n", inc.Description)
	}

	if len(inc.Services) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   🔧 Affected Services"))
		for _, svc := range inc.Services {
			fmt.Printf("   • %s\n", svc)
		}
	}

	if len(inc.Timeline) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📅 Timeline"))
		for _, event := range inc.Timeline {
			fmt.Printf("   %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(event.Timestamp.Format("15:04")),
				event.Description)
		}
	}

	fmt.Println()
}

func displayIncidentList(list *models.IncidentList) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🚨 Incidents"))
	fmt.Println()

	fmt.Printf("   Total: %d | Open: %d | Resolved: %d\n\n", list.Total, list.Open, list.Resolved)

	if len(list.Incidents) == 0 {
		fmt.Println("   Nenhum incidente encontrado")
		fmt.Println()
		return
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("   %-12s %-10s %-12s %-30s %s\n",
		lipgloss.NewStyle().Bold(true).Render("ID"),
		lipgloss.NewStyle().Bold(true).Render("Severity"),
		lipgloss.NewStyle().Bold(true).Render("Status"),
		lipgloss.NewStyle().Bold(true).Render("Title"),
		lipgloss.NewStyle().Bold(true).Render("Created"))
	fmt.Println(strings.Repeat("─", 80))

	for _, inc := range list.Incidents {
		severityStyle := getSeverityStyle(string(inc.Severity))
		statusStyle := getIncidentStatusStyle(string(inc.Status))

		title := inc.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}

		fmt.Printf("   %-12s %-10s %-12s %-30s %s\n",
			inc.ID,
			severityStyle.Render(string(inc.Severity)),
			statusStyle.Render(string(inc.Status)),
			title,
			inc.CreatedAt.Format("2006-01-02"))
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
}

func displayIncidentSummary(summary *models.IncidentSummary) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 Incident Summary"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(18)

	fmt.Printf("   %s %d\n", labelStyle.Render("Total Incidents:"), summary.TotalIncidents)

	if summary.MTTR != "" {
		fmt.Printf("   %s %s\n", labelStyle.Render("MTTR:"), summary.MTTR)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   By Status:"))
	for status, count := range summary.ByStatus {
		fmt.Printf("      %-15s %d\n", status+":", count)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   By Severity:"))
	for severity, count := range summary.BySeverity {
		style := getSeverityStyle(severity)
		fmt.Printf("      %-15s %d\n", style.Render(severity)+":  ", count)
	}

	fmt.Println()
}

func getSeverityStyle(severity string) lipgloss.Style {
	switch severity {
	case "critical":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	case "high":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	case "medium":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "low":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}

func getIncidentStatusStyle(status string) lipgloss.Style {
	switch status {
	case "open":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case "investigating":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "identified":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	case "monitoring":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	case "resolved":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}
