package tunnel

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

// TunnelUsecase handles proxy/tunnel testing
type TunnelUsecase struct {
	logger *zap.Logger
}

// NewTunnelUsecase creates a new TunnelUsecase
func NewTunnelUsecase(logger *zap.Logger) *TunnelUsecase {
	return &TunnelUsecase{logger: logger}
}

// Test tests connectivity through a proxy
func (u *TunnelUsecase) Test(ctx context.Context, config models.TunnelConfig) (*models.TunnelResult, error) {
	startTime := time.Now()

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	result := &models.TunnelResult{
		Target:    config.Target,
		ProxyUsed: config.ProxyURL,
		ProxyType: config.ProxyType,
		Tests:     []models.TunnelTest{},
	}

	// Test direct connection first
	directResult := u.testDirect(ctx, config)
	result.Tests = append(result.Tests, directResult)

	// Test proxy connection if configured
	if config.ProxyURL != "" {
		proxyResult := u.testProxy(ctx, config)
		result.Tests = append(result.Tests, proxyResult)
	}

	// Generate summary
	u.generateSummary(result)

	result.Duration = time.Since(startTime)

	return result, nil
}

func (u *TunnelUsecase) testDirect(ctx context.Context, config models.TunnelConfig) models.TunnelTest {
	test := models.TunnelTest{
		Name: "Direct Connection",
	}

	host, port := parseTarget(config.Target)
	address := fmt.Sprintf("%s:%d", host, port)

	start := time.Now()
	dialer := &net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	test.Duration = time.Since(start)
	test.DirectLatency = test.Duration

	if err != nil {
		test.Error = err.Error()
		test.Details = "Failed to connect directly"
		return test
	}
	defer conn.Close()

	test.Success = true
	test.Details = fmt.Sprintf("Connected in %v", test.Duration.Round(time.Millisecond))

	return test
}

func (u *TunnelUsecase) testProxy(ctx context.Context, config models.TunnelConfig) models.TunnelTest {
	test := models.TunnelTest{
		Name: fmt.Sprintf("Proxy Connection (%s)", config.ProxyType),
	}

	host, port := parseTarget(config.Target)
	address := fmt.Sprintf("%s:%d", host, port)

	start := time.Now()

	var conn net.Conn
	var err error

	switch config.ProxyType {
	case models.ProxyTypeHTTP, models.ProxyTypeHTTPS:
		conn, err = u.dialHTTPProxy(ctx, config.ProxyURL, address, config.Timeout)
	case models.ProxyTypeSOCKS5:
		conn, err = u.dialSOCKS5Proxy(ctx, config.ProxyURL, address, config.Timeout)
	default:
		err = fmt.Errorf("unsupported proxy type: %s", config.ProxyType)
	}

	test.Duration = time.Since(start)
	test.ProxyLatency = test.Duration

	if err != nil {
		test.Error = err.Error()
		test.Details = "Failed to connect through proxy"
		return test
	}
	if conn != nil {
		defer conn.Close()
	}

	test.Success = true
	test.Details = fmt.Sprintf("Connected in %v", test.Duration.Round(time.Millisecond))

	return test
}

func (u *TunnelUsecase) dialHTTPProxy(ctx context.Context, proxyURL, target string, timeout time.Duration) (net.Conn, error) {
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Create transport with proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxy),
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	// Test with HTTP request
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// Make a simple request to test the proxy
	testURL := fmt.Sprintf("http://%s", target)
	req, err := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return nil, nil // Connection is handled by http.Client
}

func (u *TunnelUsecase) dialSOCKS5Proxy(ctx context.Context, proxyURL, target string, timeout time.Duration) (net.Conn, error) {
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	var auth *proxy.Auth
	if parsedProxy.User != nil {
		password, _ := parsedProxy.User.Password()
		auth = &proxy.Auth{
			User:     parsedProxy.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", parsedProxy.Host, auth, &net.Dialer{Timeout: timeout})
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (u *TunnelUsecase) generateSummary(result *models.TunnelResult) {
	summary := models.TunnelSummary{}

	for _, test := range result.Tests {
		if test.Name == "Direct Connection" {
			summary.DirectWorking = test.Success
		} else {
			summary.ProxyWorking = test.Success
		}
	}

	// Compare latencies
	if len(result.Tests) >= 2 {
		direct := result.Tests[0]
		proxied := result.Tests[1]

		if direct.Success && proxied.Success {
			diff := proxied.Duration - direct.Duration
			summary.LatencyDiff = diff
			summary.ProxyFaster = diff < 0
		}
	}

	// Generate recommendation
	if summary.DirectWorking && !summary.ProxyWorking {
		summary.Recommendation = "Conexão direta funciona, mas proxy não. Verifique configuração do proxy."
	} else if !summary.DirectWorking && summary.ProxyWorking {
		summary.Recommendation = "Proxy necessário para acessar o destino. Firewall pode estar bloqueando acesso direto."
	} else if summary.DirectWorking && summary.ProxyWorking {
		if summary.ProxyFaster {
			summary.Recommendation = "Ambos funcionam. Proxy é mais rápido (considere usar)."
		} else if summary.LatencyDiff > 100*time.Millisecond {
			summary.Recommendation = "Ambos funcionam. Conexão direta é significativamente mais rápida."
		} else {
			summary.Recommendation = "Ambos funcionam com latência similar."
		}
	} else {
		summary.Recommendation = "Nenhuma conexão funcionou. Verifique rede e firewall."
	}

	result.Summary = summary
}

func parseTarget(target string) (host string, port int) {
	// Default port
	port = 443

	// Check for URL scheme
	if len(target) > 8 && (target[:7] == "http://" || target[:8] == "https://") {
		parsed, err := url.Parse(target)
		if err == nil {
			host = parsed.Hostname()
			if p := parsed.Port(); p != "" {
				fmt.Sscanf(p, "%d", &port)
			} else if parsed.Scheme == "http" {
				port = 80
			}
			return
		}
	}

	// Check for host:port format
	h, p, err := net.SplitHostPort(target)
	if err == nil {
		host = h
		fmt.Sscanf(p, "%d", &port)
		return
	}

	// Just hostname
	host = target
	return
}
