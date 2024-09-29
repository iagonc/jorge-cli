package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

// NewCreateCommand creates the "create" command with flags for name and DNS.
func NewCreateCommand() *cobra.Command {
    var name, dns string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new resource",
        Run: func(cmd *cobra.Command, args []string) {
            if name == "" || dns == "" {
                fmt.Fprintln(os.Stderr, "Error: --name and --dns are required")
                return
            }
            createResource(name, dns)
        },
    }

    // Add flags for "name" and "dns"
    cmd.Flags().StringVarP(&name, "name", "n", "", "Resource name (required)")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "Resource DNS (required)")
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("dns")

    return cmd
}

// createResource sends a request to create a resource with the provided name and DNS.
func createResource(name, dns string) {
    resource := CreateRequest{
        Name: name,
        Dns:  dns,
    }

    // Convert resource to JSON
    jsonData, err := json.Marshal(resource)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
        return
    }

    // Create HTTP request
    req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/resource", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating HTTP request: %v\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // Check response status code
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", resp.Status)
        return
    }

    // Decode the response
    var createResp CreateResponse
    if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
        fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
        return
    }

    // Success message styling
    successStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFD700")). // Gold color
        Bold(true)

    result := fmt.Sprintf(
        "Resource Created:\nID: %d\nName: %s\nDNS: %s",
        createResp.Data.ID, createResp.Data.Name, createResp.Data.Dns,
    )

    fmt.Println(successStyle.Render(result))
}
