package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/jsonq"
	"github.com/spf13/cobra"
)

// NewJSONCommand creates the json command
func NewJSONCommand(usecase *jsonq.JSONUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json [query] [file]",
		Short: "Query e manipula JSON (como jq)",
		Long: `Query e manipula dados JSON usando expressões de caminho simples.

Expressões suportadas:
  .           Documento inteiro
  .field      Acessa campo
  .a.b        Acessa campos aninhados
  .[0]        Acessa índice de array
  .items[0]   Combina campos e índices`,
		Example: `  # Query de arquivo
  jorge json '.name' data.json
  jorge json '.users[0].email' data.json

  # Query de stdin
  echo '{"name":"test"}' | jorge json '.name'
  curl -s api.com/data | jorge json '.items[0]'

  # Formatar JSON
  jorge json format data.json
  echo '{"a":1}' | jorge json format

  # Validar JSON
  jorge json validate data.json`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("especifique uma query ou subcomando")
			}

			query := args[0]
			var jsonStr string

			if len(args) > 1 {
				// Read from file
				content, err := os.ReadFile(args[1])
				if err != nil {
					return err
				}
				jsonStr = string(content)
			} else {
				// Read from stdin
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				jsonStr = string(content)
			}

			result := usecase.Query(jsonStr, query)

			if result.Error != "" {
				fmt.Printf("\n   %s %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
					result.Error)
				return nil
			}

			// Output result
			if str, ok := result.Result.(string); ok {
				fmt.Println(str)
			} else {
				out, _ := json.MarshalIndent(result.Result, "", "  ")
				fmt.Println(string(out))
			}

			return nil
		},
	}

	// Format subcommand
	formatCmd := &cobra.Command{
		Use:   "format [file]",
		Short: "Formata JSON com indentação",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonStr string
			if len(args) > 0 {
				content, err := os.ReadFile(args[0])
				if err != nil {
					return err
				}
				jsonStr = string(content)
			} else {
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				jsonStr = string(content)
			}

			formatted, err := usecase.Format(jsonStr, 2)
			if err != nil {
				return err
			}
			fmt.Println(formatted)
			return nil
		},
	}

	// Validate subcommand
	validateCmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Valida JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonStr string
			if len(args) > 0 {
				content, err := os.ReadFile(args[0])
				if err != nil {
					return err
				}
				jsonStr = string(content)
			} else {
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				jsonStr = string(content)
			}

			valid, errMsg := usecase.Validate(strings.TrimSpace(jsonStr))
			fmt.Println()
			if valid {
				fmt.Printf("   %s JSON válido\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
			} else {
				fmt.Printf("   %s JSON inválido: %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
					errMsg)
			}
			return nil
		},
	}

	// Minify subcommand
	minifyCmd := &cobra.Command{
		Use:   "minify [file]",
		Short: "Remove espaços do JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonStr string
			if len(args) > 0 {
				content, err := os.ReadFile(args[0])
				if err != nil {
					return err
				}
				jsonStr = string(content)
			} else {
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				jsonStr = string(content)
			}

			minified, err := usecase.Minify(jsonStr)
			if err != nil {
				return err
			}
			fmt.Println(minified)
			return nil
		},
	}

	// Keys subcommand
	keysCmd := &cobra.Command{
		Use:   "keys [file]",
		Short: "Lista chaves do objeto JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonStr string
			if len(args) > 0 {
				content, err := os.ReadFile(args[0])
				if err != nil {
					return err
				}
				jsonStr = string(content)
			} else {
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				jsonStr = string(content)
			}

			keys, err := usecase.Keys(jsonStr)
			if err != nil {
				return err
			}

			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}

	cmd.AddCommand(formatCmd, validateCmd, minifyCmd, keysCmd)

	return cmd
}
