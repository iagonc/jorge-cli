package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/ipinfo"
	"github.com/spf13/cobra"
)

// NewIPCommand creates the ip command
func NewIPCommand(usecase *ipinfo.IPUsecase) *cobra.Command {
	var scanPorts bool

	cmd := &cobra.Command{
		Use:   "ip <ip-or-hostname>",
		Short: "Informacoes detalhadas sobre um IP",
		Long: `Obtem informacoes detalhadas sobre um endereco IP ou hostname.

Informacoes incluem:
  - Geolocalizacao (pais, cidade, timezone)
  - ASN e provedor
  - DNS reverso
  - Scan de portas comuns (opcional)`,
		Example: `  # Info basica de um IP
  jorge ip 8.8.8.8

  # Info de um hostname
  jorge ip google.com

  # Com scan de portas
  jorge ip 1.1.1.1 --ports

  # Meu IP publico
  jorge ip --my`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			myIP, _ := cmd.Flags().GetBool("my")

			var ipStr string
			if myIP {
				ip, err := usecase.MyIP(ctx)
				if err != nil {
					return err
				}
				ipStr = ip
				fmt.Println()
				fmt.Printf("   Seu IP publico: %s\n\n",
					lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render(ip))
			} else if len(args) == 0 {
				return fmt.Errorf("IP ou hostname requerido (ou use --my)")
			} else {
				ipStr = args[0]
			}

			var info *models.IPInfo
			var err error

			if scanPorts {
				info, err = usecase.LookupWithPorts(ctx, ipStr)
			} else {
				info, err = usecase.Lookup(ctx, ipStr)
			}

			if err != nil {
				return err
			}

			displayIPInfo(info)
			return nil
		},
	}

	cmd.Flags().BoolVar(&scanPorts, "ports", false, "Escanear portas comuns")
	cmd.Flags().Bool("my", false, "Mostrar meu IP publico")

	return cmd
}

func displayIPInfo(info *models.IPInfo) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("🌐 IP Info"))
	fmt.Println()

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(15)
	valueStyle := lipgloss.NewStyle().Bold(true)

	// Basic info
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("   %s %s\n", labelStyle.Render("IP:"), valueStyle.Render(info.IP))
	fmt.Printf("   %s %s\n", labelStyle.Render("Version:"), info.Version)

	if info.IsPrivate {
		fmt.Printf("   %s %s\n", labelStyle.Render("Type:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Private"))
	} else if info.IsLoopback {
		fmt.Printf("   %s %s\n", labelStyle.Render("Type:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Loopback"))
	} else {
		fmt.Printf("   %s %s\n", labelStyle.Render("Type:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("Public"))
	}

	if info.Hostname != "" {
		fmt.Printf("   %s %s\n", labelStyle.Render("Hostname:"), info.Hostname)
	}
	fmt.Println(strings.Repeat("─", 60))

	// Geo info
	if info.Geo != nil {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📍 Geolocation"))
		fmt.Printf("   %s %s (%s)\n", labelStyle.Render("Country:"),
			info.Geo.Country, info.Geo.CountryCode)
		if info.Geo.Region != "" {
			fmt.Printf("   %s %s\n", labelStyle.Render("Region:"), info.Geo.Region)
		}
		if info.Geo.City != "" {
			fmt.Printf("   %s %s\n", labelStyle.Render("City:"), info.Geo.City)
		}
		fmt.Printf("   %s %.4f, %.4f\n", labelStyle.Render("Coordinates:"),
			info.Geo.Latitude, info.Geo.Longitude)
		if info.Geo.Timezone != "" {
			fmt.Printf("   %s %s\n", labelStyle.Render("Timezone:"), info.Geo.Timezone)
		}
	}

	// ASN info
	if info.ASN != nil {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   🏢 Network"))
		if info.ASN.Number > 0 {
			fmt.Printf("   %s AS%d\n", labelStyle.Render("ASN:"), info.ASN.Number)
		}
		if info.ASN.Organization != "" {
			fmt.Printf("   %s %s\n", labelStyle.Render("Organization:"), info.ASN.Organization)
		}
		if info.ASN.ISP != "" {
			fmt.Printf("   %s %s\n", labelStyle.Render("ISP:"), info.ASN.ISP)
		}
	}

	// DNS info
	if info.DNS != nil && len(info.DNS.PTR) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   📋 DNS"))
		for _, ptr := range info.DNS.PTR {
			fmt.Printf("   %s %s\n", labelStyle.Render("PTR:"), ptr)
		}
	}

	// Port scan results
	if len(info.Ports) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("   🔌 Ports"))

		// Sort by port number
		sort.Slice(info.Ports, func(i, j int) bool {
			return info.Ports[i].Port < info.Ports[j].Port
		})

		for _, p := range info.Ports {
			status := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("closed")
			if p.Open {
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("open")
			}
			fmt.Printf("   %s %-5d %s (%s)\n", labelStyle.Render("Port:"),
				p.Port, status, p.Service)
		}
	}

	fmt.Println()
}
