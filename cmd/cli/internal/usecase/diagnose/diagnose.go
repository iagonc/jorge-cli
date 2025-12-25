package diagnose

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// DiagnoseUsecase handles comprehensive network diagnostics
type DiagnoseUsecase struct {
	logger *zap.Logger
}

// NewDiagnoseUsecase creates a new DiagnoseUsecase
func NewDiagnoseUsecase(logger *zap.Logger) *DiagnoseUsecase {
	return &DiagnoseUsecase{
		logger: logger,
	}
}

// Diagnose performs comprehensive diagnostics on a target
func (u *DiagnoseUsecase) Diagnose(ctx context.Context, target string) (*models.DiagnoseResult, error) {
	startTime := time.Now()

	u.logger.Info("Starting comprehensive diagnosis", zap.String("target", target))

	result := &models.DiagnoseResult{
		Target:    target,
		StartTime: startTime,
		Checks:    []models.DiagnoseCheck{},
		Problems:  []models.Problem{},
		Suggestions: []models.Suggestion{},
	}

	// Parse target to extract host and determine if it's a URL
	host, port, isURL, parsedURL := u.parseTarget(target)

	// Run checks concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	checksChan := make(chan models.DiagnoseCheck, 10)

	// 1. DNS Check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := u.checkDNS(ctx, host)
		checksChan <- check
	}()

	// 2. Connectivity Check (TCP)
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := u.checkConnectivity(ctx, host, port)
		checksChan <- check
	}()

	// 3. SSL Check (if HTTPS or port 443)
	if port == 443 || (isURL && parsedURL.Scheme == "https") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			check := u.checkSSL(ctx, host, port)
			checksChan <- check
		}()
	}

	// 4. HTTP Check (if URL provided)
	if isURL {
		wg.Add(1)
		go func() {
			defer wg.Done()
			check := u.checkHTTP(ctx, parsedURL)
			checksChan <- check
		}()
	}

	// 5. Latency Check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := u.checkLatency(ctx, host, port)
		checksChan <- check
	}()

	// 6. Common Ports Check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := u.checkCommonPorts(ctx, host)
		checksChan <- check
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(checksChan)
	}()

	for check := range checksChan {
		mu.Lock()
		result.Checks = append(result.Checks, check)
		mu.Unlock()
	}

	// Analyze results and generate problems/suggestions
	u.analyzeResults(result)

	result.Duration = time.Since(startTime)

	return result, nil
}

func (u *DiagnoseUsecase) parseTarget(target string) (host string, port int, isURL bool, parsed *url.URL) {
	// Check if it's a URL
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		parsed, err := url.Parse(target)
		if err == nil {
			host = parsed.Hostname()
			port = 80
			if parsed.Scheme == "https" {
				port = 443
			}
			if parsed.Port() != "" {
				fmt.Sscanf(parsed.Port(), "%d", &port)
			}
			return host, port, true, parsed
		}
	}

	// Check if it's host:port
	if strings.Contains(target, ":") {
		h, p, err := net.SplitHostPort(target)
		if err == nil {
			host = h
			fmt.Sscanf(p, "%d", &port)
			return host, port, false, nil
		}
	}

	// Just a hostname
	return target, 443, false, nil
}

func (u *DiagnoseUsecase) checkDNS(ctx context.Context, host string) models.DiagnoseCheck {
	start := time.Now()
	check := models.DiagnoseCheck{
		Name:     "Resolução DNS",
		Category: models.CategoryDNS,
	}

	// Skip if it's an IP
	if net.ParseIP(host) != nil {
		check.Status = models.CheckSkipped
		check.Message = "Target é um IP, DNS não necessário"
		check.Duration = time.Since(start)
		return check
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, address)
		},
	}

	ips, err := resolver.LookupIP(ctx, "ip4", host)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = models.CheckFailed
		check.Message = "Falha na resolução DNS"
		check.Details = fmt.Sprintf("Não foi possível resolver o domínio '%s'. Erro: %v", host, err)
		return check
	}

	if len(ips) == 0 {
		check.Status = models.CheckFailed
		check.Message = "Nenhum IP encontrado"
		check.Details = fmt.Sprintf("O domínio '%s' não possui registros A", host)
		return check
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	check.Status = models.CheckPassed
	check.Message = fmt.Sprintf("Resolvido para %s", strings.Join(ipStrings, ", "))
	check.Details = fmt.Sprintf("Tempo de resolução: %v", check.Duration.Round(time.Millisecond))
	check.RawData = ipStrings

	return check
}

func (u *DiagnoseUsecase) checkConnectivity(ctx context.Context, host string, port int) models.DiagnoseCheck {
	start := time.Now()
	check := models.DiagnoseCheck{
		Name:     "Conectividade TCP",
		Category: models.CategoryConnectivity,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = models.CheckFailed
		check.Message = fmt.Sprintf("Não foi possível conectar na porta %d", port)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			check.Details = "Conexão expirou (timeout). Possível firewall bloqueando."
		} else if strings.Contains(err.Error(), "connection refused") {
			check.Details = "Conexão recusada. O serviço pode não estar rodando."
		} else if strings.Contains(err.Error(), "no route to host") {
			check.Details = "Sem rota para o host. Problema de rede ou host inacessível."
		} else {
			check.Details = err.Error()
		}
		return check
	}
	defer conn.Close()

	check.Status = models.CheckPassed
	check.Message = fmt.Sprintf("Porta %d acessível", port)
	check.Details = fmt.Sprintf("Conexão estabelecida em %v", check.Duration.Round(time.Millisecond))

	return check
}

func (u *DiagnoseUsecase) checkSSL(ctx context.Context, host string, port int) models.DiagnoseCheck {
	start := time.Now()
	check := models.DiagnoseCheck{
		Name:     "Certificado SSL/TLS",
		Category: models.CategorySSL,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config: &tls.Config{
			ServerName: host,
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = models.CheckFailed
		check.Message = "Falha na conexão SSL/TLS"

		if strings.Contains(err.Error(), "certificate") {
			check.Details = "Problema com o certificado: " + err.Error()
		} else if strings.Contains(err.Error(), "handshake") {
			check.Details = "Falha no handshake TLS. Possível incompatibilidade de protocolo."
		} else {
			check.Details = err.Error()
		}
		return check
	}
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	state := tlsConn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		check.Status = models.CheckFailed
		check.Message = "Nenhum certificado recebido"
		return check
	}

	cert := state.PeerCertificates[0]
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	// Check expiry
	if daysUntilExpiry < 0 {
		check.Status = models.CheckFailed
		check.Message = "Certificado EXPIRADO!"
		check.Details = fmt.Sprintf("Expirou em %s (há %d dias)", cert.NotAfter.Format("02/01/2006"), -daysUntilExpiry)
		return check
	}

	if daysUntilExpiry < 7 {
		check.Status = models.CheckFailed
		check.Message = fmt.Sprintf("Certificado expira em %d dias!", daysUntilExpiry)
		check.Details = fmt.Sprintf("Expira em %s. AÇÃO URGENTE necessária!", cert.NotAfter.Format("02/01/2006"))
		return check
	}

	if daysUntilExpiry < 30 {
		check.Status = models.CheckWarning
		check.Message = fmt.Sprintf("Certificado expira em %d dias", daysUntilExpiry)
		check.Details = fmt.Sprintf("Expira em %s. Considere renovar em breve.", cert.NotAfter.Format("02/01/2006"))
		return check
	}

	check.Status = models.CheckPassed
	check.Message = fmt.Sprintf("Certificado válido (%d dias restantes)", daysUntilExpiry)
	check.Details = fmt.Sprintf("Emitido para: %s | Expira: %s | TLS %s",
		cert.Subject.CommonName,
		cert.NotAfter.Format("02/01/2006"),
		tlsVersionName(state.Version))
	check.RawData = map[string]interface{}{
		"subject":     cert.Subject.CommonName,
		"issuer":      cert.Issuer.CommonName,
		"expiry":      cert.NotAfter,
		"days_left":   daysUntilExpiry,
		"tls_version": tlsVersionName(state.Version),
	}

	return check
}

func (u *DiagnoseUsecase) checkHTTP(ctx context.Context, parsedURL *url.URL) models.DiagnoseCheck {
	start := time.Now()
	check := models.DiagnoseCheck{
		Name:     "Resposta HTTP",
		Category: models.CategoryHTTP,
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		check.Status = models.CheckFailed
		check.Message = "Erro ao criar requisição"
		check.Details = err.Error()
		check.Duration = time.Since(start)
		return check
	}

	req.Header.Set("User-Agent", "jorge-cli/1.0")

	resp, err := client.Do(req)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = models.CheckFailed
		check.Message = "Falha na requisição HTTP"

		if strings.Contains(err.Error(), "timeout") {
			check.Details = "Timeout na requisição. O servidor pode estar sobrecarregado."
		} else if strings.Contains(err.Error(), "refused") {
			check.Details = "Conexão recusada. O servidor HTTP pode não estar rodando."
		} else if strings.Contains(err.Error(), "certificate") {
			check.Details = "Erro de certificado SSL: " + err.Error()
		} else {
			check.Details = err.Error()
		}
		return check
	}
	defer resp.Body.Close()

	// Analyze status code
	if resp.StatusCode >= 500 {
		check.Status = models.CheckFailed
		check.Message = fmt.Sprintf("Erro do servidor (HTTP %d)", resp.StatusCode)
		check.Details = "O servidor retornou um erro interno. Verifique os logs do servidor."
	} else if resp.StatusCode >= 400 {
		check.Status = models.CheckWarning
		check.Message = fmt.Sprintf("Erro do cliente (HTTP %d)", resp.StatusCode)
		if resp.StatusCode == 401 {
			check.Details = "Autenticação necessária."
		} else if resp.StatusCode == 403 {
			check.Details = "Acesso proibido. Verifique permissões."
		} else if resp.StatusCode == 404 {
			check.Details = "Recurso não encontrado. Verifique a URL."
		} else {
			check.Details = fmt.Sprintf("Status: %s", resp.Status)
		}
	} else if resp.StatusCode >= 300 {
		check.Status = models.CheckPassed
		check.Message = fmt.Sprintf("Redirecionamento (HTTP %d)", resp.StatusCode)
		if loc := resp.Header.Get("Location"); loc != "" {
			check.Details = fmt.Sprintf("Redireciona para: %s", loc)
		}
	} else {
		check.Status = models.CheckPassed
		check.Message = fmt.Sprintf("HTTP %d OK", resp.StatusCode)
		check.Details = fmt.Sprintf("Tempo de resposta: %v | Server: %s",
			check.Duration.Round(time.Millisecond),
			resp.Header.Get("Server"))
	}

	check.RawData = map[string]interface{}{
		"status_code":   resp.StatusCode,
		"response_time": check.Duration,
		"server":        resp.Header.Get("Server"),
		"content_type":  resp.Header.Get("Content-Type"),
	}

	return check
}

func (u *DiagnoseUsecase) checkLatency(ctx context.Context, host string, port int) models.DiagnoseCheck {
	check := models.DiagnoseCheck{
		Name:     "Latência",
		Category: models.CategoryLatency,
	}

	var latencies []time.Duration
	address := fmt.Sprintf("%s:%d", host, port)

	for i := 0; i < 3; i++ {
		start := time.Now()
		dialer := &net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			continue
		}
		latencies = append(latencies, time.Since(start))
		conn.Close()
	}

	if len(latencies) == 0 {
		check.Status = models.CheckSkipped
		check.Message = "Não foi possível medir latência"
		check.Details = "Falha em todas as tentativas de conexão"
		return check
	}

	var total time.Duration
	min := latencies[0]
	max := latencies[0]
	for _, l := range latencies {
		total += l
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}
	avg := total / time.Duration(len(latencies))
	check.Duration = avg

	// Classify latency
	avgMs := avg.Milliseconds()
	if avgMs < 50 {
		check.Status = models.CheckPassed
		check.Message = fmt.Sprintf("Excelente (%dms)", avgMs)
		check.Details = fmt.Sprintf("Min: %dms | Avg: %dms | Max: %dms", min.Milliseconds(), avgMs, max.Milliseconds())
	} else if avgMs < 150 {
		check.Status = models.CheckPassed
		check.Message = fmt.Sprintf("Boa (%dms)", avgMs)
		check.Details = fmt.Sprintf("Min: %dms | Avg: %dms | Max: %dms", min.Milliseconds(), avgMs, max.Milliseconds())
	} else if avgMs < 300 {
		check.Status = models.CheckWarning
		check.Message = fmt.Sprintf("Moderada (%dms)", avgMs)
		check.Details = fmt.Sprintf("Min: %dms | Avg: %dms | Max: %dms. Latência acima do ideal.", min.Milliseconds(), avgMs, max.Milliseconds())
	} else {
		check.Status = models.CheckWarning
		check.Message = fmt.Sprintf("Alta (%dms)", avgMs)
		check.Details = fmt.Sprintf("Min: %dms | Avg: %dms | Max: %dms. Conexão lenta.", min.Milliseconds(), avgMs, max.Milliseconds())
	}

	check.RawData = map[string]interface{}{
		"min_ms": min.Milliseconds(),
		"avg_ms": avgMs,
		"max_ms": max.Milliseconds(),
		"samples": len(latencies),
	}

	return check
}

func (u *DiagnoseUsecase) checkCommonPorts(ctx context.Context, host string) models.DiagnoseCheck {
	start := time.Now()
	check := models.DiagnoseCheck{
		Name:     "Portas Comuns",
		Category: models.CategoryFirewall,
	}

	ports := []struct {
		port    int
		service string
	}{
		{80, "HTTP"},
		{443, "HTTPS"},
		{22, "SSH"},
		{21, "FTP"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{27017, "MongoDB"},
	}

	var openPorts []string
	var closedCount int

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range ports {
		wg.Add(1)
		go func(port int, service string) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", host, port)
			dialer := &net.Dialer{Timeout: 2 * time.Second}
			conn, err := dialer.DialContext(ctx, "tcp", address)
			mu.Lock()
			if err == nil {
				openPorts = append(openPorts, fmt.Sprintf("%d (%s)", port, service))
				conn.Close()
			} else {
				closedCount++
			}
			mu.Unlock()
		}(p.port, p.service)
	}

	wg.Wait()
	check.Duration = time.Since(start)

	if len(openPorts) == 0 {
		check.Status = models.CheckWarning
		check.Message = "Nenhuma porta comum aberta"
		check.Details = "Todas as portas verificadas estão fechadas ou filtradas"
	} else {
		check.Status = models.CheckPassed
		check.Message = fmt.Sprintf("%d porta(s) aberta(s)", len(openPorts))
		check.Details = strings.Join(openPorts, ", ")
	}

	check.RawData = map[string]interface{}{
		"open_ports": openPorts,
		"closed":     closedCount,
	}

	return check
}

func (u *DiagnoseUsecase) analyzeResults(result *models.DiagnoseResult) {
	var passed, warnings, failed int

	for _, check := range result.Checks {
		switch check.Status {
		case models.CheckPassed:
			passed++
		case models.CheckWarning:
			warnings++
			u.addWarningProblem(result, check)
		case models.CheckFailed:
			failed++
			u.addFailedProblem(result, check)
		}
	}

	// Set overall status
	if failed > 0 {
		result.Overall = models.StatusError
		result.Summary = fmt.Sprintf("Encontrados %d problema(s) crítico(s)", failed)
	} else if warnings > 0 {
		result.Overall = models.StatusUnhealthy
		result.Summary = fmt.Sprintf("Encontrados %d aviso(s)", warnings)
	} else {
		result.Overall = models.StatusHealthy
		result.Summary = "Todos os testes passaram"
	}

	// Generate suggestions based on problems
	u.generateSuggestions(result)
}

func (u *DiagnoseUsecase) addWarningProblem(result *models.DiagnoseResult, check models.DiagnoseCheck) {
	result.Problems = append(result.Problems, models.Problem{
		Severity:    models.SeverityWarning,
		Category:    check.Category,
		Title:       check.Name + ": " + check.Message,
		Description: check.Details,
	})
}

func (u *DiagnoseUsecase) addFailedProblem(result *models.DiagnoseResult, check models.DiagnoseCheck) {
	result.Problems = append(result.Problems, models.Problem{
		Severity:    models.SeverityCritical,
		Category:    check.Category,
		Title:       check.Name + ": " + check.Message,
		Description: check.Details,
	})
}

func (u *DiagnoseUsecase) generateSuggestions(result *models.DiagnoseResult) {
	priority := 1

	for _, problem := range result.Problems {
		var suggestion models.Suggestion

		switch problem.Category {
		case models.CategoryDNS:
			suggestion = models.Suggestion{
				Priority:    priority,
				Title:       "Verificar configuração DNS",
				Description: "Verifique se o domínio está correto e se seus servidores DNS estão funcionando.",
				Command:     fmt.Sprintf("nslookup %s", result.Target),
			}

		case models.CategoryConnectivity:
			suggestion = models.Suggestion{
				Priority:    priority,
				Title:       "Verificar conectividade de rede",
				Description: "Verifique se o host está acessível e se não há firewalls bloqueando a conexão.",
				Command:     fmt.Sprintf("telnet %s", result.Target),
			}

		case models.CategorySSL:
			if strings.Contains(problem.Title, "EXPIRADO") || strings.Contains(problem.Title, "expira") {
				suggestion = models.Suggestion{
					Priority:    priority,
					Title:       "Renovar certificado SSL",
					Description: "O certificado SSL precisa ser renovado. Use Let's Encrypt ou seu provedor de certificados.",
					Link:        "https://letsencrypt.org/getting-started/",
				}
			} else {
				suggestion = models.Suggestion{
					Priority:    priority,
					Title:       "Verificar configuração SSL",
					Description: "Há um problema com o certificado SSL. Verifique se está instalado corretamente.",
					Command:     fmt.Sprintf("openssl s_client -connect %s:443", result.Target),
				}
			}

		case models.CategoryHTTP:
			suggestion = models.Suggestion{
				Priority:    priority,
				Title:       "Verificar aplicação HTTP",
				Description: "O servidor HTTP está retornando erros. Verifique os logs da aplicação.",
				Command:     fmt.Sprintf("curl -v %s", result.Target),
			}

		case models.CategoryLatency:
			suggestion = models.Suggestion{
				Priority:    priority,
				Title:       "Investigar latência alta",
				Description: "A latência está acima do ideal. Pode ser problema de rede, distância geográfica ou sobrecarga.",
				Command:     fmt.Sprintf("traceroute %s", result.Target),
			}
		}

		if suggestion.Title != "" {
			result.Suggestions = append(result.Suggestions, suggestion)
			priority++
		}
	}
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
