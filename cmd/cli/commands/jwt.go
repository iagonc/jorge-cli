package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/jwt"
	"github.com/spf13/cobra"
)

// NewJWTCommand creates the jwt command
func NewJWTCommand(usecase *jwt.JWTUsecase) *cobra.Command {
	var validate bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "jwt <token>",
		Short: "Decode e valida tokens JWT",
		Long: `Decodifica tokens JWT e mostra header, payload e claims.

Pode validar expiração e outras claims padrão.
Aceita tokens com ou sem prefixo "Bearer ".`,
		Example: `  # Decodificar token
  jorge jwt "eyJhbGciOiJIUzI1..."

  # Validar token (verifica expiração)
  jorge jwt "eyJhbGciOiJIUzI1..." --validate

  # Output em JSON
  jorge jwt "eyJhbGciOiJIUzI1..." --json

  # Ler de stdin
  echo "eyJhbGciOiJIUzI1..." | jorge jwt`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var token string

			if len(args) > 0 {
				token = args[0]
			} else {
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					bytes, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					token = strings.TrimSpace(string(bytes))
				} else {
					return fmt.Errorf("especifique o token ou use stdin")
				}
			}

			var result *models.JWTResult
			if validate {
				result = usecase.Validate(token)
			} else {
				result = usecase.Decode(token)
			}

			if jsonOutput {
				out, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(out))
			} else {
				displayJWTResult(result)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&validate, "validate", false, "Validar expiração e claims")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output em JSON")

	return cmd
}

func displayJWTResult(result *models.JWTResult) {
	fmt.Println()

	// Status
	if len(result.Errors) > 0 {
		fmt.Printf("   %s Token com problemas\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"))
		for _, err := range result.Errors {
			fmt.Printf("      • %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(err))
		}
		fmt.Println()
	} else {
		fmt.Printf("   %s Token válido\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
	}

	// Header
	if result.Header != nil {
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   HEADER"))
		fmt.Println(strings.Repeat("─", 60))
		displayJSONMap(result.Header, "   ")
		fmt.Println()
	}

	// Payload
	if result.Payload != nil {
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   PAYLOAD"))
		fmt.Println(strings.Repeat("─", 60))
		displayJSONMap(result.Payload, "   ")
		fmt.Println()
	}

	// Claims (parsed)
	if result.Claims != nil {
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   CLAIMS PADRÃO"))
		fmt.Println(strings.Repeat("─", 60))

		if result.Claims.Issuer != "" {
			fmt.Printf("   Issuer (iss):     %s\n", result.Claims.Issuer)
		}
		if result.Claims.Subject != "" {
			fmt.Printf("   Subject (sub):    %s\n", result.Claims.Subject)
		}
		if len(result.Claims.Audience) > 0 {
			fmt.Printf("   Audience (aud):   %s\n", strings.Join(result.Claims.Audience, ", "))
		}
		if result.Claims.JWTID != "" {
			fmt.Printf("   JWT ID (jti):     %s\n", result.Claims.JWTID)
		}

		if !result.Claims.IssuedAt.IsZero() {
			fmt.Printf("   Issued At (iat):  %s\n",
				result.Claims.IssuedAt.Format("02/01/2006 15:04:05"))
		}
		if !result.Claims.NotBefore.IsZero() {
			fmt.Printf("   Not Before (nbf): %s\n",
				result.Claims.NotBefore.Format("02/01/2006 15:04:05"))
		}
		if !result.Claims.ExpiresAt.IsZero() {
			expiryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			expiryStatus := ""
			if result.Claims.IsExpired {
				expiryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
				expiryStatus = " (EXPIRADO)"
			} else if result.Claims.ExpiresIn != "" {
				expiryStatus = fmt.Sprintf(" (expira em %s)", result.Claims.ExpiresIn)
			}
			fmt.Printf("   Expires At (exp): %s%s\n",
				expiryStyle.Render(result.Claims.ExpiresAt.Format("02/01/2006 15:04:05")),
				lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(expiryStatus))
		}

		fmt.Println()
	}

	// Signature (truncated)
	if result.Signature != "" {
		sig := result.Signature
		if len(sig) > 40 {
			sig = sig[:37] + "..."
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("   Signature: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(sig))
		fmt.Println()
	}
}

func displayJSONMap(m map[string]interface{}, prefix string) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			fmt.Printf("%s%s: %s\n",
				prefix,
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(k),
				val)
		case float64:
			fmt.Printf("%s%s: %v\n",
				prefix,
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(k),
				val)
		case bool:
			fmt.Printf("%s%s: %v\n",
				prefix,
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(k),
				val)
		default:
			jsonBytes, _ := json.Marshal(val)
			fmt.Printf("%s%s: %s\n",
				prefix,
				lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(k),
				string(jsonBytes))
		}
	}
}
