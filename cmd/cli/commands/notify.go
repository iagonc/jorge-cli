package commands

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/notify"
	"github.com/spf13/cobra"
)

// NewNotifyCommand creates the notify command
func NewNotifyCommand(usecase *notify.NotifyUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Envia notificacoes",
		Long: `Envia notificacoes para diversos canais.

Canais suportados:
  - Slack
  - Discord
  - PagerDuty
  - Webhook generico

Subcomandos:
  send    - Envia notificacao usando config
  slack   - Envia direto para Slack
  init    - Gera config de exemplo`,
		Example: `  # Enviar usando config
  jorge notify send notify.yaml --title "Deploy" --message "v1.2.3 deployed"

  # Enviar direto para Slack
  jorge notify slack "Deploy v1.2.3 concluido" --webhook $SLACK_WEBHOOK

  # Gerar config de exemplo
  jorge notify init > notify.yaml`,
	}

	// Send subcommand
	sendCmd := &cobra.Command{
		Use:   "send <config-file>",
		Short: "Envia notificacao usando config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			title, _ := cmd.Flags().GetString("title")
			message, _ := cmd.Flags().GetString("message")
			severity, _ := cmd.Flags().GetString("severity")
			service, _ := cmd.Flags().GetString("service")

			notif := &models.Notification{
				Title:     title,
				Message:   message,
				Severity:  severity,
				Service:   service,
				Timestamp: time.Now(),
			}

			result := usecase.SendAll(config, notif)
			displayNotifyBatchResult(result)
			return nil
		},
	}
	sendCmd.Flags().String("title", "Notification", "Titulo da notificacao")
	sendCmd.Flags().StringP("message", "m", "", "Mensagem")
	sendCmd.Flags().String("severity", "", "Severidade (critical, high, medium, low)")
	sendCmd.Flags().String("service", "", "Servico afetado")

	// Slack subcommand
	slackCmd := &cobra.Command{
		Use:   "slack <message>",
		Short: "Envia notificacao para Slack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhook, _ := cmd.Flags().GetString("webhook")

			result := usecase.QuickSlack(webhook, args[0])
			displayNotifyResult(result)
			return nil
		},
	}
	slackCmd.Flags().String("webhook", "", "Slack webhook URL (ou use SLACK_WEBHOOK_URL env)")

	// Discord subcommand
	discordCmd := &cobra.Command{
		Use:   "discord <message>",
		Short: "Envia notificacao para Discord",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhook, _ := cmd.Flags().GetString("webhook")

			if webhook == "" {
				return fmt.Errorf("--webhook is required")
			}

			config := &models.DiscordConfig{WebhookURL: webhook}
			notif := &models.Notification{
				Title:     "Notification",
				Message:   args[0],
				Timestamp: time.Now(),
			}

			result := usecase.SendDiscord(config, notif)
			displayNotifyResult(result)
			return nil
		},
	}
	discordCmd.Flags().String("webhook", "", "Discord webhook URL")

	// Webhook subcommand
	webhookCmd := &cobra.Command{
		Use:   "webhook <url> <message>",
		Short: "Envia notificacao para webhook generico",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			title, _ := cmd.Flags().GetString("title")

			config := &models.WebhookConfig{
				URL:    args[0],
				Method: "POST",
			}
			notif := &models.Notification{
				Title:     title,
				Message:   args[1],
				Timestamp: time.Now(),
			}

			result := usecase.SendWebhook(config, notif)
			displayNotifyResult(result)
			return nil
		},
	}
	webhookCmd.Flags().String("title", "Notification", "Titulo")

	// Test subcommand
	testCmd := &cobra.Command{
		Use:   "test <config-file>",
		Short: "Testa configuracao enviando mensagem de teste",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := usecase.LoadConfig(args[0])
			if err != nil {
				return err
			}

			notif := &models.Notification{
				Title:     "Test Notification",
				Message:   "This is a test message from Jorge CLI",
				Severity:  "info",
				Service:   "jorge-cli",
				Timestamp: time.Now(),
			}

			result := usecase.SendAll(config, notif)
			displayNotifyBatchResult(result)
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

	cmd.AddCommand(sendCmd, slackCmd, discordCmd, webhookCmd, testCmd, initCmd)

	return cmd
}

func displayNotifyResult(result *models.NotificationResult) {
	fmt.Println()

	if result.Success {
		fmt.Printf("   %s %s: Notificacao enviada\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
			result.Provider)
	} else {
		fmt.Printf("   %s %s: Falha - %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Provider,
			result.Error)
	}

	if result.StatusCode > 0 {
		fmt.Printf("      HTTP Status: %d\n", result.StatusCode)
	}

	fmt.Println()
}

func displayNotifyBatchResult(result *models.NotificationBatchResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📤 Notification Results"))
	fmt.Println()

	for _, r := range result.Results {
		if r.Success {
			fmt.Printf("   %s %-12s Enviado\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				r.Provider)
		} else {
			fmt.Printf("   %s %-12s Falhou: %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
				r.Provider,
				r.Error)
		}
	}

	fmt.Println()
	fmt.Printf("   Succeeded: %d | Failed: %d\n\n",
		result.Succeeded, result.Failed)
}
