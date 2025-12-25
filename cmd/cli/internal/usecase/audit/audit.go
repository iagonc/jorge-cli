package audit

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// AuditUsecase handles full system audits
type AuditUsecase struct {
	logger *zap.Logger
}

// NewAuditUsecase creates a new AuditUsecase
func NewAuditUsecase(logger *zap.Logger) *AuditUsecase {
	return &AuditUsecase{logger: logger}
}

// RunAudit performs a complete audit of a target
func (u *AuditUsecase) RunAudit(ctx context.Context, config models.AuditConfig, progressCb func(phase string, progress int)) (*models.AuditReport, error) {
	startTime := time.Now()

	report := &models.AuditReport{
		Target:          config.Target,
		Timestamp:       startTime,
		Issues:          []models.AuditIssue{},
		Warnings:        []models.AuditWarning{},
		Recommendations: []string{},
	}

	// Parse target URL
	targetURL := config.Target
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	host := parsed.Hostname()
	isHTTPS := parsed.Scheme == "https"

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Phase 1: Connectivity (20%)
	progressCb("Testando conectividade...", 0)
	report.Connectivity = u.checkConnectivity(ctx, host, targetURL, config.Timeout)

	if !report.Connectivity.DNSResolves {
		report.OverallStatus = models.AuditStatusDown
		report.Issues = append(report.Issues, models.AuditIssue{
			Severity:    "critical",
			Category:    "DNS",
			Title:       "DNS não resolve",
			Description: "O domínio não pôde ser resolvido para um endereço IP",
			Solution:    "Verifique se o domínio está correto e se o DNS está configurado",
		})
		report.Duration = time.Since(startTime)
		u.calculateScores(report)
		return report, nil
	}

	progressCb("Conectividade OK", 20)

	// Run remaining checks in parallel
	wg.Add(3)

	// Phase 2: SSL/TLS (40%)
	go func() {
		defer wg.Done()
		if isHTTPS {
			progressCb("Analisando SSL/TLS...", 25)
			ssl := u.checkSSL(ctx, host, config.Timeout)
			mu.Lock()
			report.SSL = ssl
			mu.Unlock()
		}
	}()

	// Phase 3: Security Headers (60%)
	go func() {
		defer wg.Done()
		progressCb("Verificando headers de segurança...", 40)
		headers := u.checkSecurityHeaders(ctx, targetURL, config.Timeout)
		mu.Lock()
		report.SecurityHeaders = headers
		mu.Unlock()
	}()

	// Phase 4: Performance (80%)
	go func() {
		defer wg.Done()
		progressCb("Medindo performance...", 55)
		perf := u.checkPerformance(ctx, targetURL, config.Timeout)
		mu.Lock()
		report.Performance = perf
		mu.Unlock()
	}()

	wg.Wait()

	// Phase 5: Network Path (optional)
	if !config.SkipMTR {
		progressCb("Analisando rota de rede...", 70)
		report.NetworkPath = u.checkNetworkPath(ctx, host)
	}

	// Phase 6: Port Scan (optional)
	if !config.SkipPorts {
		progressCb("Escaneando portas...", 85)
		report.PortScan = u.checkPorts(ctx, host, config.Timeout)
	}

	progressCb("Gerando relatório...", 95)

	// Calculate scores and generate summary
	u.calculateScores(report)
	u.generateIssuesAndWarnings(report)
	u.generateRecommendations(report)

	report.Duration = time.Since(startTime)

	progressCb("Concluído!", 100)

	return report, nil
}

func (u *AuditUsecase) checkConnectivity(ctx context.Context, host, targetURL string, timeout time.Duration) *models.AuditConnectivity {
	result := &models.AuditConnectivity{}

	// DNS Resolution
	dnsStart := time.Now()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	result.DNSTime = time.Since(dnsStart)

	if err != nil {
		return result
	}

	result.DNSResolves = true
	for _, ip := range ips {
		result.IPs = append(result.IPs, ip.IP.String())
	}

	// TCP Connection
	port := "443"
	if strings.HasPrefix(targetURL, "http://") {
		port = "80"
	}

	tcpStart := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	result.TCPTime = time.Since(tcpStart)

	if err == nil {
		result.TCPConnects = true
		conn.Close()
	}

	// HTTP Request
	client := &http.Client{Timeout: timeout}
	httpStart := time.Now()
	resp, err := client.Get(targetURL)
	result.HTTPTime = time.Since(httpStart)

	if err == nil {
		result.HTTPWorks = true
		result.HTTPStatus = resp.StatusCode
		resp.Body.Close()
	}

	// Calculate score
	result.Score = 0
	if result.DNSResolves {
		result.Score += 33
	}
	if result.TCPConnects {
		result.Score += 33
	}
	if result.HTTPWorks {
		result.Score += 34
	}

	return result
}

func (u *AuditUsecase) checkSSL(ctx context.Context, host string, timeout time.Duration) *models.AuditSSL {
	result := &models.AuditSSL{
		Errors: []string{},
		Chain:  []string{},
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", host+":443", &tls.Config{
		InsecureSkipVerify: false,
	})

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		// Try with insecure to get cert info anyway
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", host+":443", &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return result
		}
	}
	defer conn.Close()

	state := conn.ConnectionState()

	// TLS Version
	switch state.Version {
	case tls.VersionTLS13:
		result.Version = "TLS 1.3"
		result.Score += 25
	case tls.VersionTLS12:
		result.Version = "TLS 1.2"
		result.Score += 20
	case tls.VersionTLS11:
		result.Version = "TLS 1.1"
		result.Score += 10
	case tls.VersionTLS10:
		result.Version = "TLS 1.0"
		result.Score += 5
	}

	result.CipherSuite = tls.CipherSuiteName(state.CipherSuite)

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		result.Certificate = cert.Subject.CommonName
		result.Issuer = cert.Issuer.CommonName
		result.ExpiresAt = cert.NotAfter
		result.DaysUntilExpiry = int(time.Until(cert.NotAfter).Hours() / 24)

		// Check validity
		now := time.Now()
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			result.Valid = false
			result.Errors = append(result.Errors, "Certificado expirado ou não válido ainda")
		} else {
			result.Valid = true
			result.Score += 25
		}

		// Check expiry
		if result.DaysUntilExpiry < 30 {
			result.Score += 10
		} else if result.DaysUntilExpiry < 90 {
			result.Score += 20
		} else {
			result.Score += 25
		}

		// Certificate chain
		for _, c := range state.PeerCertificates {
			result.Chain = append(result.Chain, c.Subject.CommonName)
		}

		// Strong cipher
		if strings.Contains(result.CipherSuite, "AES") || strings.Contains(result.CipherSuite, "CHACHA") {
			result.Score += 25
		} else {
			result.Score += 10
		}
	}

	return result
}

func (u *AuditUsecase) checkSecurityHeaders(ctx context.Context, targetURL string, timeout time.Duration) *models.AuditSecurityHeaders {
	result := &models.AuditSecurityHeaders{
		Headers: []models.AuditHeaderCheck{},
		Missing: []string{},
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	// Headers to check
	headerChecks := []struct {
		name        string
		headerName  string
		description string
		weight      int
	}{
		{"HSTS", "Strict-Transport-Security", "Força conexões HTTPS", 15},
		{"CSP", "Content-Security-Policy", "Política de segurança de conteúdo", 20},
		{"X-Content-Type-Options", "X-Content-Type-Options", "Previne MIME sniffing", 10},
		{"X-Frame-Options", "X-Frame-Options", "Proteção contra clickjacking", 10},
		{"X-XSS-Protection", "X-Xss-Protection", "Proteção XSS (legacy)", 5},
		{"Referrer-Policy", "Referrer-Policy", "Controla informações de referência", 10},
		{"Permissions-Policy", "Permissions-Policy", "Controla APIs do navegador", 10},
		{"CORS", "Access-Control-Allow-Origin", "Política CORS", 5},
		{"COOP", "Cross-Origin-Opener-Policy", "Isolamento de contexto", 5},
		{"CORP", "Cross-Origin-Resource-Policy", "Política de recursos", 5},
	}

	totalWeight := 0
	earnedWeight := 0

	for _, check := range headerChecks {
		totalWeight += check.weight
		value := resp.Header.Get(check.headerName)
		present := value != ""

		headerCheck := models.AuditHeaderCheck{
			Name:        check.name,
			Present:     present,
			Value:       value,
			Description: check.description,
			Secure:      present,
		}

		if present {
			earnedWeight += check.weight
		} else {
			result.Missing = append(result.Missing, check.name)
		}

		result.Headers = append(result.Headers, headerCheck)
	}

	result.Score = (earnedWeight * 100) / totalWeight

	// Grade
	switch {
	case result.Score >= 90:
		result.Grade = "A+"
	case result.Score >= 80:
		result.Grade = "A"
	case result.Score >= 70:
		result.Grade = "B"
	case result.Score >= 60:
		result.Grade = "C"
	case result.Score >= 50:
		result.Grade = "D"
	default:
		result.Grade = "F"
	}

	return result
}

func (u *AuditUsecase) checkPerformance(ctx context.Context, targetURL string, timeout time.Duration) *models.AuditPerformance {
	result := &models.AuditPerformance{}

	var dnsStart, dnsDone time.Time
	var connectStart, connectDone time.Time
	var tlsStart, tlsDone time.Time
	var firstByte time.Time

	trace := &httptrace.ClientTrace{
		DNSStart:          func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:           func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart:      func(_, _ string) { connectStart = time.Now() },
		ConnectDone:       func(_, _ string, _ error) { connectDone = time.Now() },
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone:  func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { firstByte = time.Now() },
	}

	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "GET", targetURL, nil)
	client := &http.Client{Timeout: timeout}

	startTime := time.Now()
	resp, err := client.Do(req)

	if err != nil {
		return result
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	totalTime := time.Since(startTime)

	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		result.DNSLookup = dnsDone.Sub(dnsStart)
	}
	if !connectStart.IsZero() && !connectDone.IsZero() {
		result.TCPConnect = connectDone.Sub(connectStart)
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		result.TLSHandshake = tlsDone.Sub(tlsStart)
	}
	if !firstByte.IsZero() {
		if !tlsDone.IsZero() {
			result.ServerProcessing = firstByte.Sub(tlsDone)
		} else if !connectDone.IsZero() {
			result.ServerProcessing = firstByte.Sub(connectDone)
		}
		result.ContentTransfer = totalTime - time.Since(startTime) + time.Since(firstByte)
	}

	result.TotalTime = totalTime
	result.TTFB = result.DNSLookup + result.TCPConnect + result.TLSHandshake + result.ServerProcessing

	// Rating
	switch {
	case result.TotalTime < 500*time.Millisecond:
		result.Rating = "Excelente"
		result.Score = 100
	case result.TotalTime < 1*time.Second:
		result.Rating = "Bom"
		result.Score = 80
	case result.TotalTime < 2*time.Second:
		result.Rating = "Normal"
		result.Score = 60
	case result.TotalTime < 5*time.Second:
		result.Rating = "Lento"
		result.Score = 40
	default:
		result.Rating = "Muito Lento"
		result.Score = 20
	}

	return result
}

func (u *AuditUsecase) checkNetworkPath(ctx context.Context, host string) *models.AuditNetworkPath {
	result := &models.AuditNetworkPath{
		Hops: []models.AuditHop{},
	}

	// Simplified MTR-like check (just ping each hop we can find)
	for ttl := 1; ttl <= 30; ttl++ {
		conn, err := net.DialTimeout("ip4:icmp", host, 2*time.Second)
		if err != nil {
			// Try to resolve and ping
			ips, err := net.LookupIP(host)
			if err != nil || len(ips) == 0 {
				break
			}

			start := time.Now()
			testConn, err := net.DialTimeout("tcp", net.JoinHostPort(ips[0].String(), "443"), 2*time.Second)
			latency := time.Since(start)

			if err == nil {
				testConn.Close()
				result.Hops = append(result.Hops, models.AuditHop{
					Number:  ttl,
					Host:    host,
					Latency: latency,
					Loss:    0,
				})
				result.Completed = true
				result.TotalHops = ttl
				result.TotalLatency = latency
				break
			}
			break
		}
		conn.Close()
	}

	// Simple score based on completion
	if result.Completed {
		result.Score = 100
		if result.TotalLatency > 200*time.Millisecond {
			result.Score = 70
		}
	} else {
		result.Score = 50
	}

	return result
}

func (u *AuditUsecase) checkPorts(ctx context.Context, host string, timeout time.Duration) *models.AuditPortScan {
	result := &models.AuditPortScan{
		OpenPorts: []models.AuditPort{},
	}

	commonPorts := []struct {
		port    int
		service string
	}{
		{80, "HTTP"},
		{443, "HTTPS"},
		{22, "SSH"},
		{21, "FTP"},
		{25, "SMTP"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{27017, "MongoDB"},
		{8080, "HTTP-ALT"},
	}

	result.TotalScanned = len(commonPorts)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range commonPorts {
		wg.Add(1)
		go func(port int, service string) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
			if err == nil {
				conn.Close()
				mu.Lock()
				result.OpenPorts = append(result.OpenPorts, models.AuditPort{
					Port:    port,
					State:   "open",
					Service: service,
				})
				mu.Unlock()
			}
		}(p.port, p.service)
	}

	wg.Wait()

	// Score based on expected ports
	expectedOpen := 0
	for _, p := range result.OpenPorts {
		if p.Port == 80 || p.Port == 443 {
			expectedOpen++
		}
	}

	result.Score = 50 + (expectedOpen * 25) // 50 base + 25 for each expected port

	return result
}

func (u *AuditUsecase) calculateScores(report *models.AuditReport) {
	totalScore := 0
	sections := 0

	if report.Connectivity != nil {
		totalScore += report.Connectivity.Score
		sections++
	}
	if report.SSL != nil {
		totalScore += report.SSL.Score
		sections++
	}
	if report.SecurityHeaders != nil {
		totalScore += report.SecurityHeaders.Score
		sections++
	}
	if report.Performance != nil {
		totalScore += report.Performance.Score
		sections++
	}
	if report.NetworkPath != nil {
		totalScore += report.NetworkPath.Score
		sections++
	}
	if report.PortScan != nil {
		totalScore += report.PortScan.Score
		sections++
	}

	if sections > 0 {
		report.OverallScore = totalScore / sections
	}

	// Grade
	switch {
	case report.OverallScore >= 90:
		report.OverallGrade = "A+"
	case report.OverallScore >= 80:
		report.OverallGrade = "A"
	case report.OverallScore >= 70:
		report.OverallGrade = "B"
	case report.OverallScore >= 60:
		report.OverallGrade = "C"
	case report.OverallScore >= 50:
		report.OverallGrade = "D"
	default:
		report.OverallGrade = "F"
	}

	// Status
	switch {
	case report.OverallScore >= 80:
		report.OverallStatus = models.AuditStatusHealthy
	case report.OverallScore >= 50:
		report.OverallStatus = models.AuditStatusWarning
	case report.OverallScore > 0:
		report.OverallStatus = models.AuditStatusCritical
	default:
		report.OverallStatus = models.AuditStatusDown
	}
}

func (u *AuditUsecase) generateIssuesAndWarnings(report *models.AuditReport) {
	// SSL Issues
	if report.SSL != nil {
		if !report.SSL.Valid {
			report.Issues = append(report.Issues, models.AuditIssue{
				Severity:    "critical",
				Category:    "SSL",
				Title:       "Certificado SSL inválido",
				Description: strings.Join(report.SSL.Errors, "; "),
				Solution:    "Renove ou corrija o certificado SSL",
			})
		} else if report.SSL.DaysUntilExpiry < 30 {
			report.Warnings = append(report.Warnings, models.AuditWarning{
				Category:   "SSL",
				Message:    fmt.Sprintf("Certificado expira em %d dias", report.SSL.DaysUntilExpiry),
				Suggestion: "Renove o certificado antes da expiração",
			})
		}

		if report.SSL.Version == "TLS 1.0" || report.SSL.Version == "TLS 1.1" {
			report.Issues = append(report.Issues, models.AuditIssue{
				Severity:    "high",
				Category:    "SSL",
				Title:       "Versão TLS desatualizada",
				Description: fmt.Sprintf("Usando %s, que é considerado inseguro", report.SSL.Version),
				Solution:    "Configure o servidor para usar TLS 1.2 ou superior",
			})
		}
	}

	// Security Headers Issues
	if report.SecurityHeaders != nil {
		criticalHeaders := []string{"HSTS", "CSP"}
		for _, h := range criticalHeaders {
			for _, m := range report.SecurityHeaders.Missing {
				if m == h {
					report.Issues = append(report.Issues, models.AuditIssue{
						Severity:    "high",
						Category:    "Headers",
						Title:       fmt.Sprintf("Header %s ausente", h),
						Description: fmt.Sprintf("O header de segurança %s não está configurado", h),
						Solution:    fmt.Sprintf("Adicione o header %s na configuração do servidor", h),
					})
				}
			}
		}

		// Warnings for other missing headers
		for _, m := range report.SecurityHeaders.Missing {
			if m != "HSTS" && m != "CSP" {
				report.Warnings = append(report.Warnings, models.AuditWarning{
					Category:   "Headers",
					Message:    fmt.Sprintf("Header %s não configurado", m),
					Suggestion: "Considere adicionar para melhorar a segurança",
				})
			}
		}
	}

	// Performance Issues
	if report.Performance != nil {
		if report.Performance.TotalTime > 3*time.Second {
			report.Issues = append(report.Issues, models.AuditIssue{
				Severity:    "medium",
				Category:    "Performance",
				Title:       "Tempo de resposta alto",
				Description: fmt.Sprintf("O servidor demora %v para responder", report.Performance.TotalTime.Round(time.Millisecond)),
				Solution:    "Otimize o backend, adicione cache ou use CDN",
			})
		} else if report.Performance.TotalTime > 1*time.Second {
			report.Warnings = append(report.Warnings, models.AuditWarning{
				Category:   "Performance",
				Message:    fmt.Sprintf("Tempo de resposta de %v", report.Performance.TotalTime.Round(time.Millisecond)),
				Suggestion: "Considere otimizações de performance",
			})
		}

		if report.Performance.TLSHandshake > 500*time.Millisecond {
			report.Warnings = append(report.Warnings, models.AuditWarning{
				Category:   "Performance",
				Message:    "Handshake TLS lento",
				Suggestion: "Considere usar TLS 1.3 ou session resumption",
			})
		}
	}

	// Connectivity Issues
	if report.Connectivity != nil {
		if !report.Connectivity.HTTPWorks {
			report.Issues = append(report.Issues, models.AuditIssue{
				Severity:    "critical",
				Category:    "Connectivity",
				Title:       "HTTP não funciona",
				Description: "Não foi possível fazer requisição HTTP",
				Solution:    "Verifique se o servidor web está rodando",
			})
		}
	}
}

func (u *AuditUsecase) generateRecommendations(report *models.AuditReport) {
	// SSL Recommendations
	if report.SSL != nil {
		if report.SSL.Version != "TLS 1.3" {
			report.Recommendations = append(report.Recommendations, "Atualize para TLS 1.3 para melhor segurança e performance")
		}
		if report.SSL.DaysUntilExpiry < 90 {
			report.Recommendations = append(report.Recommendations, "Configure renovação automática de certificado (ex: certbot)")
		}
	}

	// Security Headers Recommendations
	if report.SecurityHeaders != nil && len(report.SecurityHeaders.Missing) > 0 {
		report.Recommendations = append(report.Recommendations, "Adicione headers de segurança faltantes para melhorar a proteção")
	}

	// Performance Recommendations
	if report.Performance != nil {
		if report.Performance.DNSLookup > 100*time.Millisecond {
			report.Recommendations = append(report.Recommendations, "Considere usar um DNS mais rápido ou DNS caching")
		}
		if report.Performance.TotalTime > 2*time.Second {
			report.Recommendations = append(report.Recommendations, "Implemente caching e considere usar CDN")
		}
	}
}

// ExportJSON exports the report as JSON
func (u *AuditUsecase) ExportJSON(report *models.AuditReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// ExportMarkdown exports the report as Markdown
func (u *AuditUsecase) ExportMarkdown(report *models.AuditReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Relatório de Auditoria: %s\n\n", report.Target))
	sb.WriteString(fmt.Sprintf("**Data:** %s\n\n", report.Timestamp.Format("02/01/2006 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Duração:** %s\n\n", report.Duration.Round(time.Millisecond)))

	// Overall Score
	sb.WriteString("## Resultado Geral\n\n")
	sb.WriteString(fmt.Sprintf("| Métrica | Valor |\n"))
	sb.WriteString(fmt.Sprintf("|---------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Score | %d/100 |\n", report.OverallScore))
	sb.WriteString(fmt.Sprintf("| Nota | %s |\n", report.OverallGrade))
	sb.WriteString(fmt.Sprintf("| Status | %s |\n\n", report.OverallStatus))

	// Sections
	if report.Connectivity != nil {
		sb.WriteString("## Conectividade\n\n")
		sb.WriteString(fmt.Sprintf("- DNS: %v (%s)\n", report.Connectivity.DNSResolves, report.Connectivity.DNSTime.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- TCP: %v (%s)\n", report.Connectivity.TCPConnects, report.Connectivity.TCPTime.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- HTTP: %v (Status %d, %s)\n\n", report.Connectivity.HTTPWorks, report.Connectivity.HTTPStatus, report.Connectivity.HTTPTime.Round(time.Millisecond)))
	}

	if report.SSL != nil {
		sb.WriteString("## SSL/TLS\n\n")
		sb.WriteString(fmt.Sprintf("- Versão: %s\n", report.SSL.Version))
		sb.WriteString(fmt.Sprintf("- Cipher: %s\n", report.SSL.CipherSuite))
		sb.WriteString(fmt.Sprintf("- Certificado: %s\n", report.SSL.Certificate))
		sb.WriteString(fmt.Sprintf("- Emissor: %s\n", report.SSL.Issuer))
		sb.WriteString(fmt.Sprintf("- Expira: %s (%d dias)\n\n", report.SSL.ExpiresAt.Format("02/01/2006"), report.SSL.DaysUntilExpiry))
	}

	if report.SecurityHeaders != nil {
		sb.WriteString("## Headers de Segurança\n\n")
		sb.WriteString(fmt.Sprintf("**Nota: %s** (Score: %d/100)\n\n", report.SecurityHeaders.Grade, report.SecurityHeaders.Score))
		sb.WriteString("| Header | Status |\n")
		sb.WriteString("|--------|--------|\n")
		for _, h := range report.SecurityHeaders.Headers {
			status := "❌ Ausente"
			if h.Present {
				status = "✅ Presente"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", h.Name, status))
		}
		sb.WriteString("\n")
	}

	if report.Performance != nil {
		sb.WriteString("## Performance\n\n")
		sb.WriteString(fmt.Sprintf("- DNS Lookup: %s\n", report.Performance.DNSLookup.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- TCP Connect: %s\n", report.Performance.TCPConnect.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- TLS Handshake: %s\n", report.Performance.TLSHandshake.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- Server Processing: %s\n", report.Performance.ServerProcessing.Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("- **Total: %s** (%s)\n\n", report.Performance.TotalTime.Round(time.Millisecond), report.Performance.Rating))
	}

	// Issues
	if len(report.Issues) > 0 {
		sb.WriteString("## ⚠️ Problemas Encontrados\n\n")
		for _, issue := range report.Issues {
			sb.WriteString(fmt.Sprintf("### [%s] %s\n\n", strings.ToUpper(issue.Severity), issue.Title))
			sb.WriteString(fmt.Sprintf("%s\n\n", issue.Description))
			if issue.Solution != "" {
				sb.WriteString(fmt.Sprintf("**Solução:** %s\n\n", issue.Solution))
			}
		}
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		sb.WriteString("## 💡 Recomendações\n\n")
		for i, rec := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	sb.WriteString("*Relatório gerado por jorge-cli*\n")

	return sb.String()
}

// SaveToFile saves the report to a file
func (u *AuditUsecase) SaveToFile(report *models.AuditReport, filename string, format string) error {
	var content []byte
	var err error

	switch format {
	case "json":
		content, err = u.ExportJSON(report)
	case "md", "markdown":
		content = []byte(u.ExportMarkdown(report))
	default:
		return fmt.Errorf("formato não suportado: %s", format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}
