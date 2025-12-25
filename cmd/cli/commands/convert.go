package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/convert"
	"github.com/spf13/cobra"
)

// NewConvertCommand creates the convert command
func NewConvertCommand(usecase *convert.ConvertUsecase) *cobra.Command {
	var outputFormat string
	var inputFormat string

	cmd := &cobra.Command{
		Use:   "convert <file>",
		Short: "Converte entre JSON, YAML e ENV",
		Long: `Converte arquivos entre diferentes formatos de configuração.

Formatos suportados:
  - json: JavaScript Object Notation
  - yaml: YAML Ain't Markup Language
  - env:  Environment variables (KEY=VALUE)`,
		Example: `  # Converter YAML para JSON
  jorge convert config.yaml --to json

  # Converter JSON para YAML
  jorge convert data.json --to yaml

  # Converter para ENV
  jorge convert config.yaml --to env

  # Especificar formato de entrada
  jorge convert data.txt --from json --to yaml

  # De stdin
  echo '{"name":"test"}' | jorge convert --from json --to yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string

			if len(args) > 0 {
				// Read from file
				result := usecase.ConvertFile(args[0], models.ConvertFormat(outputFormat))
				displayConvertResult(result)
				return nil
			}

			// Read from stdin
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			content = string(bytes)

			if inputFormat == "" {
				return fmt.Errorf("especifique --from quando usar stdin")
			}

			result := usecase.Convert(content, models.ConvertFormat(inputFormat), models.ConvertFormat(outputFormat))
			displayConvertResult(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFormat, "to", "json", "Formato de saída (json, yaml, env)")
	cmd.Flags().StringVar(&inputFormat, "from", "", "Formato de entrada (auto-detectado de arquivos)")

	return cmd
}

func displayConvertResult(result *models.ConvertResult) {
	if result.Error != "" {
		fmt.Printf("\n   %s %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗"),
			result.Error)
		return
	}

	fmt.Print(result.Output)
	if result.Output != "" && result.Output[len(result.Output)-1] != '\n' {
		fmt.Println()
	}
}
