package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type UpdateRequest struct {
    Name string `json:"name"`
    Dns  string `json:"dns"`
}

func NewUpdateCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "update [id]",
        Short: "Update an existing resource",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            updateResource(args[0])
        },
    }
}

func updateResource(id string) {
    resource := UpdateRequest{
        Name: "Updated Resource",
        Dns:  "updated-resource.com",
    }

    jsonData, _ := json.Marshal(resource)

    client := &http.Client{}
    req, err := http.NewRequest("PUT", "http://localhost:8080/api/v1/resource?id="+id, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error updating resource:", err)
        return
    }

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error updating resource:", err)
        return
    }
    defer resp.Body.Close()

    successStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FFD700")).
        Background(lipgloss.Color("#FAFAFA")).
        Padding(1, 2).
        Align(lipgloss.Center)

    fmt.Println(successStyle.Render("Resource updated successfully!"))
}
