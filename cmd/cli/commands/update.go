package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
)

// NewUpdateCommand creates the "update" command with flags for id, name, and DNS
func NewUpdateCommand() *cobra.Command {
    var id string
    var name string
    var dns string

    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update an existing resource",
        Run: func(cmd *cobra.Command, args []string) {
            if err := updateResource(id, name, dns); err != nil {
                utils.PrintError("updating resource", err)
            }
        },
    }

    // Add flags for "id", "name", and "dns"
    cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
    cmd.Flags().StringVarP(&name, "name", "n", "", "New resource name")
    cmd.Flags().StringVarP(&dns, "dns", "d", "", "New resource DNS")
    cmd.MarkFlagRequired("id")

    return cmd
}

func updateResource(id string, name, dns string) error {
    // Validate ID
    _, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("--id must be a valid integer")
    }

    if name == "" && dns == "" {
        return fmt.Errorf("at least one of --name or --dns must be provided")
    }

    // Prepare the update request body
    resource := models.UpdateRequest{
        Name: name,
        Dns:  dns,
    }

    // Marshal the request body to JSON
    jsonData, err := json.Marshal(resource)
    if err != nil {
        return fmt.Errorf("marshalling JSON: %w", err)
    }

    // Create the HTTP request
    url := fmt.Sprintf("%s/resource?id=%s", utils.APIBaseURL, id)
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Send the HTTP request
    resp, err := utils.SendRequest(req)
    if err != nil {
        return fmt.Errorf("updating resource: %w", err)
    }

    if resp.StatusCode == http.StatusNotFound {
        fmt.Printf("Resource with ID %s not found\n", id)
        return nil
    } else if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected server response: %s", resp.Status)
    }

    // Decode the response
    var updateResp models.UpdateResponse
    if err := utils.ParseResponse(resp, &updateResp); err != nil {
        return fmt.Errorf("decoding response: %w", err)
    }

    // Style the success message
    successStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FFD700")). // Gold color
        Padding(1, 2).
        Align(lipgloss.Center)

    // Display the updated resource details
    result := fmt.Sprintf(
        "Resource Updated:\nID: %d\nName: %s\nDNS: %s",
        updateResp.Data.ID, updateResp.Data.Name, updateResp.Data.Dns,
    )

    fmt.Println(successStyle.Render(result))
    return nil
}
