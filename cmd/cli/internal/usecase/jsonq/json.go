package jsonq

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// JSONUsecase handles JSON operations
type JSONUsecase struct {
	logger *zap.Logger
}

// NewJSONUsecase creates a new JSONUsecase
func NewJSONUsecase(logger *zap.Logger) *JSONUsecase {
	return &JSONUsecase{logger: logger}
}

// Query queries JSON with a simple path expression
func (u *JSONUsecase) Query(jsonStr, query string) *models.JSONQuery {
	result := &models.JSONQuery{
		Input: jsonStr,
		Query: query,
	}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		result.Error = fmt.Sprintf("JSON inválido: %v", err)
		return result
	}

	// Execute query
	queryResult, err := u.executeQuery(data, query)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Result = queryResult
	result.Type = u.getType(queryResult)

	return result
}

// Format formats JSON with indentation
func (u *JSONUsecase) Format(jsonStr string, indent int) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("JSON inválido: %v", err)
	}

	indentStr := strings.Repeat(" ", indent)
	formatted, err := json.MarshalIndent(data, "", indentStr)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// Minify removes whitespace from JSON
func (u *JSONUsecase) Minify(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("JSON inválido: %v", err)
	}

	minified, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}

// Validate validates JSON and returns any errors
func (u *JSONUsecase) Validate(jsonStr string) (bool, string) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// Keys returns all keys from a JSON object
func (u *JSONUsecase) Keys(jsonStr string) ([]string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("JSON deve ser um objeto: %v", err)
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (u *JSONUsecase) executeQuery(data interface{}, query string) (interface{}, error) {
	if query == "" || query == "." {
		return data, nil
	}

	// Remove leading dot
	query = strings.TrimPrefix(query, ".")
	query = strings.TrimPrefix(query, "$.")

	parts := u.parseQueryParts(query)

	current := data
	for _, part := range parts {
		var err error
		current, err = u.accessPath(current, part)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func (u *JSONUsecase) parseQueryParts(query string) []string {
	var parts []string
	var current strings.Builder
	inBracket := false

	for _, ch := range query {
		switch ch {
		case '.':
			if !inBracket && current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else if inBracket {
				current.WriteRune(ch)
			}
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
		case ']':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = false
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (u *JSONUsecase) accessPath(data interface{}, part string) (interface{}, error) {
	// Handle array index
	if idx, err := strconv.Atoi(part); err == nil {
		arr, ok := data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("não é um array no caminho '%s'", part)
		}
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("índice %d fora do range (0-%d)", idx, len(arr)-1)
		}
		return arr[idx], nil
	}

	// Handle object key
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("não é um objeto no caminho '%s'", part)
	}

	val, exists := obj[part]
	if !exists {
		return nil, fmt.Errorf("chave '%s' não encontrada", part)
	}

	return val, nil
}

func (u *JSONUsecase) getType(data interface{}) string {
	switch data.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}
