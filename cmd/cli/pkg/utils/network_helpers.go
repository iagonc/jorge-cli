package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
)

// FormatAndDisplayNetworkDebugResult formats and displays the network debug results in a friendly manner
func FormatAndDisplayNetworkDebugResult(result *models.NetworkDebugResult) {
	// Define styles using Lipgloss
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	//subtitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
	listStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FFFFFF"))

	// DNS Lookup
	fmt.Println(titleStyle.Render("✨ DNS Verification (dig):"))
	if result.DNSLookup.IPv4 != "" || result.DNSLookup.IPv6 != "" {
		fmt.Println(fmt.Sprintf("- The domain %s is associated with the following addresses:", result.DNSLookup.IPv4))
		if result.DNSLookup.IPv4 != "" {
			fmt.Println(listStyle.Render(fmt.Sprintf("- IPv4: %s", result.DNSLookup.IPv4)))
		}
		if result.DNSLookup.IPv6 != "" {
			fmt.Println(listStyle.Render(fmt.Sprintf("- IPv6: %s", result.DNSLookup.IPv6)))
		}
	} else {
		fmt.Println("- No DNS records found.")
	}
	fmt.Println()

	// NSLookup
	fmt.Println(titleStyle.Render("🔍 Address Lookup (nslookup):"))
	if result.NSLookup.IP != "" {
		fmt.Printf("- The IP address of %s is %s\n", result.NSLookup.IP, result.NSLookup.IP)
	} else {
		fmt.Println("- No IP address found.")
	}
	fmt.Println()

	// Traceroute
	fmt.Println(titleStyle.Render("🚀 Data Route (Traceroute):"))
	if len(result.Traceroute.Hops) > 0 {
		lastHop := result.Traceroute.Hops[len(result.Traceroute.Hops)-1].Address
		fmt.Printf("- Data traveled through %d points before reaching %s:\n", len(result.Traceroute.Hops), lastHop)
		for _, hop := range result.Traceroute.Hops {
			fmt.Printf("  %d. %s: Response in %s\n", hop.HopNumber, hop.Address, hop.ResponseTime)
		}
	} else {
		fmt.Println("- No traceroute data available.")
	}
	fmt.Println()

	// HTTP Request (curl)
	fmt.Println(titleStyle.Render("📡 Site Verification (curl):"))
	if result.HTTPRequest.Status != "" {
		fmt.Printf("- Site Status: Working correctly (%s)\n", result.HTTPRequest.Status)
		fmt.Printf("- Response Time: %s\n", result.HTTPRequest.ResponseTime)
		fmt.Printf("- Content Type: Web page (%s)\n", result.HTTPRequest.ContentType)
	} else {
		fmt.Println("- No HTTP request data available.")
	}
	fmt.Println()

	// Ping
	fmt.Println(titleStyle.Render("📈 Connection Test (Ping):"))
	if result.Ping.Sent > 0 {
		fmt.Printf("- Packets Sent: %d\n", result.Ping.Sent)
		fmt.Printf("- Packets Received: %d\n", result.Ping.Received)
		fmt.Printf("- Packet Loss: %.0f%%\n", result.Ping.LossPercent)
		fmt.Printf("- Average Response Time: %d ms\n", result.Ping.AvgLatency)
	} else {
		fmt.Println("- No ping data available.")
	}
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
	if result.Iftop.SendingKBps != "" || result.Iftop.ReceivingKBps != "" {
		fmt.Println("- Current Traffic:")
		fmt.Printf("  - Sending: %s\n", result.Iftop.SendingKBps)
		fmt.Printf("  - Receiving: %s\n", result.Iftop.ReceivingKBps)
		fmt.Println("- Top 3 Most Active Connections:")
		for i, conn := range result.Iftop.TopConnections {
			fmt.Printf("  %d. %s ↔ %s: Sending %s | Receiving %s\n", i+1, conn.Source, conn.Destination, conn.SentKBps, conn.ReceivedKBps)
		}
	} else {
		fmt.Println("- No network usage data available.")
	}
}