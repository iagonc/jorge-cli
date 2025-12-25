package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// IPUsecase handles IP information lookups
type IPUsecase struct {
	logger *zap.Logger
}

// NewIPUsecase creates a new IPUsecase
func NewIPUsecase(logger *zap.Logger) *IPUsecase {
	return &IPUsecase{logger: logger}
}

// Lookup gets information about an IP address
func (u *IPUsecase) Lookup(ctx context.Context, ipStr string) (*models.IPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// Try to resolve as hostname
		ips, err := net.LookupIP(ipStr)
		if err != nil {
			return nil, fmt.Errorf("IP ou hostname inválido: %s", ipStr)
		}
		if len(ips) > 0 {
			ip = ips[0]
			ipStr = ip.String()
		}
	}

	info := &models.IPInfo{
		IP:         ipStr,
		IsPrivate:  ip.IsPrivate(),
		IsLoopback: ip.IsLoopback(),
	}

	// Determine version
	if ip.To4() != nil {
		info.Version = "IPv4"
	} else {
		info.Version = "IPv6"
	}

	var wg sync.WaitGroup

	// Reverse DNS lookup
	wg.Add(1)
	go func() {
		defer wg.Done()
		names, err := net.LookupAddr(ipStr)
		if err == nil && len(names) > 0 {
			info.Hostname = strings.TrimSuffix(names[0], ".")
			info.DNS = &models.IPDNSInfo{PTR: names}
		}
	}()

	// Geo lookup (using free API)
	if !ip.IsPrivate() && !ip.IsLoopback() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			info.Geo = u.lookupGeo(ctx, ipStr)
		}()

		// ASN lookup
		wg.Add(1)
		go func() {
			defer wg.Done()
			info.ASN = u.lookupASN(ctx, ipStr)
		}()
	}

	wg.Wait()

	return info, nil
}

// LookupWithPorts also scans common ports
func (u *IPUsecase) LookupWithPorts(ctx context.Context, ipStr string) (*models.IPInfo, error) {
	info, err := u.Lookup(ctx, ipStr)
	if err != nil {
		return nil, err
	}

	// Scan common ports
	ports := []struct {
		port    int
		service string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{443, "HTTPS"},
		{21, "FTP"},
		{25, "SMTP"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{8080, "HTTP-Alt"},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range ports {
		wg.Add(1)
		go func(port int, service string) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", ipStr, port)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)

			mu.Lock()
			info.Ports = append(info.Ports, models.PortInfo{
				Port:    port,
				Open:    err == nil,
				Service: service,
			})
			mu.Unlock()

			if err == nil {
				conn.Close()
			}
		}(p.port, p.service)
	}

	wg.Wait()

	return info, nil
}

func (u *IPUsecase) lookupGeo(ctx context.Context, ip string) *models.GeoInfo {
	// Using ip-api.com (free, no API key needed)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	if data.Status != "success" {
		return nil
	}

	return &models.GeoInfo{
		Country:     data.Country,
		CountryCode: data.CountryCode,
		Region:      data.Region,
		City:        data.City,
		Latitude:    data.Lat,
		Longitude:   data.Lon,
		Timezone:    data.Timezone,
	}
}

func (u *IPUsecase) lookupASN(ctx context.Context, ip string) *models.ASNInfo {
	// Using ip-api.com response from geo lookup
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=as,isp,org", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		ISP string `json:"isp"`
		Org string `json:"org"`
		AS  string `json:"as"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	// Parse AS number
	var asn int
	var asnOrg string
	if data.AS != "" {
		fmt.Sscanf(data.AS, "AS%d %s", &asn, &asnOrg)
		if asnOrg == "" {
			asnOrg = data.Org
		}
	}

	return &models.ASNInfo{
		Number:       asn,
		Organization: asnOrg,
		ISP:          data.ISP,
	}
}

// MyIP returns the public IP of the current machine
func (u *IPUsecase) MyIP(ctx context.Context) (string, error) {
	urls := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var ip strings.Builder
		buf := make([]byte, 64)
		n, _ := resp.Body.Read(buf)
		ip.Write(buf[:n])

		return strings.TrimSpace(ip.String()), nil
	}

	return "", fmt.Errorf("não foi possível obter IP público")
}
