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

// NewCreateCommand creates the "create" command with flags for name and DNS
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

	jsonData, err := json.Marshal(resource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/api/v1/resource", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating resource: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", resp.Status)
		return
	}

	var createResp CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

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
