package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type DeleteResponse struct {
	Data    Resource `json:"data"`
	Message string   `json:"message"`
}

// NewDeleteCommand creates the "delete" command
func NewDeleteCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource by ID",
		Run: func(cmd *cobra.Command, args []string) {
			deleteResource(id)
		},
	}

	// Add flag for "id"
	cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func deleteResource(id string) {
	// Fetch the resource details by ID
	resource, err := fetchResourceByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching resource: %v\n", err)
		return
	}

	// Show resource details
	fmt.Printf("Resource Details:\nID: %d\nName: %s\nDNS: %s\n", resource.ID, resource.Name, resource.Dns)

	// Ask for confirmation
	fmt.Print("Are you sure you want to delete this resource? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
		return
	}
	confirmation = strings.TrimSpace(confirmation)

	if strings.ToLower(confirmation) != "yes" {
		fmt.Println("Delete operation canceled.")
		return
	}

	// Proceed with deletion if confirmed
	client := &http.Client{}
	baseURL := "http://localhost:8080/api/v1/resource"
	params := url.Values{}
	params.Add("id", id)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting resource: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Resource with ID %s not found\n", id)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", resp.Status)
		return
	}

	var deleteResp DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6347")). // Soft red color
		Bold(true)

	// Show details of deleted resource
	result := fmt.Sprintf(
		"Resource Deleted: \nID: %d\nName: %s\nDNS: %s",
		deleteResp.Data.ID, deleteResp.Data.Name, deleteResp.Data.Dns,
	)

	fmt.Println(successStyle.Render(result))
}

// fetchResourceByID fetches the resource details by ID before deletion
func fetchResourceByID(id string) (*Resource, error) {
	resp, err := http.Get("http://localhost:8080/api/v1/resource?id=" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %v", err)
	}
	defer resp.Body.Close()

	// Check if the resource was found
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("resource with ID %s not found", id)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected server response: %s", resp.Status)
	}

	var deleteResp DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return nil, fmt.Errorf("failed to decode resource: %v", err)
	}

	return &deleteResp.Data, nil
}
