package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/playbook"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewPlaybookCommand(usecase *playbook.PlaybookUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playbook",
		Short: "Manage and run troubleshooting playbooks",
		Long: `The playbook command allows you to create, validate, and run
custom troubleshooting playbooks defined in YAML format.

Playbooks are discovered from:
  - ./playbooks/
  - ~/.jorge/playbooks/
  - /etc/jorge/playbooks/`,
	}

	cmd.AddCommand(NewPlaybookRunCommand(usecase))
	cmd.AddCommand(NewPlaybookListCommand(usecase))
	cmd.AddCommand(NewPlaybookValidateCommand(usecase))

	return cmd
}

func NewPlaybookRunCommand(usecase *playbook.PlaybookUsecase) *cobra.Command {
	var (
		playbookPath string
		variables    []string
		dryRun       bool
		verbose      bool
		skipConfirm  bool
		startFrom    string
		stopAt       string
	)

	cmd := &cobra.Command{
		Use:   "run [playbook-name]",
		Short: "Execute a playbook",
		Example: `  # Run playbook by name
  jorge-cli playbook run network-debug --var domain=example.com

  # Run playbook from file
  jorge-cli playbook run --file ./my-playbook.yaml --var domain=test.com

  # Dry run to preview steps
  jorge-cli playbook run network-debug --dry-run --var domain=example.com

  # Skip confirmation prompts
  jorge-cli playbook run service-restart --var service=nginx --yes`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Loading playbook..."
			s.Start()

			// Load playbook
			var pb *models.Playbook
			var err error

			if playbookPath != "" {
				pb, err = usecase.LoadPlaybook(ctx, playbookPath)
			} else if len(args) == 1 {
				pb, err = usecase.LoadPlaybookByName(ctx, args[0])
			} else {
				s.Stop()
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: Specify playbook name or --file path"))
				return
			}

			s.Stop()

			if err != nil {
				usecase.Logger.Error("Failed to load playbook", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Parse variables
			vars := parseVariables(variables)

			// Check requirements
			missing, _ := usecase.CheckRequirements(pb)
			if len(missing) > 0 {
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Missing required tools: " + strings.Join(missing, ", ")))
				fmt.Println("Please install them and try again.")
				return
			}

			// Run options
			opts := playbook.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				SkipConfirm: skipConfirm,
				StartFrom:   startFrom,
				StopAt:      stopAt,
			}

			// Show playbook info
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
			fmt.Println()
			fmt.Println(titleStyle.Render("Running Playbook: " + pb.Name))
			if pb.Description != "" {
				fmt.Println("  " + pb.Description)
			}
			if dryRun {
				warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
				fmt.Println(warningStyle.Render("  [DRY RUN MODE]"))
			}
			fmt.Println()

			// Execute
			result, err := usecase.RunPlaybook(ctx, pb, vars, opts)
			if err != nil {
				usecase.Logger.Error("Playbook execution failed", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Display results
			displayPlaybookResult(result, verbose)
		},
	}

	cmd.Flags().StringVarP(&playbookPath, "file", "f", "", "Path to playbook file")
	cmd.Flags().StringArrayVar(&variables, "var", nil, "Set variable (key=value)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview steps without executing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "Skip confirmation prompts")
	cmd.Flags().StringVar(&startFrom, "start-from", "", "Start from specific step ID")
	cmd.Flags().StringVar(&stopAt, "stop-at", "", "Stop at specific step ID")

	return cmd
}

func NewPlaybookListCommand(usecase *playbook.PlaybookUsecase) *cobra.Command {
	var tags []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available playbooks",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			playbooks, err := usecase.DiscoverPlaybooks(ctx)
			if err != nil {
				usecase.Logger.Error("Failed to discover playbooks", zap.Error(err))
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Error: " + err.Error()))
				return
			}

			// Filter by tags if specified
			if len(tags) > 0 {
				playbooks = filterByTags(playbooks, tags)
			}

			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
			fmt.Println()
			fmt.Println(titleStyle.Render("Available Playbooks"))
			fmt.Println(titleStyle.Render("==================="))
			fmt.Println()

			if len(playbooks) == 0 {
				fmt.Println("No playbooks found.")
				fmt.Println()
				fmt.Println("Playbooks are searched in:")
				fmt.Println("  - ./playbooks/")
				fmt.Println("  - ~/.jorge/playbooks/")
				fmt.Println("  - /etc/jorge/playbooks/")
				return
			}

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

			fmt.Println(headerStyle.Render(fmt.Sprintf("%-25s %-10s %-40s %-20s", "NAME", "VERSION", "DESCRIPTION", "TAGS")))

			for _, pb := range playbooks {
				desc := pb.Description
				if len(desc) > 39 {
					desc = desc[:36] + "..."
				}
				tagsStr := strings.Join(pb.Tags, ", ")
				if len(tagsStr) > 19 {
					tagsStr = tagsStr[:16] + "..."
				}
				fmt.Printf("  %-25s %-10s %-40s %-20s\n", pb.Name, pb.Version, desc, tagsStr)
			}
			fmt.Println()
		},
	}

	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tags")

	return cmd
}

func NewPlaybookValidateCommand(usecase *playbook.PlaybookUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <playbook-file>",
		Short: "Validate a playbook file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			pb, err := usecase.LoadPlaybook(ctx, args[0])
			if err != nil {
				errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
				fmt.Println(errorStyle.Render("Validation failed: " + err.Error()))
				return
			}

			successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
			fmt.Println(successStyle.Render("Playbook is valid!"))
			fmt.Println()
			fmt.Printf("  Name: %s\n", pb.Name)
			fmt.Printf("  Version: %s\n", pb.Version)
			fmt.Printf("  Steps: %d\n", len(pb.Steps))
			if len(pb.Finally) > 0 {
				fmt.Printf("  Finally steps: %d\n", len(pb.Finally))
			}
			if len(pb.Requirements.Tools) > 0 {
				fmt.Printf("  Required tools: %s\n", strings.Join(pb.Requirements.Tools, ", "))
			}
		},
	}

	return cmd
}

func parseVariables(vars []string) map[string]string {
	result := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func filterByTags(playbooks []models.PlaybookInfo, tags []string) []models.PlaybookInfo {
	var filtered []models.PlaybookInfo
	for _, pb := range playbooks {
		for _, tag := range tags {
			for _, pbTag := range pb.Tags {
				if strings.EqualFold(pbTag, tag) {
					filtered = append(filtered, pb)
					break
				}
			}
		}
	}
	return filtered
}

func displayPlaybookResult(result *models.PlaybookResult, verbose bool) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(titleStyle.Render("Playbook Execution Summary"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("  %s %s\n", labelStyle.Render("Playbook:"), result.Playbook)
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), result.Duration.Round(time.Millisecond))
	fmt.Println()

	// Count statuses
	var success, failed, skipped int
	for _, sr := range result.StepResults {
		switch sr.Status {
		case models.StatusSuccess:
			success++
		case models.StatusFailed:
			failed++
		case models.StatusSkipped:
			skipped++
		}
	}

	fmt.Println(titleStyle.Render("Steps:"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Passed:"), successStyle.Render(fmt.Sprintf("%d", success)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Failed:"), errorStyle.Render(fmt.Sprintf("%d", failed)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Skipped:"), warningStyle.Render(fmt.Sprintf("%d", skipped)))
	fmt.Println()

	// Show step details
	if verbose || failed > 0 {
		fmt.Println(titleStyle.Render("Step Details:"))
		for _, sr := range result.StepResults {
			var statusStr string
			switch sr.Status {
			case models.StatusSuccess:
				statusStr = successStyle.Render("[SUCCESS]")
			case models.StatusFailed:
				statusStr = errorStyle.Render("[FAILED]")
			case models.StatusSkipped:
				statusStr = warningStyle.Render("[SKIPPED]")
			}

			fmt.Printf("  %s %s (%s)\n", statusStr, sr.StepName, sr.Duration.Round(time.Millisecond))

			if sr.Error != "" {
				fmt.Printf("    %s\n", errorStyle.Render("Error: "+sr.Error))
			}
			if sr.SkipReason != "" {
				fmt.Printf("    %s\n", warningStyle.Render("Reason: "+sr.SkipReason))
			}
			if verbose && sr.Stdout != "" {
				fmt.Printf("    Output: %s\n", strings.TrimSpace(sr.Stdout))
			}
		}
		fmt.Println()
	}

	// Final status
	if result.Status == models.StatusSuccess {
		fmt.Println(successStyle.Render("Playbook completed successfully!"))
	} else {
		fmt.Println(errorStyle.Render("Playbook failed."))
		if result.Error != "" {
			fmt.Println(errorStyle.Render("Error: " + result.Error))
		}
	}
	fmt.Println()
}
