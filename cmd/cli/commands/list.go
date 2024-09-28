package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type resource struct {
    ID        uint   `json:"ID"`
    Name      string `json:"name"`
    Dns       string `json:"dns"`
    CreatedAt string `json:"CreatedAt"`
    UpdatedAt string `json:"UpdatedAt"`
}

type ApiResponse struct {
    Data    []resource `json:"data"`
    Message string     `json:"message"`
}

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
    response, err := http.Get("http://localhost:8080/api/v1/resources")
    if err != nil {
        fmt.Println("Error fetching resources:", err)
        return
    }
    defer response.Body.Close()

    body, err := io.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading response:", err)
        return
    }

    var apiResponse ApiResponse
    json.Unmarshal(body, &apiResponse)

    // Estilos com Lipgloss
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

// Função para formatar as datas
func formatDate(dateStr string) string {
    parsedTime, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        return dateStr
    }
    return parsedTime.Format("2006-01-02 15:04")
}
