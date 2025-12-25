package diff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// DiffUsecase handles diff operations
type DiffUsecase struct {
	logger *zap.Logger
}

// NewDiffUsecase creates a new DiffUsecase
func NewDiffUsecase(logger *zap.Logger) *DiffUsecase {
	return &DiffUsecase{logger: logger}
}

// DiffFiles compares two files
func (u *DiffUsecase) DiffFiles(file1, file2 string, semantic bool) (*models.DiffResult, error) {
	content1, err := os.ReadFile(file1)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler %s: %v", file1, err)
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler %s: %v", file2, err)
	}

	if semantic {
		return u.diffSemantic(file1, file2, string(content1), string(content2))
	}

	return u.diffText(file1, file2, string(content1), string(content2))
}

// DiffText compares two text strings line by line
func (u *DiffUsecase) diffText(file1, file2, content1, content2 string) (*models.DiffResult, error) {
	result := &models.DiffResult{
		File1:   file1,
		File2:   file2,
		Changes: []models.DiffChange{},
	}

	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	// Simple line-by-line diff (LCS would be better for production)
	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}

	for i := 0; i < maxLen; i++ {
		var line1, line2 string
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			if i >= len(lines1) {
				result.Changes = append(result.Changes, models.DiffChange{
					Type:     models.DiffTypeAdded,
					Line:     i + 1,
					NewValue: line2,
				})
				result.Stats.Added++
			} else if i >= len(lines2) {
				result.Changes = append(result.Changes, models.DiffChange{
					Type:     models.DiffTypeRemoved,
					Line:     i + 1,
					OldValue: line1,
				})
				result.Stats.Removed++
			} else {
				result.Changes = append(result.Changes, models.DiffChange{
					Type:     models.DiffTypeModified,
					Line:     i + 1,
					OldValue: line1,
					NewValue: line2,
				})
				result.Stats.Modified++
			}
		}
	}

	result.Identical = len(result.Changes) == 0
	return result, nil
}

// DiffSemantic compares two structured files (JSON/YAML)
func (u *DiffUsecase) diffSemantic(file1, file2, content1, content2 string) (*models.DiffResult, error) {
	result := &models.DiffResult{
		File1:   file1,
		File2:   file2,
		Changes: []models.DiffChange{},
	}

	var data1, data2 interface{}

	// Try JSON first
	if err := json.Unmarshal([]byte(content1), &data1); err != nil {
		// Try YAML
		if err := yaml.Unmarshal([]byte(content1), &data1); err != nil {
			return u.diffText(file1, file2, content1, content2)
		}
	}

	if err := json.Unmarshal([]byte(content2), &data2); err != nil {
		if err := yaml.Unmarshal([]byte(content2), &data2); err != nil {
			return u.diffText(file1, file2, content1, content2)
		}
	}

	u.compareValues("", data1, data2, result)

	result.Identical = len(result.Changes) == 0
	return result, nil
}

func (u *DiffUsecase) compareValues(path string, v1, v2 interface{}, result *models.DiffResult) {
	if reflect.DeepEqual(v1, v2) {
		return
	}

	// Both nil
	if v1 == nil && v2 == nil {
		return
	}

	// One is nil
	if v1 == nil {
		result.Changes = append(result.Changes, models.DiffChange{
			Type:     models.DiffTypeAdded,
			Path:     path,
			NewValue: fmt.Sprintf("%v", v2),
		})
		result.Stats.Added++
		return
	}
	if v2 == nil {
		result.Changes = append(result.Changes, models.DiffChange{
			Type:     models.DiffTypeRemoved,
			Path:     path,
			OldValue: fmt.Sprintf("%v", v1),
		})
		result.Stats.Removed++
		return
	}

	// Different types
	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		result.Changes = append(result.Changes, models.DiffChange{
			Type:     models.DiffTypeModified,
			Path:     path,
			OldValue: fmt.Sprintf("%v", v1),
			NewValue: fmt.Sprintf("%v", v2),
		})
		result.Stats.Modified++
		return
	}

	// Compare maps
	if m1, ok := v1.(map[string]interface{}); ok {
		m2 := v2.(map[string]interface{})
		u.compareMaps(path, m1, m2, result)
		return
	}

	// Compare arrays
	if a1, ok := v1.([]interface{}); ok {
		a2 := v2.([]interface{})
		u.compareArrays(path, a1, a2, result)
		return
	}

	// Simple values
	result.Changes = append(result.Changes, models.DiffChange{
		Type:     models.DiffTypeModified,
		Path:     path,
		OldValue: fmt.Sprintf("%v", v1),
		NewValue: fmt.Sprintf("%v", v2),
	})
	result.Stats.Modified++
}

func (u *DiffUsecase) compareMaps(path string, m1, m2 map[string]interface{}, result *models.DiffResult) {
	// Check all keys in m1
	for k, v1 := range m1 {
		newPath := k
		if path != "" {
			newPath = path + "." + k
		}

		if v2, ok := m2[k]; ok {
			u.compareValues(newPath, v1, v2, result)
		} else {
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeRemoved,
				Path:     newPath,
				OldValue: fmt.Sprintf("%v", v1),
			})
			result.Stats.Removed++
		}
	}

	// Check for keys in m2 not in m1
	for k, v2 := range m2 {
		if _, ok := m1[k]; !ok {
			newPath := k
			if path != "" {
				newPath = path + "." + k
			}
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeAdded,
				Path:     newPath,
				NewValue: fmt.Sprintf("%v", v2),
			})
			result.Stats.Added++
		}
	}
}

func (u *DiffUsecase) compareArrays(path string, a1, a2 []interface{}, result *models.DiffResult) {
	maxLen := len(a1)
	if len(a2) > maxLen {
		maxLen = len(a2)
	}

	for i := 0; i < maxLen; i++ {
		newPath := fmt.Sprintf("%s[%d]", path, i)

		if i >= len(a1) {
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeAdded,
				Path:     newPath,
				NewValue: fmt.Sprintf("%v", a2[i]),
			})
			result.Stats.Added++
		} else if i >= len(a2) {
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeRemoved,
				Path:     newPath,
				OldValue: fmt.Sprintf("%v", a1[i]),
			})
			result.Stats.Removed++
		} else {
			u.compareValues(newPath, a1[i], a2[i], result)
		}
	}
}

// DiffEnv compares two .env files
func (u *DiffUsecase) DiffEnv(file1, file2 string) (*models.DiffResult, error) {
	env1, err := u.parseEnvFile(file1)
	if err != nil {
		return nil, err
	}

	env2, err := u.parseEnvFile(file2)
	if err != nil {
		return nil, err
	}

	result := &models.DiffResult{
		File1:   file1,
		File2:   file2,
		Changes: []models.DiffChange{},
	}

	// Check all keys in env1
	for k, v1 := range env1 {
		if v2, ok := env2[k]; ok {
			if v1 != v2 {
				result.Changes = append(result.Changes, models.DiffChange{
					Type:     models.DiffTypeModified,
					Path:     k,
					OldValue: v1,
					NewValue: v2,
				})
				result.Stats.Modified++
			}
		} else {
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeRemoved,
				Path:     k,
				OldValue: v1,
			})
			result.Stats.Removed++
		}
	}

	// Check for keys in env2 not in env1
	for k, v2 := range env2 {
		if _, ok := env1[k]; !ok {
			result.Changes = append(result.Changes, models.DiffChange{
				Type:     models.DiffTypeAdded,
				Path:     k,
				NewValue: v2,
			})
			result.Stats.Added++
		}
	}

	result.Identical = len(result.Changes) == 0
	return result, nil
}

func (u *DiffUsecase) parseEnvFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result, scanner.Err()
}
