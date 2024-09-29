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

type UpdateRequest struct {
	Name string `json:"name,omitempty"`
	Dns  string `json:"dns,omitempty"`
}

type UpdateResponse struct {
	Data    Resource `json:"data"`
	Message string   `json:"message"`
}

// NewUpdateCommand creates the "update" command with flags for id, name, and DNS
func NewUpdateCommand() *cobra.Command {
	var id string
	var name string
	var dns string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing resource",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" {
				fmt.Fprintln(os.Stderr, "Error: --id flag is required")
				cmd.Help()
				return
			}
			if name == "" && dns == "" {
				fmt.Fprintln(os.Stderr, "Error: at least one of --name or --dns must be provided")
				cmd.Help()
				return
			}
			updateResource(id, name, dns)
		},
	}

	// Add flags for "id", "name", and "dns"
	cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "New resource name")
	cmd.Flags().StringVarP(&dns, "dns", "d", "", "New resource DNS")

	// Mark "id" flag as required
	cmd.MarkFlagRequired("id")

	return cmd
}

func updateResource(id string, name, dns string) {
	// Prepare the update request body
	resource := UpdateRequest{
		Name: name,
		Dns:  dns,
	}

	// Marshal the request body to JSON
	jsonData, err := json.Marshal(resource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		return
	}

	// Create the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "http://localhost:8080/api/v1/resource?id="+id, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating resource: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "Resource with ID %s not found\n", id)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", resp.Status)
		return
	}

	// Decode the response
	var updateResp UpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	// Style the success message
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")). // Gold color
		Padding(1, 2).
		Align(lipgloss.Center)

	// Display the updated resource details
	result := fmt.Sprintf(
		"Resource Updated: \nID: %d\nName: %s\nDNS: %s",
		updateResp.Data.ID, updateResp.Data.Name, updateResp.Data.Dns,
	)

	fmt.Println(successStyle.Render(result))
}
