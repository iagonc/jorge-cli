package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/encode"
	"github.com/spf13/cobra"
)

// NewEncodeCommand creates the encode command
func NewEncodeCommand(usecase *encode.EncodeUsecase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode <operation> [input]",
		Short: "Encode, decode e hash de strings",
		Long: `Utilitário para encoding, decoding e hashing.

Operações disponíveis:
  Encoding:
    base64         Encode para Base64
    base64-decode  Decode de Base64
    url            URL encode
    url-decode     URL decode
    hex            Encode para hexadecimal
    hex-decode     Decode de hexadecimal

  Hashing:
    md5            Hash MD5
    sha1           Hash SHA1
    sha256         Hash SHA256
    sha512         Hash SHA512
    bcrypt         Hash bcrypt (para senhas)`,
		Example: `  # Encoding
  jorge encode base64 "hello world"
  jorge encode base64-decode "aGVsbG8gd29ybGQ="
  jorge encode url "hello world"
  jorge encode hex "hello"

  # Hashing
  jorge encode md5 "password"
  jorge encode sha256 "secret"
  jorge encode bcrypt "mypassword"

  # Ler de stdin
  echo "hello" | jorge encode base64
  cat file.txt | jorge encode sha256`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			operation := models.EncodeOperation(args[0])

			var input string
			if len(args) > 1 {
				input = strings.Join(args[1:], " ")
			} else {
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					bytes, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					input = strings.TrimSpace(string(bytes))
				} else {
					return fmt.Errorf("especifique o input ou use stdin")
				}
			}

			result := usecase.Encode(operation, input)
			displayEncodeResult(result)
			return nil
		},
	}

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista operações disponíveis",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("📋 Operações Disponíveis"))
			fmt.Println()

			ops := usecase.GetAvailableOperations()
			for _, op := range ops {
				fmt.Printf("   • %s\n", op)
			}
			fmt.Println()
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}

func displayEncodeResult(result *models.EncodeResult) {
	fmt.Println()

	if result.Error != "" {
		fmt.Printf("   %s %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Error)
		return
	}

	// Operation
	fmt.Printf("   %s: %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Operação"),
		lipgloss.NewStyle().Bold(true).Render(string(result.Operation)))

	// Input (truncated if too long)
	inputDisplay := result.Input
	if len(inputDisplay) > 50 {
		inputDisplay = inputDisplay[:47] + "..."
	}
	fmt.Printf("   %s: %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Input"),
		inputDisplay)

	// Output
	fmt.Println()
	fmt.Printf("   %s\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("Resultado:"))
	fmt.Printf("   %s\n", result.Output)
	fmt.Println()
}
