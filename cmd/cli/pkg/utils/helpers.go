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

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
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

// FormatAndDisplayNetworkDebugResult formats and displays the network debug results in a friendly manner
func FormatAndDisplayNetworkDebugResult(result *models.NetworkDebugResult) {
	// Define styles using Lipgloss
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	listStyle := lipgloss.NewStyle().PaddingLeft(2)

	// DNS Lookup
	fmt.Println(titleStyle.Render("✨ DNS Verification (dig):"))
	fmt.Printf("- The domain %s is associated with the following addresses:\n", result.DNSLookup.IPv4)
	fmt.Println(listStyle.Render(fmt.Sprintf("- IPv4: %s", result.DNSLookup.IPv4)))
	fmt.Println(listStyle.Render(fmt.Sprintf("- IPv6: %s", result.DNSLookup.IPv6)))
	fmt.Println()

	// NSLookup
	fmt.Println(titleStyle.Render("🔍 Address Lookup (nslookup):"))
	fmt.Printf("- The IP address of %s is %s\n", result.NSLookup.IP, result.NSLookup.IP)
	fmt.Println()

	// Traceroute
	fmt.Println(titleStyle.Render("🚀 Data Route (Traceroute):"))
	fmt.Printf("- Data traveled through %d points before reaching %s:\n", len(result.Traceroute.Hops), result.Traceroute.Hops[len(result.Traceroute.Hops)-1].Address)
	for _, hop := range result.Traceroute.Hops {
		fmt.Printf("  %d. %s: Response in %s\n", hop.HopNumber, hop.Address, hop.ResponseTime)
	}
	fmt.Println()

	// HTTP Request (curl)
	fmt.Println(titleStyle.Render("📡 Site Verification (curl):"))
	fmt.Printf("- Site Status: Working correctly (%s)\n", result.HTTPRequest.Status)
	fmt.Printf("- Response Time: %s\n", result.HTTPRequest.ResponseTime)
	fmt.Printf("- Content Type: Web page (%s)\n", result.HTTPRequest.ContentType)
	fmt.Println()

	// Ping
	fmt.Println(titleStyle.Render("📈 Connection Test (Ping):"))
	fmt.Printf("- Packets Sent: %d\n", result.Ping.Sent)
	fmt.Printf("- Packets Received: %d\n", result.Ping.Received)
	fmt.Printf("- Packet Loss: %.0f%%\n", result.Ping.LossPercent)
	fmt.Printf("- Average Response Time: %d ms\n", result.Ping.AvgLatency)
	fmt.Println()

	// Netstat
	fmt.Println(titleStyle.Render("🖥️ Active Connections (Netstat):"))
	if len(result.Netstat.Connections) == 0 {
		fmt.Println("- No active connections found.")
	} else {
		fmt.Println("- Active Connections:")
		for _, conn := range result.Netstat.Connections {
			fmt.Printf("  - %s %s → %s (%s)\n", conn.Protocol, conn.LocalAddress, conn.RemoteAddress, conn.Status)
		}
	}
	fmt.Println()

	// Iftop
	fmt.Println(titleStyle.Render("📊 Current Network Usage (Iftop - Interface: eth0):"))
	fmt.Printf("- Current Traffic:\n")
	fmt.Printf("  - Sending: %s\n", result.Iftop.SendingKBps)
	fmt.Printf("  - Receiving: %s\n", result.Iftop.ReceivingKBps)
	fmt.Println("- Top 3 Most Active Connections:")
	for i, conn := range result.Iftop.TopConnections {
		fmt.Printf("  %d. %s ↔ %s: Sending %s | Receiving %s\n", i+1, conn.Source, conn.Destination, conn.SentKBps, conn.ReceivedKBps)
	}
}
