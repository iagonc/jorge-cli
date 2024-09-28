package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type CreateRequest struct {
    Name string `json:"name"`
    Dns  string `json:"dns"`
}

type CreateResponse struct {
    Data    Resource `json:"data"`
    Message string   `json:"message"`
}

// NewCreateCommand creates the "create" command with flags for name and dns
func NewCreateCommand() *cobra.Command {
    var name string
    var dns string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new resource",
        Run: func(cmd *cobra.Command, args []string) {
            createResource(name, dns)
        },
    }

    // Add flags for "name" and "dns"
    cmd.Flags().StringVarP(&name, "name", "n", "", "Resource name (required)")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "Resource DNS (required)")

    // Mark flags as required
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("dns")

    return cmd
}

func createResource(name, dns string) {
    resource := CreateRequest{
        Name: name,
        Dns:  dns,
    }

    jsonData, _ := json.Marshal(resource)

    resp, err := http.Post("http://localhost:8080/api/v1/resource", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error creating resource:", err)
        return
    }
    defer resp.Body.Close()

    var createResp CreateResponse
    json.NewDecoder(resp.Body).Decode(&createResp)

    successStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00FF00")). // Soft green color
        Bold(true)

    // Display created resource details
    result := fmt.Sprintf(
        "Resource Created: \nID: %d\nName: %s\nDNS: %s",
        createResp.Data.ID, createResp.Data.Name, createResp.Data.Dns,
    )

    fmt.Println(successStyle.Render(result))
}
