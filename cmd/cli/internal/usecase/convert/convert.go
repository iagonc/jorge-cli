package convert

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// ConvertUsecase handles format conversions
type ConvertUsecase struct {
	logger *zap.Logger
}

// NewConvertUsecase creates a new ConvertUsecase
func NewConvertUsecase(logger *zap.Logger) *ConvertUsecase {
	return &ConvertUsecase{logger: logger}
}

// Convert converts between formats
func (u *ConvertUsecase) Convert(input string, from, to models.ConvertFormat) *models.ConvertResult {
	result := &models.ConvertResult{
		InputFormat:  from,
		OutputFormat: to,
		Input:        input,
	}

	// Parse input
	var data interface{}
	var err error

	switch from {
	case models.FormatJSON:
		err = json.Unmarshal([]byte(input), &data)
	case models.FormatYAML:
		err = yaml.Unmarshal([]byte(input), &data)
	case models.FormatENV:
		data, err = u.parseEnv(input)
	default:
		err = fmt.Errorf("formato de entrada não suportado: %s", from)
	}

	if err != nil {
		result.Error = fmt.Sprintf("Erro ao parsear %s: %v", from, err)
		return result
	}

	// Output
	var output []byte

	switch to {
	case models.FormatJSON:
		output, err = json.MarshalIndent(data, "", "  ")
	case models.FormatYAML:
		output, err = yaml.Marshal(data)
	case models.FormatENV:
		output, err = u.toEnv(data)
	default:
		err = fmt.Errorf("formato de saída não suportado: %s", to)
	}

	if err != nil {
		result.Error = fmt.Sprintf("Erro ao converter para %s: %v", to, err)
		return result
	}

	result.Output = string(output)
	return result
}

// ConvertFile converts a file between formats
func (u *ConvertUsecase) ConvertFile(inputPath string, to models.ConvertFormat) *models.ConvertResult {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return &models.ConvertResult{Error: fmt.Sprintf("Erro ao ler arquivo: %v", err)}
	}

	from := u.detectFormat(inputPath, string(content))
	return u.Convert(string(content), from, to)
}

// DetectFormat detects the format of content
func (u *ConvertUsecase) detectFormat(filename, content string) models.ConvertFormat {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".json":
		return models.FormatJSON
	case ".yaml", ".yml":
		return models.FormatYAML
	case ".toml":
		return models.FormatTOML
	case ".env":
		return models.FormatENV
	}

	// Try to detect from content
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return models.FormatJSON
	}

	if strings.Contains(content, ":") && !strings.Contains(content, "=") {
		return models.FormatYAML
	}

	if strings.Contains(content, "=") {
		return models.FormatENV
	}

	return models.FormatJSON // Default
}

func (u *ConvertUsecase) parseEnv(input string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		result[key] = value
	}

	return result, nil
}

func (u *ConvertUsecase) toEnv(data interface{}) ([]byte, error) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("dados devem ser um objeto para converter para ENV")
	}

	var lines []string
	for k, v := range obj {
		key := strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		value := fmt.Sprintf("%v", v)

		// Quote if contains spaces
		if strings.Contains(value, " ") {
			value = fmt.Sprintf("\"%s\"", value)
		}

		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n")), nil
}

// GetSupportedFormats returns supported formats
func (u *ConvertUsecase) GetSupportedFormats() []string {
	return []string{"json", "yaml", "env"}
}
