package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPClient defines an interface for sending HTTP requests.
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient creates a new HTTP client with the specified timeout.
func NewHTTPClient(timeout time.Duration) HTTPClient {
    return &http.Client{
        Timeout: timeout,
    }
}

// ParseErrorResponse parses the API error response and returns a formatted error.
func ParseErrorResponse(resp *http.Response) error {
    var errResp struct {
        Error   string `json:"error"`
        Message string `json:"message"`
    }
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
    }
    if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
        return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(bodyBytes))
    }
    return fmt.Errorf("API Error: %s - %s", errResp.Error, errResp.Message)
}

// ConfirmAction prompts the user for confirmation.
func ConfirmAction(prompt string) bool {
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print(prompt)
        input, err := reader.ReadString('\n')
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
            return false
        }
        input = strings.TrimSpace(strings.ToLower(input))
        switch input {
        case "yes", "y":
            return true
        case "no", "n":
            return false
        default:
            fmt.Println("Invalid input. Please type 'yes' or 'no'.")
        }
    }
}
