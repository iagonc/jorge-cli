package headers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// HeadersUsecase handles security headers analysis
type HeadersUsecase struct {
	logger *zap.Logger
}

// NewHeadersUsecase creates a new HeadersUsecase
func NewHeadersUsecase(logger *zap.Logger) *HeadersUsecase {
	return &HeadersUsecase{logger: logger}
}

// Analyze performs security headers analysis
func (u *HeadersUsecase) Analyze(ctx context.Context, url string) (*models.HeadersResult, error) {
	startTime := time.Now()

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "jorge-cli/1.0 security-scanner")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &models.HeadersResult{
		URL:        url,
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Checks:     []models.SecurityHeaderCheck{},
		Duration:   time.Since(startTime),
	}

	// Copy headers
	for k, v := range resp.Header {
		result.Headers[k] = strings.Join(v, ", ")
	}

	// Perform security checks
	u.checkHSTS(result, resp.Header)
	u.checkCSP(result, resp.Header)
	u.checkXContentTypeOptions(result, resp.Header)
	u.checkXFrameOptions(result, resp.Header)
	u.checkXXSSProtection(result, resp.Header)
	u.checkReferrerPolicy(result, resp.Header)
	u.checkPermissionsPolicy(result, resp.Header)
	u.checkCORS(result, resp.Header)
	u.checkCOOP(result, resp.Header)
	u.checkCORP(result, resp.Header)

	// Calculate score and grade
	u.calculateScore(result)

	return result, nil
}

func (u *HeadersUsecase) checkHSTS(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "HSTS",
		Header:      "Strict-Transport-Security",
		Description: "Força conexões HTTPS",
		Impact:      "Previne ataques de downgrade e cookie hijacking",
	}

	value := headers.Get("Strict-Transport-Security")
	if value != "" {
		check.Present = true
		check.Value = value

		if strings.Contains(value, "max-age=") {
			check.Status = models.HeaderStatusPass
			if strings.Contains(value, "includeSubDomains") {
				check.Description = "HSTS ativo com subdomínios"
			}
			if strings.Contains(value, "preload") {
				check.Description = "HSTS ativo com preload"
			}
		} else {
			check.Status = models.HeaderStatusWarn
			check.Suggestion = "Adicione max-age com pelo menos 31536000 (1 ano)"
		}
	} else {
		check.Status = models.HeaderStatusFail
		check.Suggestion = "Adicione: Strict-Transport-Security: max-age=31536000; includeSubDomains"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkCSP(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "CSP",
		Header:      "Content-Security-Policy",
		Description: "Política de segurança de conteúdo",
		Impact:      "Previne XSS e injeção de dados",
	}

	value := headers.Get("Content-Security-Policy")
	if value != "" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusPass

		// Check for unsafe directives
		if strings.Contains(value, "unsafe-inline") || strings.Contains(value, "unsafe-eval") {
			check.Status = models.HeaderStatusWarn
			check.Suggestion = "Evite 'unsafe-inline' e 'unsafe-eval' para maior segurança"
		}
	} else {
		check.Status = models.HeaderStatusFail
		check.Suggestion = "Adicione uma política CSP restritiva"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkXContentTypeOptions(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "X-Content-Type-Options",
		Header:      "X-Content-Type-Options",
		Description: "Previne MIME type sniffing",
		Impact:      "Evita execução de scripts disfarçados",
	}

	value := headers.Get("X-Content-Type-Options")
	if value == "nosniff" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusPass
	} else if value != "" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusWarn
		check.Suggestion = "Use 'nosniff' como valor"
	} else {
		check.Status = models.HeaderStatusFail
		check.Suggestion = "Adicione: X-Content-Type-Options: nosniff"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkXFrameOptions(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "X-Frame-Options",
		Header:      "X-Frame-Options",
		Description: "Controla embedding em iframes",
		Impact:      "Previne clickjacking",
	}

	value := headers.Get("X-Frame-Options")
	if value != "" {
		check.Present = true
		check.Value = value
		if value == "DENY" || value == "SAMEORIGIN" {
			check.Status = models.HeaderStatusPass
		} else {
			check.Status = models.HeaderStatusWarn
			check.Suggestion = "Use 'DENY' ou 'SAMEORIGIN'"
		}
	} else {
		check.Status = models.HeaderStatusFail
		check.Suggestion = "Adicione: X-Frame-Options: DENY"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkXXSSProtection(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "X-XSS-Protection",
		Header:      "X-XSS-Protection",
		Description: "Filtro XSS do navegador (legado)",
		Impact:      "Proteção adicional contra XSS",
	}

	value := headers.Get("X-XSS-Protection")
	if value != "" {
		check.Present = true
		check.Value = value
		if value == "0" {
			check.Status = models.HeaderStatusPass
			check.Description = "Desabilitado (recomendado com CSP)"
		} else if strings.HasPrefix(value, "1") {
			check.Status = models.HeaderStatusInfo
			check.Description = "Habilitado (considere usar CSP)"
		}
	} else {
		check.Status = models.HeaderStatusInfo
		check.Suggestion = "Considere usar CSP em vez deste header"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkReferrerPolicy(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "Referrer-Policy",
		Header:      "Referrer-Policy",
		Description: "Controla informações de referência",
		Impact:      "Protege privacidade do usuário",
	}

	value := headers.Get("Referrer-Policy")
	if value != "" {
		check.Present = true
		check.Value = value
		safeValues := []string{"no-referrer", "same-origin", "strict-origin", "strict-origin-when-cross-origin"}
		isSafe := false
		for _, sv := range safeValues {
			if value == sv {
				isSafe = true
				break
			}
		}
		if isSafe {
			check.Status = models.HeaderStatusPass
		} else {
			check.Status = models.HeaderStatusWarn
			check.Suggestion = "Use 'strict-origin-when-cross-origin' ou mais restritivo"
		}
	} else {
		check.Status = models.HeaderStatusWarn
		check.Suggestion = "Adicione: Referrer-Policy: strict-origin-when-cross-origin"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkPermissionsPolicy(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "Permissions-Policy",
		Header:      "Permissions-Policy",
		Description: "Controla APIs do navegador",
		Impact:      "Limita funcionalidades disponíveis",
	}

	value := headers.Get("Permissions-Policy")
	if value == "" {
		value = headers.Get("Feature-Policy") // Legacy
	}

	if value != "" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusPass
	} else {
		check.Status = models.HeaderStatusInfo
		check.Suggestion = "Considere restringir geolocation, camera, microphone, etc"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkCORS(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "CORS",
		Header:      "Access-Control-Allow-Origin",
		Description: "Controle de acesso cross-origin",
		Impact:      "Define quem pode acessar recursos",
	}

	value := headers.Get("Access-Control-Allow-Origin")
	if value != "" {
		check.Present = true
		check.Value = value
		if value == "*" {
			check.Status = models.HeaderStatusWarn
			check.Suggestion = "Evite '*' para recursos sensíveis"
		} else {
			check.Status = models.HeaderStatusPass
		}
	} else {
		check.Status = models.HeaderStatusInfo
		check.Description = "CORS não configurado (default: same-origin)"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkCOOP(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "COOP",
		Header:      "Cross-Origin-Opener-Policy",
		Description: "Isolamento de contexto de navegação",
		Impact:      "Proteção contra Spectre e ataques cross-origin",
	}

	value := headers.Get("Cross-Origin-Opener-Policy")
	if value != "" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusPass
	} else {
		check.Status = models.HeaderStatusInfo
		check.Suggestion = "Considere: Cross-Origin-Opener-Policy: same-origin"
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) checkCORP(result *models.HeadersResult, headers http.Header) {
	check := models.SecurityHeaderCheck{
		Name:        "CORP",
		Header:      "Cross-Origin-Resource-Policy",
		Description: "Política de recursos cross-origin",
		Impact:      "Controla quem pode carregar recursos",
	}

	value := headers.Get("Cross-Origin-Resource-Policy")
	if value != "" {
		check.Present = true
		check.Value = value
		check.Status = models.HeaderStatusPass
	} else {
		check.Status = models.HeaderStatusInfo
	}

	result.Checks = append(result.Checks, check)
}

func (u *HeadersUsecase) calculateScore(result *models.HeadersResult) {
	totalChecks := len(result.Checks)
	if totalChecks == 0 {
		result.SecurityScore = 0
		result.Grade = "F"
		return
	}

	points := 0
	maxPoints := totalChecks * 10

	for _, check := range result.Checks {
		switch check.Status {
		case models.HeaderStatusPass:
			points += 10
		case models.HeaderStatusWarn:
			points += 5
		case models.HeaderStatusInfo:
			points += 3
		case models.HeaderStatusFail:
			points += 0
		}
	}

	result.SecurityScore = (points * 100) / maxPoints

	// Assign grade
	switch {
	case result.SecurityScore >= 90:
		result.Grade = "A+"
	case result.SecurityScore >= 80:
		result.Grade = "A"
	case result.SecurityScore >= 70:
		result.Grade = "B"
	case result.SecurityScore >= 60:
		result.Grade = "C"
	case result.SecurityScore >= 50:
		result.Grade = "D"
	default:
		result.Grade = "F"
	}

	// Generate suggestions
	for _, check := range result.Checks {
		if check.Suggestion != "" && (check.Status == models.HeaderStatusFail || check.Status == models.HeaderStatusWarn) {
			result.Suggestions = append(result.Suggestions, fmt.Sprintf("%s: %s", check.Name, check.Suggestion))
		}
	}
}
