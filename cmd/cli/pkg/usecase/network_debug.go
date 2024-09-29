package usecase

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
	"go.uber.org/zap"
)

type NetworkDebugUsecase struct {
	Logger *zap.Logger
}

func NewNetworkDebugUsecase(logger *zap.Logger) *NetworkDebugUsecase {
	return &NetworkDebugUsecase{
		Logger: logger,
	}
}

func (u *NetworkDebugUsecase) NetworkDebug(ctx context.Context, domain string) (*models.NetworkDebugResult, error) {
	result := &models.NetworkDebugResult{}

	// Execute all tools in parallel
	errChan := make(chan error, 7)
	resultChan := make(chan struct{}, 7)

	// DNS Lookup (dig)
	go func() {
		dns, err := runDig(domain)
		if err != nil {
			errChan <- fmt.Errorf("dig error: %w", err)
			return
		}
		result.DNSLookup = dns
		resultChan <- struct{}{}
	}()

	// NSLookup
	go func() {
		ns, err := runNSLookup(domain)
		if err != nil {
			errChan <- fmt.Errorf("nslookup error: %w", err)
			return
		}
		result.NSLookup = ns
		resultChan <- struct{}{}
	}()

	// Traceroute
	go func() {
		tr, err := runTraceroute(domain)
		if err != nil {
			errChan <- fmt.Errorf("traceroute error: %w", err)
			return
		}
		result.Traceroute = tr
		resultChan <- struct{}{}
	}()

	// HTTP Request (curl)
	go func() {
		curl, err := runCurl(domain)
		if err != nil {
			errChan <- fmt.Errorf("curl error: %w", err)
			return
		}
		result.HTTPRequest = curl
		resultChan <- struct{}{}
	}()

	// Ping
	go func() {
		ping, err := runPing(domain)
		if err != nil {
			errChan <- fmt.Errorf("ping error: %w", err)
			return
		}
		result.Ping = ping
		resultChan <- struct{}{}
	}()

	// Netstat
	go func() {
		netstat, err := runNetstat()
		if err != nil {
			errChan <- fmt.Errorf("netstat error: %w", err)
			return
		}
		result.Netstat = netstat
		resultChan <- struct{}{}
	}()

	// Iftop
	go func() {
		iftop, err := runIftop("eth0") // Assuming eth0
		if err != nil {
			errChan <- fmt.Errorf("iftop error: %w", err)
			return
		}
		result.Iftop = iftop
		resultChan <- struct{}{}
	}()

	// Wait for all goroutines to finish
	for i := 0; i < 7; i++ {
		select {
		case <-resultChan:
			// Received a result, continue waiting
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return result, nil
}

// Helper functions to execute network tools

func runDig(domain string) (models.DNSLookupResult, error) {
	cmd := exec.Command("dig", "+short", domain)
	output, err := cmd.Output()
	if err != nil {
		return models.DNSLookupResult{}, err
	}

	var ipv4, ipv6 string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			ipv6 = line
		} else {
			ipv4 = line
		}
	}

	return models.DNSLookupResult{
		IPv4: ipv4,
		IPv6: ipv6,
	}, nil
}

func runNSLookup(domain string) (models.NSLookupResult, error) {
	cmd := exec.Command("nslookup", domain)
	output, err := cmd.Output()
	if err != nil {
		return models.NSLookupResult{}, err
	}

	lines := strings.Split(string(output), "\n")
	var ip string
	for _, line := range lines {
		if strings.Contains(line, "Address:") && !strings.Contains(line, "#") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ip = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	return models.NSLookupResult{
		IP: ip,
	}, nil
}

func runTraceroute(domain string) (models.TracerouteResult, error) {
	cmd := exec.Command("traceroute", "-m", "5", domain)
	output, err := cmd.Output()
	if err != nil {
		return models.TracerouteResult{}, err
	}

	lines := strings.Split(string(output), "\n")
	var hops []models.TracerouteHop
	for _, line := range lines[1:] { // Skip the first line
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		hopNumber, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		address := parts[1]
		responseTime := parts[len(parts)-2] // Assuming the time is before the last "ms"

		hops = append(hops, models.TracerouteHop{
			HopNumber:    hopNumber,
			Address:      address,
			ResponseTime: responseTime + " ms",
		})
	}

	return models.TracerouteResult{
		Hops: hops,
	}, nil
}

func runCurl(domain string) (models.HTTPRequestResult, error) {
	start := time.Now()
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code} %{time_total} %{content_type}", domain)
	output, err := cmd.Output()
	if err != nil {
		return models.HTTPRequestResult{}, err
	}

	parts := strings.Fields(string(output))
	if len(parts) < 3 {
		return models.HTTPRequestResult{}, fmt.Errorf("unexpected curl output: %s", string(output))
	}

	status := parts[0]
	responseTime := fmt.Sprintf("%.0f ms", time.Since(start).Seconds()*1000)
	contentType := parts[2]

	return models.HTTPRequestResult{
		Status:       fmt.Sprintf("HTTP %s", status),
		ResponseTime: responseTime,
		ContentType:  contentType,
	}, nil
}

func runPing(domain string) (models.PingResult, error) {
	cmd := exec.Command("ping", "-c", "4", domain)
	output, err := cmd.Output()
	if err != nil {
		return models.PingResult{}, err
	}

	var sent, received int
	var lossPercent float64
	var avgLatency int

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "packets transmitted") {
			// Example: 4 packets transmitted, 4 received, 0% packet loss, time 3005ms
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				fmt.Sscanf(parts[0], "%d packets transmitted", &sent)
				fmt.Sscanf(parts[1], " %d received", &received)
				fmt.Sscanf(parts[2], " %f%% packet loss", &lossPercent)
			}
		}
		if strings.Contains(line, "rtt min/avg/max/mdev") {
			// Example: rtt min/avg/max/mdev = 10.123/15.456/20.789/2.345 ms
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				stats := strings.Split(strings.TrimSpace(parts[1]), "/")
				if len(stats) >= 2 {
					avg, err := strconv.ParseFloat(stats[1], 64)
					if err == nil {
						avgLatency = int(avg)
					}
				}
			}
		}
	}

	return models.PingResult{
		Sent:         sent,
		Received:     received,
		Lost:         sent - received,
		LossPercent:  lossPercent,
		AvgLatency:   avgLatency,
	}, nil
}

func runNetstat() (models.NetstatResult, error) {
	cmd := exec.Command("netstat", "-tunapl")
	output, err := cmd.Output()
	if err != nil {
		return models.NetstatResult{}, err
	}

	lines := strings.Split(string(output), "\n")
	var connections []models.NetstatConnection
	for _, line := range lines {
		if strings.HasPrefix(line, "tcp") || strings.HasPrefix(line, "udp") {
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				connections = append(connections, models.NetstatConnection{
					Protocol:      parts[0],
					LocalAddress:  parts[3],
					RemoteAddress: parts[4],
					Status:        parts[5],
				})
			}
		}
	}

	return models.NetstatResult{
		Connections: connections,
	}, nil
}

func runIftop(interfaceName string) (models.IftopResult, error) {
	// Note: iftop typically requires root privileges. Ensure the CLI has necessary permissions.
	// We'll run iftop in text mode for 5 seconds and capture the output.
	cmd := exec.Command("sudo", "iftop", "-t", "-s", "5", "-i", interfaceName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return models.IftopResult{}, err
	}

	lines := strings.Split(string(out.String()), "\n")
	var sending, receiving string
	var topConns []models.IftopConnection

	for _, line := range lines {
		if strings.Contains(line, "=>") || strings.Contains(line, "<=") {
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				topConns = append(topConns, models.IftopConnection{
					Source:        parts[0],
					Destination:   parts[2],
					SentKBps:      parts[4],
					ReceivedKBps: parts[5],
				})
			}
		}
		if strings.Contains(line, "Total send rate") {
			// Example: Total send rate: 120.00 KB/s
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				sending = strings.TrimSpace(strings.TrimSuffix(parts[1], " KB/s"))
			}
		}
		if strings.Contains(line, "Total receive rate") {
			// Example: Total receive rate: 250.00 KB/s
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				receiving = strings.TrimSpace(strings.TrimSuffix(parts[1], " KB/s"))
			}
		}
	}

	// Select Top 3 connections
	limit := 3
	if len(topConns) < 3 {
		limit = len(topConns)
	}
	topConns = topConns[:limit]

	return models.IftopResult{
		SendingKBps:    sending + " KB/s",
		ReceivingKBps:  receiving + " KB/s",
		TopConnections: topConns,
	}, nil
}
