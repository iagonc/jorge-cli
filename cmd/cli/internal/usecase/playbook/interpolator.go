package playbook

import (
	"bytes"
	"os"
	"strings"
	"text/template"
	"time"
)

// Interpolator handles variable substitution in playbooks
type Interpolator struct {
	funcMap template.FuncMap
}

// NewInterpolator creates a new Interpolator
func NewInterpolator() *Interpolator {
	return &Interpolator{
		funcMap: defaultFuncMap(),
	}
}

func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"default": func(defaultVal, val string) string {
			if val == "" {
				return defaultVal
			}
			return val
		},
		"env": os.Getenv,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"trim":      strings.TrimSpace,
		"split": func(sep, s string) []string {
			return strings.Split(s, sep)
		},
		"join": func(sep string, parts []string) string {
			return strings.Join(parts, sep)
		},
		"now": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"date": func(format string) string {
			return time.Now().Format(format)
		},
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"quote": func(s string) string {
			return "\"" + s + "\""
		},
		"eq": func(a, b string) bool {
			return a == b
		},
		"ne": func(a, b string) bool {
			return a != b
		},
	}
}

// Interpolate processes a string with variable substitution
func (i *Interpolator) Interpolate(input string, vars map[string]string) (string, error) {
	// Convert map[string]string to map[string]interface{} for template
	data := make(map[string]interface{})
	for k, v := range vars {
		data[k] = v
	}

	tmpl, err := template.New("").Funcs(i.funcMap).Parse(input)
	if err != nil {
		// If template parsing fails, return original string
		return input, nil
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// If template execution fails, return original string
		return input, nil
	}

	return buf.String(), nil
}

// InterpolateMap processes all string values in a map
func (i *Interpolator) InterpolateMap(input map[string]interface{}, vars map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range input {
		if strVal, ok := v.(string); ok {
			interpolated, err := i.Interpolate(strVal, vars)
			if err != nil {
				return nil, err
			}
			result[k] = interpolated
		} else {
			result[k] = v
		}
	}

	return result, nil
}
