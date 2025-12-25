package diagnose

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// DoctorUsecase handles local network health checks
type DoctorUsecase struct {
	logger *zap.Logger
}

// NewDoctorUsecase creates a new DoctorUsecase
func NewDoctorUsecase(logger *zap.Logger) *DoctorUsecase {
	return &DoctorUsecase{
		logger: logger,
	}
}

// RunDoctor performs local network health checks
func (u *DoctorUsecase) RunDoctor(ctx context.Context) (*models.DoctorResult, error) {
	startTime := time.Now()

	u.logger.Info("Starting local network doctor")

	result := &models.DoctorResult{
		Timestamp:   startTime,
		Problems:    []models.Problem{},
		Suggestions: []models.Suggestion{},
	}

	// Check local connectivity
	result.Connectivity = u.checkLocalConnectivity()

	// Check DNS
	result.DNS = u.checkDNS(ctx)

	// Check gateway
	result.Gateway = u.checkGateway(ctx)

	// Check internet
	result.Internet = u.checkInternet(ctx)

	// Analyze and generate problems/suggestions
	u.analyzeDoctor(result)

	result.Duration = time.Since(startTime)

	return result, nil
}

func (u *DoctorUsecase) checkLocalConnectivity() models.ConnectivityInfo {
	info := models.ConnectivityInfo{
		LocalIPs:   []string{},
		Interfaces: []string{},
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		info.Interfaces = append(info.Interfaces, iface.Name)

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if ip.To4() != nil {
				info.HasIPv4 = true
				info.LocalIPs = append(info.LocalIPs, ip.String())
			} else {
				info.HasIPv6 = true
			}
		}
	}

	return info
}

func (u *DoctorUsecase) checkDNS(ctx context.Context) models.DNSInfo {
	info := models.DNSInfo{
		Servers: []string{},
	}

	// Try to get DNS servers from system
	info.Servers = u.getSystemDNSServers()

	// Test DNS resolution
	start := time.Now()
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, address)
		},
	}

	_, err := resolver.LookupIP(ctx, "ip4", "google.com")
	info.ResponseTime = time.Since(start)

	info.CanResolve = err == nil

	return info
}

func (u *DoctorUsecase) getSystemDNSServers() []string {
	var servers []string

	// Try common DNS check methods
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Try reading resolv.conf
		cmd := exec.Command("cat", "/etc/resolv.conf")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "nameserver") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						servers = append(servers, parts[1])
					}
				}
			}
		}
	}

	// Fallback: test known public DNS
	if len(servers) == 0 {
		servers = []string{"8.8.8.8", "1.1.1.1"}
	}

	return servers
}

func (u *DoctorUsecase) checkGateway(ctx context.Context) models.GatewayInfo {
	info := models.GatewayInfo{}

	// Try to find default gateway
	info.IP = u.findDefaultGateway()

	if info.IP == "" {
		return info
	}

	// Ping the gateway
	start := time.Now()
	dialer := &net.Dialer{Timeout: 3 * time.Second}

	// Try to connect to common gateway port (usually doesn't work, but we try)
	// Better approach: try ICMP or just check if we can reach beyond
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(info.IP, "80"))
	if err == nil {
		conn.Close()
		info.Reachable = true
		info.ResponseTime = time.Since(start)
	} else {
		// Try UDP instead (DNS port)
		conn, err = dialer.DialContext(ctx, "udp", net.JoinHostPort(info.IP, "53"))
		if err == nil {
			conn.Close()
			info.Reachable = true
			info.ResponseTime = time.Since(start)
		}
	}

	// If still not reachable, try ping command
	if !info.Reachable {
		info.Reachable = u.pingHost(info.IP)
		if info.Reachable {
			info.ResponseTime = time.Since(start)
		}
	}

	return info
}

func (u *DoctorUsecase) findDefaultGateway() string {
	var gateway string

	if runtime.GOOS == "linux" {
		cmd := exec.Command("ip", "route", "show", "default")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			for i, part := range parts {
				if part == "via" && i+1 < len(parts) {
					gateway = parts[i+1]
					break
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("route", "-n", "get", "default")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "gateway:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						gateway = parts[1]
					}
					break
				}
			}
		}
	}

	return gateway
}

func (u *DoctorUsecase) pingHost(host string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", host)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", host)
	}

	err := cmd.Run()
	return err == nil
}

func (u *DoctorUsecase) checkInternet(ctx context.Context) models.InternetInfo {
	info := models.InternetInfo{}

	// Test internet connectivity using HTTP
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	endpoints := []string{
		"https://www.google.com/generate_204",
		"https://www.cloudflare.com/cdn-cgi/trace",
		"https://connectivitycheck.gstatic.com/generate_204",
	}

	for _, endpoint := range endpoints {
		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		info.Connected = true
		info.ResponseTime = time.Since(start)

		// Try to get public IP from cloudflare
		if strings.Contains(endpoint, "cloudflare") {
			body, _ := io.ReadAll(resp.Body)
			lines := strings.Split(string(body), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "ip=") {
					info.PublicIP = strings.TrimPrefix(line, "ip=")
					break
				}
			}
		}

		break
	}

	return info
}

func (u *DoctorUsecase) analyzeDoctor(result *models.DoctorResult) {
	priority := 1

	// Check local connectivity
	if !result.Connectivity.HasIPv4 && !result.Connectivity.HasIPv6 {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityCritical,
			Category:    models.CategoryConnectivity,
			Title:       "Sem endereço IP",
			Description: "Nenhum endereço IP configurado. Verifique a conexão de rede.",
		})
		result.Suggestions = append(result.Suggestions, models.Suggestion{
			Priority:    priority,
			Title:       "Verificar cabo/WiFi",
			Description: "Verifique se o cabo de rede está conectado ou se o WiFi está ativado.",
		})
		priority++
	}

	// Check DNS
	if !result.DNS.CanResolve {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityCritical,
			Category:    models.CategoryDNS,
			Title:       "DNS não funciona",
			Description: "Não foi possível resolver nomes de domínio.",
		})
		result.Suggestions = append(result.Suggestions, models.Suggestion{
			Priority:    priority,
			Title:       "Configurar DNS alternativo",
			Description: "Tente usar DNS do Google (8.8.8.8) ou Cloudflare (1.1.1.1).",
			Command:     "echo 'nameserver 8.8.8.8' | sudo tee /etc/resolv.conf",
		})
		priority++
	} else if result.DNS.ResponseTime > 500*time.Millisecond {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityWarning,
			Category:    models.CategoryDNS,
			Title:       "DNS lento",
			Description: fmt.Sprintf("Resolução DNS demorou %v.", result.DNS.ResponseTime.Round(time.Millisecond)),
		})
		result.Suggestions = append(result.Suggestions, models.Suggestion{
			Priority:    priority,
			Title:       "Usar DNS mais rápido",
			Description: "Considere usar DNS do Cloudflare (1.1.1.1) que geralmente é mais rápido.",
		})
		priority++
	}

	// Check gateway
	if result.Gateway.IP != "" && !result.Gateway.Reachable {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityCritical,
			Category:    models.CategoryConnectivity,
			Title:       "Gateway inacessível",
			Description: fmt.Sprintf("Não foi possível alcançar o gateway (%s).", result.Gateway.IP),
		})
		result.Suggestions = append(result.Suggestions, models.Suggestion{
			Priority:    priority,
			Title:       "Verificar roteador",
			Description: "Verifique se o roteador está ligado e funcionando corretamente.",
		})
		priority++
	}

	// Check internet
	if !result.Internet.Connected {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityCritical,
			Category:    models.CategoryConnectivity,
			Title:       "Sem acesso à Internet",
			Description: "Não foi possível acessar a Internet.",
		})
		result.Suggestions = append(result.Suggestions, models.Suggestion{
			Priority:    priority,
			Title:       "Verificar conexão com provedor",
			Description: "A rede local parece funcionar, mas sem acesso à Internet. Contate seu provedor.",
		})
		priority++
	} else if result.Internet.ResponseTime > 1*time.Second {
		result.Problems = append(result.Problems, models.Problem{
			Severity:    models.SeverityWarning,
			Category:    models.CategoryLatency,
			Title:       "Internet lenta",
			Description: fmt.Sprintf("Conexão com a Internet está lenta (%v).", result.Internet.ResponseTime.Round(time.Millisecond)),
		})
	}

	// Set overall status
	criticalCount := 0
	warningCount := 0
	for _, p := range result.Problems {
		if p.Severity == models.SeverityCritical {
			criticalCount++
		} else if p.Severity == models.SeverityWarning {
			warningCount++
		}
	}

	if criticalCount > 0 {
		result.Overall = models.StatusError
	} else if warningCount > 0 {
		result.Overall = models.StatusUnhealthy
	} else {
		result.Overall = models.StatusHealthy
	}
}
