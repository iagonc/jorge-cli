package commands

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
)

// NewDeleteCommand creates the "delete" command
func NewDeleteCommand() *cobra.Command {
    var id string

    cmd := &cobra.Command{
        Use:   "delete",
        Short: "Delete a resource by ID",
        Run: func(cmd *cobra.Command, args []string) {
            if err := deleteResource(id); err != nil {
                utils.PrintError("deleting resource", err)
            }
        },
    }

    // Add flag for "id"
    cmd.Flags().StringVarP(&id, "id", "i", "", "Resource ID (required)")
    cmd.MarkFlagRequired("id")

    return cmd
}

func deleteResource(id string) error {
    // Validate ID
    _, err := strconv.Atoi(id)
    if err != nil {
        return fmt.Errorf("--id must be a valid integer")
    }

    // Fetch the resource details by ID
    resource, err := fetchResourceByID(id)
    if err != nil {
        return fmt.Errorf("fetching resource: %w", err)
    }

    // Show resource details
    fmt.Printf("Resource Details:\nID: %d\nName: %s\nDNS: %s\n", resource.ID, resource.Name, resource.Dns)

    // Ask for confirmation
    if !utils.ConfirmAction("Are you sure you want to delete this resource? (yes/no): ") {
        fmt.Println("Delete operation canceled.")
        return nil
    }

    // Proceed with deletion if confirmed
    baseURL := fmt.Sprintf("%s/resource", utils.APIBaseURL)
    params := url.Values{}
    params.Add("id", id)
    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    req, err := http.NewRequest("DELETE", fullURL, nil)
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }

    resp, err := utils.SendRequest(req)
    if err != nil {
        return fmt.Errorf("deleting resource: %w", err)
    }

    if resp.StatusCode == http.StatusNotFound {
        fmt.Printf("Resource with ID %s not found\n", id)
        return nil
    } else if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected server response: %s", resp.Status)
    }

    var deleteResp models.DeleteResponse
    if err := utils.ParseResponse(resp, &deleteResp); err != nil {
        return fmt.Errorf("decoding response: %w", err)
    }

    successStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF6347")). // Soft red color
        Bold(true)

    // Show details of deleted resource
    result := fmt.Sprintf(
        "Resource Deleted:\nID: %d\nName: %s\nDNS: %s",
        deleteResp.Data.ID, deleteResp.Data.Name, deleteResp.Data.Dns,
    )

    fmt.Println(successStyle.Render(result))
    return nil
}

// fetchResourceByID fetches the resource details by ID before deletion
func fetchResourceByID(id string) (*models.Resource, error) {
    url := fmt.Sprintf("%s/resource?id=%s", utils.APIBaseURL, id)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    resp, err := utils.SendRequest(req)
    if err != nil {
        return nil, fmt.Errorf("sending request: %w", err)
    }

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("resource with ID %s not found", id)
    } else if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected server response: %s", resp.Status)
    }

    var getResp models.DeleteResponse
    if err := utils.ParseResponse(resp, &getResp); err != nil {
        return nil, fmt.Errorf("decoding response: %w", err)
    }

    return &getResp.Data, nil
}
