package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
)

// NewCreateCommand creates the "create" command with flags for name and DNS.
func NewCreateCommand() *cobra.Command {
    var name, dns string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new resource",
        Run: func(cmd *cobra.Command, args []string) {
            if err := createResource(name, dns); err != nil {
                utils.PrintError("creating resource", err)
            }
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
func createResource(name, dns string) error {
    resource := models.CreateRequest{
        Name: name,
        Dns:  dns,
    }

    // Convert resource to JSON
    jsonData, err := json.Marshal(resource)
    if err != nil {
        return fmt.Errorf("marshalling JSON: %w", err)
    }

    // Create HTTP request
    url := fmt.Sprintf("%s/resource", utils.APIBaseURL)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("creating HTTP request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Send the request
    resp, err := utils.SendRequest(req)
    if err != nil {
        return fmt.Errorf("sending request: %w", err)
    }

    // Check response status code
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected server response: %s", resp.Status)
    }

    // Decode the response
    var createResp models.CreateResponse
    if err := utils.ParseResponse(resp, &createResp); err != nil {
        return fmt.Errorf("decoding response: %w", err)
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
    return nil
}
