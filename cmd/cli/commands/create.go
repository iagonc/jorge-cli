package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/services"
)

func NewCreateCommand(service *services.ResourceService) *cobra.Command {
    var name, dns string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new resource",
        Run: func(cmd *cobra.Command, args []string) {
            ctx := cmd.Context()
            resource, err := service.CreateResource(ctx, name, dns)
            if err != nil {
                service.Logger.Error("Error creating resource", zap.Error(err))
                return
            }

            successStyle := lipgloss.NewStyle().
                Foreground(lipgloss.Color("#FFD700")). // Gold color
                Bold(true)

            result := successStyle.Render(
                fmt.Sprintf("Resource Created:\nID: %d\nName: %s\nDNS: %s",
                    resource.ID, resource.Name, resource.Dns),
            )

            fmt.Println(result)
        },
    }

    // Add flags for "name" and "dns"
    cmd.Flags().StringVarP(&name, "name", "n", "", "Resource name (required)")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "Resource DNS (required)")
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("dns")

    return cmd
}
