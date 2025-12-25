package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// JWTUsecase handles JWT operations
type JWTUsecase struct {
	logger *zap.Logger
}

// NewJWTUsecase creates a new JWTUsecase
func NewJWTUsecase(logger *zap.Logger) *JWTUsecase {
	return &JWTUsecase{logger: logger}
}

// Decode decodes a JWT token without validation
func (u *JWTUsecase) Decode(token string) *models.JWTResult {
	result := &models.JWTResult{
		Raw:    token,
		Errors: []string{},
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Split into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		result.Errors = append(result.Errors, "Token inválido: esperado 3 partes separadas por '.'")
		return result
	}

	// Decode header
	headerBytes, err := u.base64Decode(parts[0])
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Erro ao decodificar header: %v", err))
	} else {
		var header map[string]interface{}
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Erro ao parsear header JSON: %v", err))
		} else {
			result.Header = header
		}
	}

	// Decode payload
	payloadBytes, err := u.base64Decode(parts[1])
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Erro ao decodificar payload: %v", err))
	} else {
		var payload map[string]interface{}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Erro ao parsear payload JSON: %v", err))
		} else {
			result.Payload = payload
			result.Claims = u.extractClaims(payload)
		}
	}

	// Store signature (as base64)
	result.Signature = parts[2]

	// Check validity
	if len(result.Errors) == 0 {
		result.Valid = true
	}

	return result
}

// Validate validates JWT expiration and other claims
func (u *JWTUsecase) Validate(token string) *models.JWTResult {
	result := u.Decode(token)

	if result.Claims != nil {
		now := time.Now()

		// Check expiration
		if !result.Claims.ExpiresAt.IsZero() {
			if now.After(result.Claims.ExpiresAt) {
				result.Claims.IsExpired = true
				result.Errors = append(result.Errors, "Token expirado")
				result.Valid = false
			} else {
				remaining := result.Claims.ExpiresAt.Sub(now)
				result.Claims.ExpiresIn = u.formatDuration(remaining)
			}
		}

		// Check not before
		if !result.Claims.NotBefore.IsZero() {
			if now.Before(result.Claims.NotBefore) {
				result.Errors = append(result.Errors, "Token ainda não é válido (nbf)")
				result.Valid = false
			}
		}
	}

	return result
}

func (u *JWTUsecase) base64Decode(input string) ([]byte, error) {
	// Add padding if necessary
	switch len(input) % 4 {
	case 2:
		input += "=="
	case 3:
		input += "="
	}

	// Try URL-safe base64 first
	decoded, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		// Try standard base64
		decoded, err = base64.StdEncoding.DecodeString(input)
	}
	return decoded, err
}

func (u *JWTUsecase) extractClaims(payload map[string]interface{}) *models.JWTClaims {
	claims := &models.JWTClaims{}

	if iss, ok := payload["iss"].(string); ok {
		claims.Issuer = iss
	}
	if sub, ok := payload["sub"].(string); ok {
		claims.Subject = sub
	}
	if jti, ok := payload["jti"].(string); ok {
		claims.JWTID = jti
	}

	// Audience can be string or array
	if aud, ok := payload["aud"].(string); ok {
		claims.Audience = []string{aud}
	} else if audArr, ok := payload["aud"].([]interface{}); ok {
		for _, a := range audArr {
			if s, ok := a.(string); ok {
				claims.Audience = append(claims.Audience, s)
			}
		}
	}

	// Time claims (Unix timestamps)
	if exp, ok := payload["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if nbf, ok := payload["nbf"].(float64); ok {
		claims.NotBefore = time.Unix(int64(nbf), 0)
	}
	if iat, ok := payload["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}

	return claims
}

func (u *JWTUsecase) formatDuration(d time.Duration) string {
	if d < 0 {
		return "expirado"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
