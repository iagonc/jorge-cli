package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/usecase"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewNetworkDebugCommand(usecase *usecase.NetworkDebugUsecase) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "network-debug",
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
			result, err := usecase.NetworkDebug(ctx, domain)
			if err != nil {
				usecase.Logger.Error("Error executing network debug", zap.Error(err))
				fmt.Printf("Error executing network debug: %v\n", err)
				return
			}

			// Format and display the result
			utils.FormatAndDisplayNetworkDebugResult(result)

			// Success message
			successStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981")). // Green
				Padding(0, 2)
			fmt.Println(successStyle.Render("🔧 Network debug executed successfully!"))
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to perform network diagnostics")
	cmd.MarkFlagRequired("domain")

	return cmd
}
