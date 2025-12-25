package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/internal/schemas"
)

// NetworkTools provides real network diagnostic functions
type NetworkTools struct{}

// NewNetworkTools creates a new NetworkTools instance
func NewNetworkTools() *NetworkTools {
	return &NetworkTools{}
}

// Ping executes a real ping command
func (t *NetworkTools) Ping(target string, timeout int) *schemas.DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Use ping command with count of 3
	cmd := exec.CommandContext(ctx, "ping", "-c", "3", "-W", strconv.Itoa(timeout), target)
	output, err := cmd.CombinedOutput()

	latency := time.Since(start).Milliseconds()
	result := &schemas.DiagnosticResult{
		Type:    "ping",
		Target:  target,
		Latency: latency,
		Output:  string(output),
		Details: make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	// Parse average latency from ping output
	avgLatency := t.parsePingLatency(string(output))
	if avgLatency > 0 {
		result.Latency = avgLatency
		result.Details["avg_latency_ms"] = avgLatency
	}

	// Parse packet loss
	packetLoss := t.parsePacketLoss(string(output))
	result.Details["packet_loss_percent"] = packetLoss

	result.Success = packetLoss < 100
	return result
}

// DNS performs DNS lookup using nslookup or dig
func (t *NetworkTools) DNS(target string, timeout int) *schemas.DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Try nslookup first, fall back to host command
	cmd := exec.CommandContext(ctx, "nslookup", target)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Try with host command as fallback
		cmd = exec.CommandContext(ctx, "host", target)
		output, err = cmd.CombinedOutput()
	}

	latency := time.Since(start).Milliseconds()
	result := &schemas.DiagnosticResult{
		Type:    "dns",
		Target:  target,
		Latency: latency,
		Output:  string(output),
		Details: make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	// Parse resolved IPs
	ips := t.parseResolvedIPs(string(output))
	result.Details["resolved_ips"] = ips
	result.Success = len(ips) > 0

	return result
}

// TCP checks if a port is open using netcat or direct connection
func (t *NetworkTools) TCP(target string, port int, timeout int) *schemas.DiagnosticResult {
	start := time.Now()
	address := fmt.Sprintf("%s:%d", target, port)

	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)

	latency := time.Since(start).Milliseconds()
	result := &schemas.DiagnosticResult{
		Type:    "tcp",
		Target:  address,
		Latency: latency,
		Details: make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = fmt.Sprintf("Connection to %s failed: %s", address, err.Error())
		return result
	}

	conn.Close()
	result.Success = true
	result.Output = fmt.Sprintf("Connection to %s succeeded", address)
	result.Details["port"] = port
	result.Details["status"] = "open"

	return result
}

// HTTP performs an HTTP request and checks the response
func (t *NetworkTools) HTTP(target string, method string, expectedStatus int, timeout int) *schemas.DiagnosticResult {
	start := time.Now()

	// Ensure URL has scheme
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return &schemas.DiagnosticResult{
			Type:    "http",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}
	}

	req.Header.Set("User-Agent", "Jorge-Monitor/1.0")

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	result := &schemas.DiagnosticResult{
		Type:    "http",
		Target:  target,
		Latency: latency,
		Details: make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = fmt.Sprintf("HTTP request failed: %s", err.Error())
		return result
	}
	defer resp.Body.Close()

	result.Details["status_code"] = resp.StatusCode
	result.Details["status"] = resp.Status
	result.Details["content_length"] = resp.ContentLength
	result.Output = fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)

	// Check if status matches expected
	if expectedStatus > 0 {
		result.Success = resp.StatusCode == expectedStatus
	} else {
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
	}

	return result
}

// SSL checks SSL certificate information
func (t *NetworkTools) SSL(target string, port int, timeout int) *schemas.DiagnosticResult {
	start := time.Now()

	if port == 0 {
		port = 443
	}

	address := fmt.Sprintf("%s:%d", target, port)

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: time.Duration(timeout) * time.Second},
		"tcp",
		address,
		&tls.Config{
			InsecureSkipVerify: false,
			ServerName:         target,
		},
	)

	latency := time.Since(start).Milliseconds()
	result := &schemas.DiagnosticResult{
		Type:    "ssl",
		Target:  target,
		Latency: latency,
		Details: make(map[string]interface{}),
	}

	if err != nil {
		// Try with InsecureSkipVerify to still get cert info
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: time.Duration(timeout) * time.Second},
			"tcp",
			address,
			&tls.Config{
				InsecureSkipVerify: true,
				ServerName:         target,
			},
		)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result
		}
		result.Details["cert_valid"] = false
	} else {
		result.Details["cert_valid"] = true
	}

	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) > 0 {
		cert := certs[0]
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

		result.Details["issuer"] = cert.Issuer.CommonName
		result.Details["subject"] = cert.Subject.CommonName
		result.Details["not_before"] = cert.NotBefore.Format(time.RFC3339)
		result.Details["not_after"] = cert.NotAfter.Format(time.RFC3339)
		result.Details["days_left"] = daysLeft
		result.Details["dns_names"] = cert.DNSNames

		result.Output = fmt.Sprintf("Certificate valid for %d days (expires %s)",
			daysLeft, cert.NotAfter.Format("2006-01-02"))

		result.Success = daysLeft > 0
		if daysLeft <= 30 {
			result.Details["warning"] = "Certificate expires within 30 days"
		}
	}

	return result
}

// Traceroute performs a traceroute to the target
func (t *NetworkTools) Traceroute(target string, timeout int) *schemas.DiagnosticResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Use traceroute command with max 15 hops
	cmd := exec.CommandContext(ctx, "traceroute", "-m", "15", "-w", "2", target)
	output, err := cmd.CombinedOutput()

	latency := time.Since(start).Milliseconds()
	result := &schemas.DiagnosticResult{
		Type:    "trace",
		Target:  target,
		Latency: latency,
		Output:  string(output),
		Details: make(map[string]interface{}),
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		result.Error = "traceroute timed out"
		result.Success = false
		return result
	}

	// Parse hops from output
	hops := t.parseTracerouteHops(string(output))
	result.Details["hops"] = hops
	result.Details["hop_count"] = len(hops)
	result.Success = len(hops) > 0

	return result
}

// Helper functions

func (t *NetworkTools) parsePingLatency(output string) int64 {
	// Parse "rtt min/avg/max/mdev = 10.123/15.456/20.789/3.456 ms"
	re := regexp.MustCompile(`rtt min/avg/max/mdev = [\d.]+/([\d.]+)/`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if avg, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return int64(avg)
		}
	}
	return 0
}

func (t *NetworkTools) parsePacketLoss(output string) float64 {
	// Parse "3 packets transmitted, 3 received, 0% packet loss"
	re := regexp.MustCompile(`(\d+)% packet loss`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if loss, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return loss
		}
	}
	return 100
}

func (t *NetworkTools) parseResolvedIPs(output string) []string {
	var ips []string
	// Parse IP addresses from nslookup/host output
	re := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	matches := re.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) > 0 {
			ip := match[1]
			// Skip common DNS server IPs
			if !strings.HasPrefix(ip, "127.") {
				ips = append(ips, ip)
			}
		}
	}
	return ips
}

func (t *NetworkTools) parseTracerouteHops(output string) []map[string]interface{} {
	var hops []map[string]interface{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "traceroute") {
			continue
		}

		hop := make(map[string]interface{})
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hop["hop"] = parts[0]
			if parts[1] == "*" {
				hop["host"] = "*"
				hop["timeout"] = true
			} else {
				hop["host"] = parts[1]
				// Try to parse latency
				for _, p := range parts {
					if strings.HasSuffix(p, "ms") {
						hop["latency"] = strings.TrimSuffix(p, "ms")
						break
					}
				}
			}
			hops = append(hops, hop)
		}
	}

	return hops
}
