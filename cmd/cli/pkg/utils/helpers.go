package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient creates a new HTTP client with the specified timeout.
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// ParseErrorResponse parses the API error response and returns a formatted error.
func ParseErrorResponse(resp *http.Response) error {
	contentType := resp.Header.Get("Content-Type")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if strings.Contains(contentType, "application/json") {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
			return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(bodyBytes))
		}
		return fmt.Errorf("API Error: %s - %s", errResp.Error, errResp.Message)
	}

	// Handle non-JSON error responses
	return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(bodyBytes))
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

// ParseID parses a string ID to an integer.
func ParseID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID '%s': %w", idStr, err)
	}
	return id, nil
}

// FormatDate formats the date string into a human-readable format.
func FormatDate(dateStr string) string {
	parsedTime, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		return dateStr
	}
	return parsedTime.Format("2006-01-02 15:04")
}

func HandleSignals(cancel context.CancelFunc, logger *zap.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	cancel()
}

// ValidateCreateInputs validates the inputs for the create command
func ValidateCreateInputs(name, dns string) error {
	if len(name) < 3 {
		return fmt.Errorf("name must be at least 3 characters long")
	}
	if len(dns) < 3 {
		return fmt.Errorf("dns must be at least 3 characters long")
	}
	return nil
}
