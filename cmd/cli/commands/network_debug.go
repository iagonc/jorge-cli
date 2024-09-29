package commands

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/network"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
	"github.com/spf13/cobra"
)

func NewNetworkDebugCommand(usecase *network.NetworkDebugUsecase) *cobra.Command {
    var domain string

    cmd := &cobra.Command{
        Use:   "debug",
        Short: "Performs network diagnostics in a user-friendly manner",
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

            s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
            s.Suffix = " Running network diagnostics, it may take a few minutes..."
            s.Start()

            result, errorsList := usecase.NetworkDebug(ctx, domain)

            s.Stop()

            utils.FormatAndDisplayNetworkDebugResult(result, domain)

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
