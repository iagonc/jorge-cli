package compare

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// CompareUsecase handles endpoint comparison
type CompareUsecase struct {
	logger *zap.Logger
}

// NewCompareUsecase creates a new CompareUsecase
func NewCompareUsecase(logger *zap.Logger) *CompareUsecase {
	return &CompareUsecase{logger: logger}
}

// Compare compares two endpoints
func (u *CompareUsecase) Compare(ctx context.Context, target1, target2 string) (*models.CompareResult, error) {
	startTime := time.Now()

	result := &models.CompareResult{
		Differences: []models.CompareDiff{},
	}

	// Test both endpoints concurrently
	ch1 := make(chan models.EndpointResult, 1)
	ch2 := make(chan models.EndpointResult, 1)

	go func() {
		ch1 <- u.testEndpoint(ctx, target1)
	}()
	go func() {
		ch2 <- u.testEndpoint(ctx, target2)
	}()

	result.Endpoint1 = <-ch1
	result.Endpoint2 = <-ch2

	// Compare results
	u.compareEndpoints(result)

	result.Duration = time.Since(startTime)

	return result, nil
}

func (u *CompareUsecase) testEndpoint(ctx context.Context, target string) models.EndpointResult {
	result := models.EndpointResult{
		Target:  target,
		Headers: make(map[string]string),
	}

	// Parse target
	host, port, isHTTPS := parseTarget(target)

	// DNS lookup
	dnsStart := time.Now()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	result.DNSTime = time.Since(dnsStart)

	if err != nil {
		result.Error = fmt.Sprintf("DNS error: %v", err)
		return result
	}

	if len(ips) > 0 {
		result.IP = ips[0].String()
	}

	// TCP connect
	address := fmt.Sprintf("%s:%d", host, port)
	connectStart := time.Now()
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	result.ConnectTime = time.Since(connectStart)

	if err != nil {
		result.Error = fmt.Sprintf("Connect error: %v", err)
		return result
	}
	conn.Close()

	result.Reachable = true

	// TLS if HTTPS
	if isHTTPS || port == 443 {
		tlsStart := time.Now()
		tlsDialer := &tls.Dialer{
			NetDialer: &net.Dialer{Timeout: 10 * time.Second},
			Config:    &tls.Config{ServerName: host},
		}
		tlsConn, err := tlsDialer.DialContext(ctx, "tcp", address)
		result.TLSTime = time.Since(tlsStart)

		if err != nil {
			result.Error = fmt.Sprintf("TLS error: %v", err)
		} else {
			state := tlsConn.(*tls.Conn).ConnectionState()
			result.TLSVersion = tlsVersionName(state.Version)
			tlsConn.Close()
		}
	}

	// HTTP request if applicable
	if isHTTPS || port == 80 || port == 443 || port == 8080 || port == 8443 {
		scheme := "http"
		if isHTTPS || port == 443 || port == 8443 {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s:%d", scheme, host, port)

		client := &http.Client{Timeout: 15 * time.Second}
		httpStart := time.Now()
		resp, err := client.Get(url)
		result.ResponseTime = time.Since(httpStart)

		if err == nil {
			result.StatusCode = resp.StatusCode
			for k, v := range resp.Header {
				result.Headers[k] = strings.Join(v, ", ")
			}
			resp.Body.Close()
		}
	}

	return result
}

func (u *CompareUsecase) compareEndpoints(result *models.CompareResult) {
	e1, e2 := result.Endpoint1, result.Endpoint2

	// Reachability
	if e1.Reachable != e2.Reachable {
		result.Differences = append(result.Differences, models.CompareDiff{
			Field:       "Reachability",
			Value1:      fmt.Sprintf("%v", e1.Reachable),
			Value2:      fmt.Sprintf("%v", e2.Reachable),
			Severity:    "critical",
			Description: "Os endpoints têm acessibilidade diferente",
		})
	}

	// Status code
	if e1.StatusCode != e2.StatusCode && e1.StatusCode > 0 && e2.StatusCode > 0 {
		severity := "warning"
		if (e1.StatusCode >= 400) != (e2.StatusCode >= 400) {
			severity = "critical"
		}
		result.Differences = append(result.Differences, models.CompareDiff{
			Field:       "HTTP Status",
			Value1:      fmt.Sprintf("%d", e1.StatusCode),
			Value2:      fmt.Sprintf("%d", e2.StatusCode),
			Severity:    severity,
			Description: "Os endpoints retornam status HTTP diferentes",
		})
	}

	// TLS version
	if e1.TLSVersion != e2.TLSVersion && e1.TLSVersion != "" && e2.TLSVersion != "" {
		severity := "info"
		if (e1.TLSVersion == "1.0" || e1.TLSVersion == "1.1") || (e2.TLSVersion == "1.0" || e2.TLSVersion == "1.1") {
			severity = "warning"
		}
		result.Differences = append(result.Differences, models.CompareDiff{
			Field:       "TLS Version",
			Value1:      e1.TLSVersion,
			Value2:      e2.TLSVersion,
			Severity:    severity,
			Description: "Os endpoints usam versões TLS diferentes",
		})
	}

	// Response time difference
	if e1.ResponseTime > 0 && e2.ResponseTime > 0 {
		diff := e1.ResponseTime - e2.ResponseTime
		if diff < 0 {
			diff = -diff
		}
		if diff > 500*time.Millisecond {
			faster := e1.Target
			if e2.ResponseTime < e1.ResponseTime {
				faster = e2.Target
			}
			result.Differences = append(result.Differences, models.CompareDiff{
				Field:       "Response Time",
				Value1:      e1.ResponseTime.Round(time.Millisecond).String(),
				Value2:      e2.ResponseTime.Round(time.Millisecond).String(),
				Severity:    "warning",
				Description: fmt.Sprintf("%s é mais rápido por %v", faster, diff.Round(time.Millisecond)),
			})
		}
	}

	// Generate summary
	if len(result.Differences) == 0 {
		result.Summary = "Os endpoints são equivalentes"
	} else {
		critical := 0
		for _, d := range result.Differences {
			if d.Severity == "critical" {
				critical++
			}
		}
		if critical > 0 {
			result.Summary = fmt.Sprintf("Encontradas %d diferença(s), %d crítica(s)", len(result.Differences), critical)
		} else {
			result.Summary = fmt.Sprintf("Encontradas %d diferença(s) menores", len(result.Differences))
		}
	}
}

func parseTarget(target string) (host string, port int, isHTTPS bool) {
	if strings.HasPrefix(target, "https://") {
		isHTTPS = true
		target = strings.TrimPrefix(target, "https://")
		port = 443
	} else if strings.HasPrefix(target, "http://") {
		target = strings.TrimPrefix(target, "http://")
		port = 80
	} else {
		port = 443 // Default to HTTPS
	}

	// Remove path
	if idx := strings.Index(target, "/"); idx != -1 {
		target = target[:idx]
	}

	// Check for port
	if idx := strings.LastIndex(target, ":"); idx != -1 {
		host = target[:idx]
		fmt.Sscanf(target[idx+1:], "%d", &port)
	} else {
		host = target
	}

	return
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}
