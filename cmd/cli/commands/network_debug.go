package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/usecase"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"
	"github.com/spf13/cobra"
)

func NewNetworkDebugCommand(usecase *usecase.NetworkDebugUsecase) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Performs network diagnostics in a friendly manner",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			// Check if all tools are installed
			tools := []string{"iftop", "dig", "nslookup", "traceroute", "curl", "ping", "netstat"}
			missingTools := []string{}
			for _, tool := range tools {
				if _, err := exec.LookPath(tool); err != nil {
					missingTools = append(missingTools, tool)
				}
			}

			if len(missingTools) > 0 {
				fmt.Printf("⚠️  The following tools are missing: %s\n", strings.Join(missingTools, ", "))
				fmt.Println("Please install them to use the network-debug command.")
				fmt.Println("Installation example on Ubuntu/Debian:")
				fmt.Printf("  sudo apt install %s\n", strings.Join(missingTools, " "))
				return
			}

			// Execute the network debug usecase
			result, errorsList := usecase.NetworkDebug(ctx, domain)

			// Format and display the result
			utils.FormatAndDisplayNetworkDebugResult(result)

			// Display errors, if any
			if len(errorsList) > 0 {
				errorStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FF6347")) // Soft red color
				fmt.Println(errorStyle.Render("⚠️  Some tools encountered errors:"))
				for _, err := range errorsList {
					fmt.Printf("- %v\n", err)
				}
			}

			// Success message if no errors
			if len(errorsList) == 0 {
				successStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#10B981")). // Green
					Padding(0, 2)
				fmt.Println(successStyle.Render("🔧 Network diagnostics executed successfully!"))
			}
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to perform network diagnostics")
	cmd.MarkFlagRequired("domain")

	return cmd
}
