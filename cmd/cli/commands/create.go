package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"

	"go.uber.org/zap"
)

func NewCreateCommand(usecase *resource.ResourceUsecase) *cobra.Command {
	var name, dns string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new resource",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			// Validate inputs
			if err := utils.ValidateCreateInputs(name, dns); err != nil {
				usecase.Logger.Error("Invalid input", zap.Error(err))
				fmt.Println(err)
				return
			}

			resource, err := usecase.CreateResource(ctx, name, dns)
			if err != nil {
				usecase.Logger.Error("Error creating resource", zap.Error(err))
				fmt.Println("Error creating resource:", err)
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

	cmd.Flags().StringVarP(&name, "name", "n", "", "Resource name (required)")
	cmd.Flags().StringVarP(&dns, "dns", "d", "", "Resource DNS (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("dns")

	return cmd
}
