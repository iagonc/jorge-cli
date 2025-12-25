package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/database"
	"github.com/spf13/cobra"
)

// NewDatabaseCommand creates the db command
func NewDatabaseCommand(usecase *database.DatabaseUsecase) *cobra.Command {
	var dbType string

	cmd := &cobra.Command{
		Use:   "db",
		Short: "Ferramentas para bancos de dados",
		Long: `Ferramentas para gerenciamento e diagnostico de bancos de dados.

Databases suportados:
  - PostgreSQL (postgres)
  - MySQL (mysql)

Subcomandos:
  connect  - Testa conexao
  tables   - Lista tabelas
  query    - Executa query (somente SELECT)
  health   - Health check`,
		Example: `  # Testar conexao PostgreSQL
  jorge db connect "postgres://user:pass@localhost:5432/mydb" --type postgres

  # Listar tabelas
  jorge db tables "postgres://user:pass@localhost:5432/mydb" --type postgres

  # Executar query
  jorge db query "postgres://..." --type postgres "SELECT * FROM users LIMIT 10"

  # Health check
  jorge db health "postgres://..." --type postgres`,
	}

	cmd.PersistentFlags().StringVar(&dbType, "type", "postgres", "Tipo de database (postgres, mysql)")

	// Connect subcommand
	connectCmd := &cobra.Command{
		Use:   "connect <dsn>",
		Short: "Testa conexao com o banco",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			info, err := usecase.Connect(ctx, args[0], dbType)
			if err != nil {
				return err
			}

			displayDatabaseInfo(info)
			return nil
		},
	}

	// Tables subcommand
	tablesCmd := &cobra.Command{
		Use:   "tables <dsn>",
		Short: "Lista tabelas do banco",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			tables, err := usecase.ListTables(ctx, args[0], dbType)
			if err != nil {
				return err
			}

			displayTables(tables)
			return nil
		},
	}

	// Query subcommand
	queryCmd := &cobra.Command{
		Use:   "query <dsn> <sql>",
		Short: "Executa query SELECT",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			limit, _ := cmd.Flags().GetInt("limit")
			result, err := usecase.Query(ctx, args[0], dbType, args[1], limit)
			if err != nil {
				return err
			}

			displayQueryResult(result)
			return nil
		},
	}
	queryCmd.Flags().Int("limit", 100, "Limite de linhas")

	// Health subcommand
	healthCmd := &cobra.Command{
		Use:   "health <dsn>",
		Short: "Health check do banco",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			health, err := usecase.HealthCheck(ctx, args[0], dbType)
			if err != nil {
				return err
			}

			displayDatabaseHealth(health)
			return nil
		},
	}

	cmd.AddCommand(connectCmd, tablesCmd, queryCmd, healthCmd)

	return cmd
}

func displayDatabaseInfo(info *models.DatabaseInfo) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🗄️  Database Info"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(18)

	if !info.Connected {
		fmt.Printf("   %s Nao foi possivel conectar: %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			info.Error)
		return
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   %s %s\n", labelStyle.Render("Type:"),
		lipgloss.NewStyle().Bold(true).Render(strings.ToUpper(info.Type)))
	fmt.Printf("   %s %s:%d\n", labelStyle.Render("Host:"), info.Host, info.Port)
	fmt.Printf("   %s %s\n", labelStyle.Render("Database:"), info.Database)
	fmt.Printf("   %s %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
		"Connected")
	fmt.Println(strings.Repeat("─", 60))

	// Version
	if info.Version != "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📋 Server"))
		// Truncate long version strings
		version := info.Version
		if len(version) > 50 {
			version = version[:50] + "..."
		}
		fmt.Printf("   %s %s\n", labelStyle.Render("Version:"), version)
	}

	// Connections
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("   👥 Connections"))
	fmt.Printf("   %s %d\n", labelStyle.Render("Active:"), info.ActiveConns)
	if info.MaxConns > 0 {
		fmt.Printf("   %s %d\n", labelStyle.Render("Max:"), info.MaxConns)
	}

	// Size
	if info.DatabaseSize != "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   💾 Storage"))
		fmt.Printf("   %s %s\n", labelStyle.Render("Database Size:"), info.DatabaseSize)
		if info.TableCount > 0 {
			fmt.Printf("   %s %d\n", labelStyle.Render("Tables:"), info.TableCount)
		}
	}

	fmt.Println()
}

func displayTables(tables []models.TableInfo) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📋 Tables"))
	fmt.Println()

	if len(tables) == 0 {
		fmt.Println("   Nenhuma tabela encontrada")
		fmt.Println()
		return
	}

	// Header
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %-30s %10s %15s %8s\n",
		lipgloss.NewStyle().Bold(true).Render("Name"),
		lipgloss.NewStyle().Bold(true).Render("Rows"),
		lipgloss.NewStyle().Bold(true).Render("Size"),
		lipgloss.NewStyle().Bold(true).Render("Indexes"))
	fmt.Println(strings.Repeat("─", 70))

	for _, t := range tables {
		fmt.Printf("   %-30s %10d %15s %8d\n",
			t.Name, t.RowCount, t.Size, t.IndexCount)
	}

	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   Total: %d tables\n", len(tables))
	fmt.Println()
}

func displayQueryResult(result *models.QueryResult) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📊 Query Result"))
	fmt.Println()

	if result.Error != "" {
		fmt.Printf("   %s %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error:"),
			result.Error)
		return
	}

	fmt.Printf("   Duration: %s | Rows: %d\n", result.Duration, result.RowCount)
	fmt.Println()

	if result.RowCount == 0 {
		fmt.Println("   Nenhum resultado")
		fmt.Println()
		return
	}

	// Calculate column widths
	widths := make(map[string]int)
	for _, col := range result.Columns {
		widths[col] = len(col)
	}
	for _, row := range result.Rows {
		for col, val := range row {
			valStr := fmt.Sprintf("%v", val)
			if len(valStr) > widths[col] {
				widths[col] = len(valStr)
			}
			if widths[col] > 40 {
				widths[col] = 40
			}
		}
	}

	// Print header
	fmt.Print("   ")
	for _, col := range result.Columns {
		fmt.Printf("%-*s  ", widths[col], lipgloss.NewStyle().Bold(true).Render(col))
	}
	fmt.Println()

	fmt.Print("   ")
	for _, col := range result.Columns {
		fmt.Print(strings.Repeat("─", widths[col]), "  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range result.Rows {
		fmt.Print("   ")
		for _, col := range result.Columns {
			val := fmt.Sprintf("%v", row[col])
			if len(val) > 40 {
				val = val[:37] + "..."
			}
			fmt.Printf("%-*s  ", widths[col], val)
		}
		fmt.Println()
	}

	fmt.Println()
}

func displayDatabaseHealth(health *models.DatabaseHealth) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🏥 Database Health"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(15)

	fmt.Printf("   %s %s\n", labelStyle.Render("Type:"), strings.ToUpper(health.Type))
	fmt.Printf("   %s %s\n", labelStyle.Render("Host:"), health.Host)

	if health.Healthy {
		fmt.Printf("   %s %s\n", labelStyle.Render("Status:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("HEALTHY"))
		fmt.Printf("   %s %s\n", labelStyle.Render("Latency:"), health.Latency)
	} else {
		fmt.Printf("   %s %s\n", labelStyle.Render("Status:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("UNHEALTHY"))
		fmt.Printf("   %s %s\n", labelStyle.Render("Error:"), health.Error)
	}

	fmt.Println()
}
