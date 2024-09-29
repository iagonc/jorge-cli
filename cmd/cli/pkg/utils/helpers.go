package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// SendRequest sends an HTTP request and returns the response.
func SendRequest(req *http.Request) (*http.Response, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    return client.Do(req)
}

// ParseResponse parses the JSON response into the provided interface.
func ParseResponse(resp *http.Response, result interface{}) error {
    defer resp.Body.Close()
    decoder := json.NewDecoder(resp.Body)
    return decoder.Decode(result)
}

// PrintError prints an error message to stderr.
func PrintError(action string, err error) {
    fmt.Fprintf(os.Stderr, "Error %s: %v\n", action, err)
}

// ConfirmAction prompts the user for confirmation.
func ConfirmAction(prompt string) bool {
    fmt.Print(prompt)
    reader := bufio.NewReader(os.Stdin)
    confirmation, err := reader.ReadString('\n')
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
        return false
    }
    confirmation = strings.TrimSpace(strings.ToLower(confirmation))
    return confirmation == "yes" || confirmation == "y"
}
