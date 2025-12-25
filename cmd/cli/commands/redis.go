package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/redis"
	"github.com/spf13/cobra"
)

// NewRedisCommand creates the redis command
func NewRedisCommand(usecase *redis.RedisUsecase) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "redis",
		Short: "Ferramentas para Redis",
		Long: `Ferramentas para gerenciamento e diagnostico de Redis.

Subcomandos:
  ping   - Testa conectividade
  info   - Mostra informacoes do servidor
  keys   - Lista chaves por pattern
  flush  - Limpa um database`,
		Example: `  # Testar conectividade
  jorge redis ping localhost:6379

  # Ver informacoes do servidor
  jorge redis info localhost:6379

  # Listar chaves
  jorge redis keys localhost:6379 "user:*"

  # Com senha
  jorge redis info localhost:6379 --password mypass`,
	}

	cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Senha do Redis")

	// Ping subcommand
	pingCmd := &cobra.Command{
		Use:   "ping <address>",
		Short: "Testa conectividade com Redis",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			ok, latency, err := usecase.Ping(ctx, args[0])
			if err != nil {
				fmt.Printf("\n   %s Redis unreachable: %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
					err.Error())
				return nil
			}

			if ok {
				fmt.Printf("\n   %s PONG (%s)\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
					latency.Round(time.Microsecond))
			}
			return nil
		},
	}

	// Info subcommand
	infoCmd := &cobra.Command{
		Use:   "info <address>",
		Short: "Mostra informacoes do servidor Redis",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			info, err := usecase.GetInfo(ctx, args[0], password)
			if err != nil {
				return err
			}

			displayRedisInfo(info)
			return nil
		},
	}

	// Keys subcommand
	keysCmd := &cobra.Command{
		Use:   "keys <address> <pattern>",
		Short: "Lista chaves por pattern",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			count, _ := cmd.Flags().GetInt("count")
			result, err := usecase.ScanKeys(ctx, args[0], password, args[1], count)
			if err != nil {
				return err
			}

			displayRedisScanResult(result)
			return nil
		},
	}
	keysCmd.Flags().Int("count", 100, "Numero maximo de chaves")

	// Flush subcommand
	flushCmd := &cobra.Command{
		Use:   "flush <address>",
		Short: "Limpa um database Redis",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			db, _ := cmd.Flags().GetInt("db")
			confirm, _ := cmd.Flags().GetBool("yes")

			if !confirm {
				fmt.Printf("\n   %s Use --yes para confirmar o flush do database %d\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("!"),
					db)
				return nil
			}

			err := usecase.FlushDB(ctx, args[0], password, db)
			if err != nil {
				return err
			}

			fmt.Printf("\n   %s Database %d limpo com sucesso\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				db)
			return nil
		},
	}
	flushCmd.Flags().Int("db", 0, "Database a limpar")
	flushCmd.Flags().Bool("yes", false, "Confirmar operacao")

	cmd.AddCommand(pingCmd, infoCmd, keysCmd, flushCmd)

	return cmd
}

func displayRedisInfo(info *models.RedisInfo) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("📦 Redis Info"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(18)

	if !info.Connected {
		fmt.Printf("   %s Nao foi possivel conectar: %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			info.Error)
		return
	}

	// Server info
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   %s %s\n", labelStyle.Render("Address:"), info.Address)
	fmt.Printf("   %s %s\n", labelStyle.Render("Version:"), info.Version)
	fmt.Printf("   %s %s\n", labelStyle.Render("Mode:"), info.Mode)
	fmt.Printf("   %s %s\n", labelStyle.Render("Role:"), info.Role)
	fmt.Printf("   %s %s\n", labelStyle.Render("Uptime:"), formatDuration(time.Duration(info.Uptime)*time.Second))
	fmt.Println(strings.Repeat("─", 60))

	// Clients
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   👥 Clients"))
	fmt.Printf("   %s %d\n", labelStyle.Render("Connected:"), info.ConnectedClients)

	// Memory
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   💾 Memory"))
	fmt.Printf("   %s %s\n", labelStyle.Render("Used:"), info.UsedMemoryHuman)
	if info.MaxMemoryHuman != "" && info.MaxMemoryHuman != "0B" {
		fmt.Printf("   %s %s (%.1f%%)\n", labelStyle.Render("Max:"),
			info.MaxMemoryHuman, info.MemoryUsagePercent)
	}

	// Stats
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📊 Stats"))
	fmt.Printf("   %s %d\n", labelStyle.Render("Total Keys:"), info.TotalKeys)
	fmt.Printf("   %s %d\n", labelStyle.Render("Expired Keys:"), info.ExpiredKeys)
	fmt.Printf("   %s %d\n", labelStyle.Render("Evicted Keys:"), info.EvictedKeys)
	if info.HitRate > 0 {
		fmt.Printf("   %s %.1f%%\n", labelStyle.Render("Hit Rate:"), info.HitRate)
	}

	// Databases
	if len(info.Databases) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   🗄️  Databases"))
		for name, db := range info.Databases {
			fmt.Printf("   %s keys=%d, expires=%d\n",
				labelStyle.Render(name+":"), db.Keys, db.Expires)
		}
	}

	fmt.Println()
}

func displayRedisScanResult(result *models.RedisScanResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("🔑 Redis Keys"))
	fmt.Println()

	fmt.Printf("   Pattern: %s\n", result.Pattern)
	fmt.Printf("   Found: %d keys\n\n", result.Total)

	if result.Total == 0 {
		fmt.Println("   Nenhuma chave encontrada")
		fmt.Println()
		return
	}

	for _, key := range result.Keys {
		fmt.Printf("   • %s\n", key.Key)
	}
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
