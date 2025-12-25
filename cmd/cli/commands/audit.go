package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/audit"
	"github.com/spf13/cobra"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	scoreBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(1, 3).
			Align(lipgloss.Center)

	passStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// NewAuditCommand creates the audit command
func NewAuditCommand(usecase *audit.AuditUsecase) *cobra.Command {
	var skipMTR bool
	var skipPorts bool
	var outputFormat string
	var saveToFile string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "audit <url>",
		Short: "Auditoria completa com relatório detalhado",
		Long: `Executa uma auditoria completa de um endpoint e gera um relatório detalhado.

A auditoria inclui:
  ✓ Teste de conectividade (DNS, TCP, HTTP)
  ✓ Análise SSL/TLS (certificado, versão, cipher)
  ✓ Headers de segurança (HSTS, CSP, X-Frame-Options, etc)
  ✓ Performance (timing de cada fase)
  ✓ Análise de rota de rede (opcional)
  ✓ Scan de portas comuns (opcional)

Gera um relatório com:
  • Score geral (0-100) e nota (A+ a F)
  • Problemas encontrados categorizados por severidade
  • Recomendações de melhoria

Formatos de saída: terminal, json, markdown`,
		Example: `  # Auditoria completa
  jorge audit https://api.exemplo.com

  # Auditoria rápida (sem MTR e port scan)
  jorge audit https://api.exemplo.com --skip-mtr --skip-ports

  # Salvar relatório em JSON
  jorge audit https://api.exemplo.com -o json --save report.json

  # Salvar relatório em Markdown
  jorge audit https://api.exemplo.com -o md --save report.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			config := models.AuditConfig{
				Target:       target,
				Timeout:      timeout,
				SkipMTR:      skipMTR,
				SkipPorts:    skipPorts,
				OutputFormat: outputFormat,
				SaveToFile:   saveToFile,
			}

			// Print header
			fmt.Println()
			printAuditHeader(target)

			// Run audit with progress
			ctx := cmd.Context()
			report, err := usecase.RunAudit(ctx, config, func(phase string, progress int) {
				printProgress(phase, progress)
			})

			if err != nil {
				return err
			}

			// Clear progress line
			fmt.Print("\r\033[K")

			// Display results based on format
			if outputFormat == "json" {
				jsonBytes, err := usecase.ExportJSON(report)
				if err != nil {
					return err
				}
				fmt.Println(string(jsonBytes))
			} else if outputFormat == "md" || outputFormat == "markdown" {
				fmt.Println(usecase.ExportMarkdown(report))
			} else {
				displayAuditReport(report)
			}

			// Save to file if requested
			if saveToFile != "" {
				format := outputFormat
				if format == "" || format == "terminal" {
					format = "md"
				}
				if err := usecase.SaveToFile(report, saveToFile, format); err != nil {
					return fmt.Errorf("erro ao salvar arquivo: %w", err)
				}
				fmt.Printf("\n   📄 Relatório salvo em: %s\n\n", saveToFile)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipMTR, "skip-mtr", false, "Pular análise de rota de rede")
	cmd.Flags().BoolVar(&skipPorts, "skip-ports", false, "Pular scan de portas")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "terminal", "Formato de saída (terminal, json, md)")
	cmd.Flags().StringVar(&saveToFile, "save", "", "Salvar relatório em arquivo")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Timeout para cada teste")

	return cmd
}

func printAuditHeader(target string) {
	fmt.Println(titleStyle.Render("📋 AUDITORIA COMPLETA"))
	fmt.Printf("   Alvo: %s\n", target)
	fmt.Println()
}

func printProgress(phase string, progress int) {
	bar := ""
	filled := progress / 5 // 20 chars total
	for i := 0; i < 20; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	fmt.Printf("\r   [%s] %3d%% %s", bar, progress, phase)
}

func displayAuditReport(report *models.AuditReport) {
	fmt.Println()

	// Overall Score Box
	printScoreBox(report)

	// Sections
	if report.Connectivity != nil {
		printConnectivitySection(report.Connectivity)
	}

	if report.SSL != nil {
		printSSLSection(report.SSL)
	}

	if report.SecurityHeaders != nil {
		printSecurityHeadersSection(report.SecurityHeaders)
	}

	if report.Performance != nil {
		printPerformanceSection(report.Performance)
	}

	if report.NetworkPath != nil {
		printNetworkPathSection(report.NetworkPath)
	}

	if report.PortScan != nil {
		printPortScanSection(report.PortScan)
	}

	// Issues and Warnings
	if len(report.Issues) > 0 {
		printIssuesSection(report.Issues)
	}

	if len(report.Warnings) > 0 {
		printWarningsSection(report.Warnings)
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		printRecommendationsSection(report.Recommendations)
	}

	// Footer
	printReportFooter(report)
}

func printScoreBox(report *models.AuditReport) {
	// Grade color
	var gradeStyle lipgloss.Style
	switch {
	case report.OverallScore >= 80:
		gradeStyle = scoreBoxStyle.Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0"))
	case report.OverallScore >= 60:
		gradeStyle = scoreBoxStyle.Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0"))
	case report.OverallScore >= 40:
		gradeStyle = scoreBoxStyle.Background(lipgloss.Color("208")).Foreground(lipgloss.Color("0"))
	default:
		gradeStyle = scoreBoxStyle.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("255"))
	}

	// Status icon and text
	var statusIcon, statusText string
	var statusStyle lipgloss.Style
	switch report.OverallStatus {
	case models.AuditStatusHealthy:
		statusIcon = "✓"
		statusText = "SAUDÁVEL"
		statusStyle = passStyle
	case models.AuditStatusWarning:
		statusIcon = "!"
		statusText = "ATENÇÃO"
		statusStyle = warnStyle
	case models.AuditStatusCritical:
		statusIcon = "✗"
		statusText = "CRÍTICO"
		statusStyle = failStyle
	default:
		statusIcon = "✗"
		statusText = "OFFLINE"
		statusStyle = failStyle
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()

	// Center the score box
	gradeBox := gradeStyle.Render(fmt.Sprintf(" %s ", report.OverallGrade))
	scoreText := fmt.Sprintf("%d/100", report.OverallScore)
	statusLabel := statusStyle.Bold(true).Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	fmt.Printf("   Nota: %s  Score: %s  Status: %s\n",
		gradeBox,
		lipgloss.NewStyle().Bold(true).Render(scoreText),
		statusLabel)

	fmt.Println()
	fmt.Println(strings.Repeat("═", 70))
}

func printConnectivitySection(conn *models.AuditConnectivity) {
	fmt.Println()
	fmt.Printf("   %s\n", sectionStyle.Render(" 🔌 CONECTIVIDADE "))
	fmt.Println()

	// DNS
	dnsIcon := passStyle.Render("✓")
	if !conn.DNSResolves {
		dnsIcon = failStyle.Render("✗")
	}
	fmt.Printf("   %s DNS Resolution      %s\n", dnsIcon, mutedStyle.Render(conn.DNSTime.Round(time.Millisecond).String()))
	if len(conn.IPs) > 0 {
		fmt.Printf("      IPs: %s\n", mutedStyle.Render(strings.Join(conn.IPs, ", ")))
	}

	// TCP
	tcpIcon := passStyle.Render("✓")
	if !conn.TCPConnects {
		tcpIcon = failStyle.Render("✗")
	}
	fmt.Printf("   %s TCP Connection      %s\n", tcpIcon, mutedStyle.Render(conn.TCPTime.Round(time.Millisecond).String()))

	// HTTP
	httpIcon := passStyle.Render("✓")
	if !conn.HTTPWorks {
		httpIcon = failStyle.Render("✗")
	}
	httpStatus := fmt.Sprintf("HTTP %d", conn.HTTPStatus)
	if conn.HTTPStatus >= 400 {
		httpStatus = failStyle.Render(httpStatus)
	} else if conn.HTTPStatus >= 300 {
		httpStatus = warnStyle.Render(httpStatus)
	} else {
		httpStatus = passStyle.Render(httpStatus)
	}
	fmt.Printf("   %s HTTP Request        %s  %s\n", httpIcon, mutedStyle.Render(conn.HTTPTime.Round(time.Millisecond).String()), httpStatus)

	fmt.Printf("\n   Score: %d/100\n", conn.Score)
}

func printSSLSection(ssl *models.AuditSSL) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", sectionStyle.Render(" 🔒 SSL/TLS "))
	fmt.Println()

	// Valid
	validIcon := passStyle.Render("✓")
	validText := passStyle.Render("Válido")
	if !ssl.Valid {
		validIcon = failStyle.Render("✗")
		validText = failStyle.Render("Inválido")
	}
	fmt.Printf("   %s Certificado         %s\n", validIcon, validText)

	// Version
	versionStyle := passStyle
	if ssl.Version == "TLS 1.0" || ssl.Version == "TLS 1.1" {
		versionStyle = failStyle
	} else if ssl.Version == "TLS 1.2" {
		versionStyle = warnStyle
	}
	fmt.Printf("   • Versão:            %s\n", versionStyle.Render(ssl.Version))

	// Cipher
	fmt.Printf("   • Cipher:            %s\n", mutedStyle.Render(ssl.CipherSuite))

	// Certificate info
	fmt.Printf("   • Certificado:       %s\n", ssl.Certificate)
	fmt.Printf("   • Emissor:           %s\n", mutedStyle.Render(ssl.Issuer))

	// Expiry
	expiryStyle := passStyle
	if ssl.DaysUntilExpiry < 30 {
		expiryStyle = failStyle
	} else if ssl.DaysUntilExpiry < 90 {
		expiryStyle = warnStyle
	}
	fmt.Printf("   • Expira em:         %s (%s dias)\n",
		ssl.ExpiresAt.Format("02/01/2006"),
		expiryStyle.Render(fmt.Sprintf("%d", ssl.DaysUntilExpiry)))

	// Errors
	if len(ssl.Errors) > 0 {
		fmt.Printf("\n   %s\n", failStyle.Render("Erros:"))
		for _, e := range ssl.Errors {
			fmt.Printf("      • %s\n", failStyle.Render(e))
		}
	}

	fmt.Printf("\n   Score: %d/100\n", ssl.Score)
}

func printSecurityHeadersSection(headers *models.AuditSecurityHeaders) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", sectionStyle.Render(" 🛡️  HEADERS DE SEGURANÇA "))
	fmt.Println()

	// Grade
	gradeStyle := passStyle
	if headers.Score < 60 {
		gradeStyle = failStyle
	} else if headers.Score < 80 {
		gradeStyle = warnStyle
	}
	fmt.Printf("   Nota: %s  Score: %d/100\n\n", gradeStyle.Bold(true).Render(headers.Grade), headers.Score)

	// Headers table
	for _, h := range headers.Headers {
		icon := failStyle.Render("✗")
		status := failStyle.Render("ausente")
		if h.Present {
			icon = passStyle.Render("✓")
			status = passStyle.Render("presente")
		}
		fmt.Printf("   %s %-24s %s\n", icon, h.Name, status)
	}
}

func printPerformanceSection(perf *models.AuditPerformance) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", sectionStyle.Render(" ⚡ PERFORMANCE "))
	fmt.Println()

	// Timing breakdown
	fmt.Printf("   DNS Lookup:          %s\n", mutedStyle.Render(perf.DNSLookup.Round(time.Millisecond).String()))
	fmt.Printf("   TCP Connect:         %s\n", mutedStyle.Render(perf.TCPConnect.Round(time.Millisecond).String()))
	fmt.Printf("   TLS Handshake:       %s\n", mutedStyle.Render(perf.TLSHandshake.Round(time.Millisecond).String()))
	fmt.Printf("   Server Processing:   %s\n", mutedStyle.Render(perf.ServerProcessing.Round(time.Millisecond).String()))

	// Total and rating
	ratingStyle := passStyle
	if perf.Score < 60 {
		ratingStyle = failStyle
	} else if perf.Score < 80 {
		ratingStyle = warnStyle
	}

	fmt.Println()
	fmt.Printf("   Total:               %s\n", lipgloss.NewStyle().Bold(true).Render(perf.TotalTime.Round(time.Millisecond).String()))
	fmt.Printf("   TTFB:                %s\n", mutedStyle.Render(perf.TTFB.Round(time.Millisecond).String()))
	fmt.Printf("   Rating:              %s\n", ratingStyle.Render(perf.Rating))

	fmt.Printf("\n   Score: %d/100\n", perf.Score)
}

func printNetworkPathSection(path *models.AuditNetworkPath) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", sectionStyle.Render(" 🛤️  ROTA DE REDE "))
	fmt.Println()

	if path.Completed {
		fmt.Printf("   %s Destino alcançado em %d hops\n", passStyle.Render("✓"), path.TotalHops)
		fmt.Printf("   Latência total: %s\n", mutedStyle.Render(path.TotalLatency.Round(time.Millisecond).String()))
	} else {
		fmt.Printf("   %s Destino não alcançado\n", failStyle.Render("✗"))
	}

	if path.Bottleneck != "" {
		fmt.Printf("   %s Gargalo: %s\n", warnStyle.Render("⚠"), path.Bottleneck)
	}

	fmt.Printf("\n   Score: %d/100\n", path.Score)
}

func printPortScanSection(ports *models.AuditPortScan) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", sectionStyle.Render(" 🔍 PORTAS "))
	fmt.Println()

	fmt.Printf("   Portas escaneadas: %d\n", ports.TotalScanned)
	fmt.Printf("   Portas abertas: %d\n\n", len(ports.OpenPorts))

	for _, p := range ports.OpenPorts {
		fmt.Printf("   %s %d/%s (%s)\n",
			passStyle.Render("●"),
			p.Port,
			"tcp",
			mutedStyle.Render(p.Service))
	}

	fmt.Printf("\n   Score: %d/100\n", ports.Score)
}

func printIssuesSection(issues []models.AuditIssue) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("⚠️  PROBLEMAS ENCONTRADOS"))
	fmt.Println()

	for _, issue := range issues {
		severityStyle := warnStyle
		if issue.Severity == "critical" || issue.Severity == "high" {
			severityStyle = failStyle
		}

		fmt.Printf("   %s [%s] %s\n",
			severityStyle.Render("●"),
			strings.ToUpper(issue.Severity),
			lipgloss.NewStyle().Bold(true).Render(issue.Title))
		fmt.Printf("      %s\n", mutedStyle.Render(issue.Description))
		if issue.Solution != "" {
			fmt.Printf("      %s %s\n", infoStyle.Render("→"), issue.Solution)
		}
		fmt.Println()
	}
}

func printWarningsSection(warnings []models.AuditWarning) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", warnStyle.Bold(true).Render("⚡ AVISOS"))
	fmt.Println()

	for _, w := range warnings {
		fmt.Printf("   %s [%s] %s\n",
			warnStyle.Render("●"),
			w.Category,
			w.Message)
		if w.Suggestion != "" {
			fmt.Printf("      %s %s\n", infoStyle.Render("→"), mutedStyle.Render(w.Suggestion))
		}
	}
}

func printRecommendationsSection(recommendations []string) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("   %s\n", infoStyle.Bold(true).Render("💡 RECOMENDAÇÕES"))
	fmt.Println()

	for i, rec := range recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}
}

func printReportFooter(report *models.AuditReport) {
	fmt.Println()
	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("   Auditoria concluída em %s\n", report.Duration.Round(time.Millisecond))
	fmt.Printf("   %s\n", mutedStyle.Render(report.Timestamp.Format("02/01/2006 15:04:05")))
	fmt.Println()
}
