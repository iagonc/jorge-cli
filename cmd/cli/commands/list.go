package commands

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
)

// NewListCommand creates the "list" command
func NewListCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List all resources",
        Run: func(cmd *cobra.Command, args []string) {
            if err := listResources(); err != nil {
                utils.PrintError("listing resources", err)
            }
        },
    }
}

func listResources() error {
    url := fmt.Sprintf("%s/resources", utils.APIBaseURL)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }

    resp, err := utils.SendRequest(req)
    if err != nil {
        return fmt.Errorf("sending request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected server response: %s", resp.Status)
    }

    var apiResponse models.ApiResponse
    if err := utils.ParseResponse(resp, &apiResponse); err != nil {
        return fmt.Errorf("decoding response: %w", err)
    }

    // Styles with Lipgloss
    headerStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FAFAFA")).
        Background(lipgloss.Color("#7D56F4")).
        Padding(0, 1).
        Align(lipgloss.Left)

    rowStyle := lipgloss.NewStyle().
        Padding(0, 1).
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("#7D56F4"))

    tableHeader := headerStyle.Render(fmt.Sprintf("%-5s %-20s %-30s %-20s %-20s", "ID", "Name", "DNS", "CreatedAt", "UpdatedAt"))
    fmt.Println(tableHeader)

    for _, resource := range apiResponse.Data {
        createdAtFormatted := formatDate(resource.CreatedAt)
        updatedAtFormatted := formatDate(resource.UpdatedAt)

        resourceRow := fmt.Sprintf(
            "%-5d %-20s %-30s %-20s %-20s",
            resource.ID, resource.Name, resource.Dns, createdAtFormatted, updatedAtFormatted,
        )
        fmt.Println(rowStyle.Render(resourceRow))
    }

    return nil
}

// formatDate formats the date string into a human-readable format
func formatDate(dateStr string) string {
    parsedTime, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        return dateStr
    }
    return parsedTime.Format("2006-01-02 15:04")
}
