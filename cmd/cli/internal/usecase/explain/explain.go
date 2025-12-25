package explain

import (
	"strings"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// ExplainUsecase handles error explanations
type ExplainUsecase struct {
	logger *zap.Logger
}

// NewExplainUsecase creates a new ExplainUsecase
func NewExplainUsecase(logger *zap.Logger) *ExplainUsecase {
	return &ExplainUsecase{logger: logger}
}

// Explain finds an explanation for an error
func (u *ExplainUsecase) Explain(errorText string) *models.ErrorExplanation {
	errorText = strings.ToLower(errorText)

	// Try to find exact match first
	for key, explanation := range models.KnownErrors {
		if strings.Contains(errorText, strings.ToLower(key)) {
			return &explanation
		}
	}

	// Try related errors
	for _, explanation := range models.KnownErrors {
		for _, related := range explanation.RelatedErrors {
			if strings.Contains(errorText, strings.ToLower(related)) {
				return &explanation
			}
		}
	}

	// Check for common patterns
	if strings.Contains(errorText, "refused") {
		exp := models.KnownErrors["connection refused"]
		return &exp
	}

	if strings.Contains(errorText, "timeout") || strings.Contains(errorText, "timed out") || strings.Contains(errorText, "deadline") {
		exp := models.KnownErrors["timeout"]
		return &exp
	}

	if strings.Contains(errorText, "dns") || strings.Contains(errorText, "lookup") || strings.Contains(errorText, "resolve") || strings.Contains(errorText, "nxdomain") {
		exp := models.KnownErrors["no such host"]
		return &exp
	}

	if strings.Contains(errorText, "certificate") || strings.Contains(errorText, "x509") || strings.Contains(errorText, "tls") || strings.Contains(errorText, "ssl") {
		exp := models.KnownErrors["certificate"]
		return &exp
	}

	if strings.Contains(errorText, "permission") || strings.Contains(errorText, "denied") || strings.Contains(errorText, "access") {
		exp := models.KnownErrors["permission denied"]
		return &exp
	}

	if strings.Contains(errorText, "unreachable") || strings.Contains(errorText, "no route") {
		exp := models.KnownErrors["network unreachable"]
		return &exp
	}

	if strings.Contains(errorText, "reset") {
		exp := models.KnownErrors["connection reset"]
		return &exp
	}

	if strings.Contains(errorText, "too many") || strings.Contains(errorText, "open files") {
		exp := models.KnownErrors["too many open files"]
		return &exp
	}

	// Unknown error
	return &models.ErrorExplanation{
		Error:       errorText,
		Category:    models.ErrorCategoryUnknown,
		Title:       "Erro Desconhecido",
		Description: "Este erro não está em nossa base de conhecimento.",
		CommonCauses: []string{
			"Erro específico da aplicação",
			"Configuração incorreta",
			"Bug no código",
		},
		Solutions: []models.ErrorSolution{
			{
				Title:       "Pesquisar o erro",
				Description: "Procure a mensagem exata no Google ou Stack Overflow",
				Priority:    1,
			},
			{
				Title:       "Verificar logs",
				Description: "Analise os logs completos para mais contexto",
				Priority:    2,
			},
			{
				Title:       "Usar jorge diagnose",
				Description: "Execute um diagnóstico completo do alvo",
				Command:     "jorge diagnose <host>",
				Priority:    3,
			},
		},
	}
}

// ListKnownErrors returns all known errors
func (u *ExplainUsecase) ListKnownErrors() []string {
	var errors []string
	for key := range models.KnownErrors {
		errors = append(errors, key)
	}
	return errors
}

// GetCategories returns all error categories with counts
func (u *ExplainUsecase) GetCategories() map[models.ErrorCategory]int {
	categories := make(map[models.ErrorCategory]int)
	for _, exp := range models.KnownErrors {
		categories[exp.Category]++
	}
	return categories
}
