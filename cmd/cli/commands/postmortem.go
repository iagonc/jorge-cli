package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/postmortem"
	"github.com/spf13/cobra"
)

// NewPostmortemCommand creates the postmortem command
func NewPostmortemCommand(usecase *postmortem.PostmortemUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "postmortem",
		Aliases: []string{"pm"},
		Short:   "Gerenciamento de postmortems",
		Long: `Ferramentas para criar e gerenciar postmortems de incidentes.

Subcomandos:
  create   - Cria novo postmortem
  list     - Lista postmortems
  get      - Mostra detalhes
  update   - Atualiza status
  cause    - Adiciona root cause
  action   - Adiciona action item
  export   - Exporta para markdown/json
  template - Gera template`,
		Example: `  # Criar postmortem
  jorge postmortem create INC-123 "Database Outage" --severity critical

  # Adicionar root cause
  jorge postmortem cause PM-123 "Connection pool exhaustion"

  # Adicionar action item
  jorge postmortem action PM-123 "Implement connection pooling" --type prevent --priority P1 --owner "@dev"

  # Exportar para markdown
  jorge postmortem export PM-123 --format markdown`,
	}

	// Create subcommand
	createCmd := &cobra.Command{
		Use:   "create <incident-id> <title>",
		Short: "Cria novo postmortem",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			severity, _ := cmd.Flags().GetString("severity")

			pm, err := usecase.Create(args[0], args[1], severity)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Postmortem criado: %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				lipgloss.NewStyle().Bold(true).Render(pm.ID))

			return nil
		},
	}
	createCmd.Flags().String("severity", "medium", "Severidade (critical, high, medium, low)")

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista postmortems",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			limit, _ := cmd.Flags().GetInt("limit")

			list, err := usecase.List(status, limit)
			if err != nil {
				return err
			}

			displayPostmortemList(list)
			return nil
		},
	}
	listCmd.Flags().String("status", "", "Filtrar por status")
	listCmd.Flags().Int("limit", 20, "Limite de resultados")

	// Get subcommand
	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Mostra detalhes do postmortem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := usecase.Get(args[0])
			if err != nil {
				return err
			}

			displayPostmortem(pm)
			return nil
		},
	}

	// Update subcommand
	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Atualiza status do postmortem",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			statusStr, _ := cmd.Flags().GetString("status")

			if statusStr == "" {
				return fmt.Errorf("--status is required")
			}

			status := models.PostmortemStatus(statusStr)
			pm, err := usecase.UpdateStatus(args[0], status)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Postmortem atualizado para: %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				pm.Status)

			return nil
		},
	}
	updateCmd.Flags().String("status", "", "Novo status (draft, in_review, approved, published)")

	// Summary subcommand
	summaryCmd := &cobra.Command{
		Use:   "summary <id> <text>",
		Short: "Define o summary do postmortem",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := usecase.SetSummary(args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Summary atualizado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			return nil
		},
	}

	// Cause subcommand
	causeCmd := &cobra.Command{
		Use:   "cause <id> <description>",
		Short: "Adiciona root cause",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := usecase.AddRootCause(args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Root cause adicionado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			return nil
		},
	}

	// Action subcommand
	actionCmd := &cobra.Command{
		Use:   "action <id> <description>",
		Short: "Adiciona action item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			actionType, _ := cmd.Flags().GetString("type")
			priority, _ := cmd.Flags().GetString("priority")
			owner, _ := cmd.Flags().GetString("owner")

			item := models.ActionItem{
				Description: args[1],
				Type:        actionType,
				Priority:    priority,
				Owner:       owner,
			}

			pm, err := usecase.AddActionItem(args[0], item)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Action item adicionado: %s\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				pm.ActionItems[len(pm.ActionItems)-1].ID)
			return nil
		},
	}
	actionCmd.Flags().String("type", "prevent", "Tipo (prevent, detect, mitigate, process)")
	actionCmd.Flags().String("priority", "P2", "Prioridade (P0, P1, P2, P3)")
	actionCmd.Flags().String("owner", "", "Responsavel")

	// Timeline subcommand
	timelineCmd := &cobra.Command{
		Use:   "timeline <id> <time> <description>",
		Short: "Adiciona entrada no timeline",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			entryType, _ := cmd.Flags().GetString("type")

			// Parse time (HH:MM format)
			t, err := time.Parse("15:04", args[1])
			if err != nil {
				return fmt.Errorf("invalid time format, use HH:MM")
			}

			// Set date to today
			now := time.Now()
			t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())

			entry := models.TimelineEntry{
				Time:        t,
				Description: args[2],
				Type:        entryType,
			}

			_, err = usecase.AddTimeline(args[0], entry)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Timeline entry adicionado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			return nil
		},
	}
	timelineCmd.Flags().String("type", "response", "Tipo (detection, response, mitigation, resolution)")

	// Lesson subcommand
	lessonCmd := &cobra.Command{
		Use:   "lesson <id> <description>",
		Short: "Adiciona lesson learned",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := usecase.AddLesson(args[0], args[1])
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Lesson learned adicionado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			return nil
		},
	}

	// Export subcommand
	exportCmd := &cobra.Command{
		Use:   "export <id>",
		Short: "Exporta postmortem",
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
	exportCmd.Flags().String("format", "markdown", "Formato (json, yaml, markdown)")

	// Template subcommand
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Gera template de postmortem",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usecase.GenerateTemplate())
		},
	}

	cmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, summaryCmd, causeCmd, actionCmd, timelineCmd, lessonCmd, exportCmd, templateCmd)

	return cmd
}

func displayPostmortem(pm *models.Postmortem) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📋 Postmortem: " + pm.ID))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(15)

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   %s %s\n", labelStyle.Render("Title:"),
		lipgloss.NewStyle().Bold(true).Render(pm.Title))
	fmt.Printf("   %s %s\n", labelStyle.Render("Incident:"), pm.IncidentID)
	fmt.Printf("   %s %s\n", labelStyle.Render("Severity:"), pm.Severity)
	fmt.Printf("   %s %s\n", labelStyle.Render("Status:"), pm.Status)
	fmt.Printf("   %s %s\n", labelStyle.Render("Date:"), pm.Date.Format("2006-01-02"))
	fmt.Println(strings.Repeat("─", 60))

	if pm.Summary != "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📝 Summary"))
		fmt.Printf("   %s\n", pm.Summary)
	}

	if len(pm.RootCauses) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   🔍 Root Causes"))
		for i, cause := range pm.RootCauses {
			fmt.Printf("   %d. %s\n", i+1, cause)
		}
	}

	if len(pm.Timeline) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📅 Timeline"))
		for _, entry := range pm.Timeline {
			fmt.Printf("   %s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(entry.Time.Format("15:04")),
				entry.Description)
		}
	}

	if len(pm.ActionItems) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   ✅ Action Items"))
		for _, item := range pm.ActionItems {
			statusIcon := "○"
			if item.Status == "done" {
				statusIcon = "●"
			} else if item.Status == "in_progress" {
				statusIcon = "◐"
			}
			fmt.Printf("   %s [%s] %s (%s) - %s\n",
				statusIcon, item.Priority, item.Description, item.Type, item.Owner)
		}
	}

	if len(pm.LessonsLearned) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   💡 Lessons Learned"))
		for _, lesson := range pm.LessonsLearned {
			fmt.Printf("   • %s\n", lesson)
		}
	}

	fmt.Println()
}

func displayPostmortemList(list *models.PostmortemList) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📋 Postmortems"))
	fmt.Println()

	if len(list.Postmortems) == 0 {
		fmt.Println("   Nenhum postmortem encontrado")
		fmt.Println()
		return
	}

	fmt.Printf("   Total: %d\n\n", list.Total)

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("   %-12s %-12s %-12s %-30s %s\n",
		lipgloss.NewStyle().Bold(true).Render("ID"),
		lipgloss.NewStyle().Bold(true).Render("Incident"),
		lipgloss.NewStyle().Bold(true).Render("Status"),
		lipgloss.NewStyle().Bold(true).Render("Title"),
		lipgloss.NewStyle().Bold(true).Render("Date"))
	fmt.Println(strings.Repeat("─", 80))

	for _, pm := range list.Postmortems {
		title := pm.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}

		fmt.Printf("   %-12s %-12s %-12s %-30s %s\n",
			pm.ID,
			pm.IncidentID,
			pm.Status,
			title,
			pm.Date.Format("2006-01-02"))
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Println()
}
