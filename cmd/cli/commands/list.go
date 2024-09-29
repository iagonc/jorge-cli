package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Resource represents a resource with metadata
type ResourceWithTimestamps struct {
    ID        uint   `json:"ID"`
    Name      string `json:"name"`
    Dns       string `json:"dns"`
    CreatedAt string `json:"CreatedAt"`
    UpdatedAt string `json:"UpdatedAt"`
}

// ApiResponse represents the API response structure
type ApiResponse struct {
	Data    []ResourceWithTimestamps `json:"data"`
	Message string     `json:"message"`
}

// NewListCommand creates the "list" command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all resources",
		Run: func(cmd *cobra.Command, args []string) {
			listResources()
		},
	}
}

func listResources() {
	resp, err := http.Get("http://localhost:8080/api/v1/resources")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching resources: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", resp.Status)
		return
	}

	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
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
}

// formatDate formats the date string into a human-readable format
func formatDate(dateStr string) string {
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return parsedTime.Format("2006-01-02 15:04")
}
