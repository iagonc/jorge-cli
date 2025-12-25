package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/chaos"
	"github.com/spf13/cobra"
)

// NewChaosCommand creates the chaos command
func NewChaosCommand(usecase *chaos.ChaosUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chaos",
		Short: "Chaos engineering experiments",
		Long: `Ferramentas para chaos engineering e testes de resiliencia.

Tipos de experimentos:
  - CPU stress
  - Memory stress
  - Network latency
  - HTTP fault injection

Subcomandos:
  run      - Executa um experimento
  scenario - Executa cenario de arquivo
  list     - Lista experimentos em execucao
  stop     - Para um experimento
  init     - Gera cenario de exemplo`,
		Example: `  # CPU stress por 30s
  jorge chaos run cpu --duration 30s --workers 4

  # Memory stress
  jorge chaos run memory --duration 60s --size 256MB

  # HTTP fault injection
  jorge chaos run http-fault --port 8888 --error-rate 50

  # Executar cenario
  jorge chaos scenario chaos.yaml`,
	}

	// Run subcommand
	runCmd := &cobra.Command{
		Use:   "run <type>",
		Short: "Executa experimento",
		Long: `Tipos disponiveis:
  cpu      - Stress de CPU
  memory   - Stress de memoria
  http-fault - Injecao de falhas HTTP`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration, _ := cmd.Flags().GetString("duration")
			workers, _ := cmd.Flags().GetInt("workers")
			cpuPercent, _ := cmd.Flags().GetInt("cpu-percent")
			memorySize, _ := cmd.Flags().GetString("size")
			port, _ := cmd.Flags().GetInt("port")
			errorRate, _ := cmd.Flags().GetInt("error-rate")
			statusCode, _ := cmd.Flags().GetInt("status-code")

			expType := args[0]
			exp := &models.ChaosExperiment{
				ID:       fmt.Sprintf("exp-%d", time.Now().Unix()),
				Name:     fmt.Sprintf("%s experiment", expType),
				Duration: duration,
				Status:   models.ExperimentPending,
				Config: models.ChaosConfig{
					CPUPercent:  cpuPercent,
					Workers:     workers,
					MemoryBytes: memorySize,
					Port:        port,
					ErrorRate:   errorRate,
					StatusCode:  statusCode,
				},
			}

			switch expType {
			case "cpu":
				exp.Type = models.ChaosTypeCPUStress
			case "memory":
				exp.Type = models.ChaosTypeMemoryStress
			case "http-fault":
				exp.Type = models.ChaosTypeHTTPFault
			case "network-latency":
				exp.Type = models.ChaosTypeNetworkLatency
			default:
				return fmt.Errorf("unknown experiment type: %s", expType)
			}

			fmt.Println()
			fmt.Printf("   %s Starting %s experiment...\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⚡"),
				expType)
			fmt.Printf("   Duration: %s\n", duration)
			fmt.Println()

			ctx := context.Background()
			results, err := usecase.RunExperiment(ctx, exp)
			if err != nil {
				return err
			}

			displayChaosResults(results)
			return nil
		},
	}
	runCmd.Flags().String("duration", "30s", "Duracao do experimento")
	runCmd.Flags().Int("workers", 4, "Numero de workers (CPU)")
	runCmd.Flags().Int("cpu-percent", 80, "Percentual de CPU")
	runCmd.Flags().String("size", "100MB", "Tamanho da memoria")
	runCmd.Flags().Int("port", 8888, "Porta para fault injection")
	runCmd.Flags().Int("error-rate", 50, "Taxa de erro (%)")
	runCmd.Flags().Int("status-code", 500, "Codigo HTTP de erro")

	// Scenario subcommand
	scenarioCmd := &cobra.Command{
		Use:   "scenario <file>",
		Short: "Executa cenario de arquivo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scenario, err := usecase.LoadScenario(args[0])
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Running scenario: %s\n",
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⚡"),
				scenario.Name)
			fmt.Printf("   %s\n", scenario.Description)
			fmt.Println()

			// Verify steady state before
			if len(scenario.Steady) > 0 {
				fmt.Println(lipgloss.NewStyle().Bold(true).Render("   Verifying Steady State..."))
				for _, check := range scenario.Steady {
					ok, msg := usecase.VerifySteadyState(context.Background(), check)
					if ok {
						fmt.Printf("   %s %s: %s\n",
							lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
							check.Name, msg)
					} else {
						fmt.Printf("   %s %s: %s\n",
							lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
							check.Name, msg)
					}
				}
				fmt.Println()
			}

			// Run experiments
			for i, exp := range scenario.Experiments {
				fmt.Printf("   [%d/%d] Running: %s\n", i+1, len(scenario.Experiments), exp.Name)

				ctx := context.Background()
				results, err := usecase.RunExperiment(ctx, &exp)
				if err != nil {
					fmt.Printf("   %s Error: %s\n",
						lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
						err.Error())
					continue
				}

				if results.Success {
					fmt.Printf("   %s Completed\n",
						lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
				} else {
					fmt.Printf("   %s %s\n",
						lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
						results.Summary)
				}
			}

			// Verify steady state after
			if len(scenario.Steady) > 0 {
				fmt.Println()
				fmt.Println(lipgloss.NewStyle().Bold(true).Render("   Verifying Steady State (post)..."))
				for _, check := range scenario.Steady {
					ok, msg := usecase.VerifySteadyState(context.Background(), check)
					if ok {
						fmt.Printf("   %s %s: %s\n",
							lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
							check.Name, msg)
					} else {
						fmt.Printf("   %s %s: %s\n",
							lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
							check.Name, msg)
					}
				}
			}

			fmt.Println()
			return nil
		},
	}

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista experimentos em execucao",
		Run: func(cmd *cobra.Command, args []string) {
			running := usecase.ListRunning()

			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⚡ Running Experiments"))
			fmt.Println()

			if len(running) == 0 {
				fmt.Println("   Nenhum experimento em execucao")
				fmt.Println()
				return
			}

			for _, exp := range running {
				fmt.Printf("   %s [%s] %s (%s)\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("●"),
					exp.ID,
					exp.Name,
					exp.Type)
			}
			fmt.Println()
		},
	}

	// Stop subcommand
	stopCmd := &cobra.Command{
		Use:   "stop <id>",
		Short: "Para um experimento",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := usecase.StopExperiment(args[0]); err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("   %s Experiment %s stopped\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				args[0])
			return nil
		},
	}

	// Verify subcommand
	verifyCmd := &cobra.Command{
		Use:   "verify <target>",
		Short: "Verifica estado de um target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkType, _ := cmd.Flags().GetString("type")

			check := models.SteadyStateCheck{
				Name:   "Manual Check",
				Type:   checkType,
				Target: args[0],
			}

			ok, msg := usecase.VerifySteadyState(context.Background(), check)

			fmt.Println()
			if ok {
				fmt.Printf("   %s %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
					msg)
			} else {
				fmt.Printf("   %s %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
					msg)
			}
			fmt.Println()

			return nil
		},
	}
	verifyCmd.Flags().String("type", "http", "Tipo de verificacao (http, tcp)")

	// Init subcommand
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Gera cenario de exemplo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(usecase.GenerateScenario())
		},
	}

	cmd.AddCommand(runCmd, scenarioCmd, listCmd, stopCmd, verifyCmd, initCmd)

	return cmd
}

func displayChaosResults(results *models.ChaosResults) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("⚡ Experiment Results"))
	fmt.Println()

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	if !results.Success {
		successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   Success: %s\n", successStyle.Render(fmt.Sprintf("%v", results.Success)))
	fmt.Printf("   Summary: %s\n", results.Summary)
	fmt.Println(strings.Repeat("─", 60))

	if len(results.Metrics) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   Metrics:"))
		for k, v := range results.Metrics {
			fmt.Printf("      %s: %s\n", k, v)
		}
	}

	if len(results.Observations) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   Observations:"))
		for _, obs := range results.Observations {
			fmt.Printf("      • %s\n", obs)
		}
	}

	if len(results.Errors) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("   Errors:"))
		for _, err := range results.Errors {
			fmt.Printf("      • %s\n", err)
		}
	}

	fmt.Println()
}
