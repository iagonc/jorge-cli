package utils

// import (
// 	"fmt"

// 	"github.com/charmbracelet/lipgloss"
// 	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
// )

// func FormatAndDisplayNetworkDebugResult(result *models.NetworkDebugResult) {
//     // Definir estilos usando Lipgloss
//     titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
//     listStyle := lipgloss.NewStyle().PaddingLeft(2)

//     // DNS Lookup
//     fmt.Println(titleStyle.Render("✨ Verificação de DNS (dig):"))
//     fmt.Printf("- O domínio %s está associado aos seguintes endereços:\n", result.DNSLookup.IPv4)
//     fmt.Println(listStyle.Render(fmt.Sprintf("- IPv4: %s", result.DNSLookup.IPv4)))
//     fmt.Println(listStyle.Render(fmt.Sprintf("- IPv6: %s", result.DNSLookup.IPv6)))
//     fmt.Println()

//     // NSLookup
//     fmt.Println(titleStyle.Render("🔍 Consulta de Endereço (nslookup):"))
//     fmt.Printf("- O endereço IP de %s é %s\n", result.NSLookup.IP, result.NSLookup.IP)
//     fmt.Println()

//     // Traceroute
//     fmt.Println(titleStyle.Render("🚀 Rota dos Dados (Traceroute):"))
//     fmt.Printf("- Os dados viajaram por %d pontos antes de chegar a %s:\n", len(result.Traceroute.Hops), result.Traceroute.Hops[len(result.Traceroute.Hops)-1].Address)
//     for _, hop := range result.Traceroute.Hops {
//         fmt.Printf("  %d. %s: Resposta em %s\n", hop.HopNumber, hop.Address, hop.ResponseTime)
//     }
//     fmt.Println()

//     // HTTP Request (curl)
//     fmt.Println(titleStyle.Render("📡 Verificação de Site (curl):"))
//     fmt.Printf("- Status do Site: %s\n", result.HTTPRequest.Status)
//     fmt.Printf("- Tempo de Resposta: %s\n", result.HTTPRequest.ResponseTime)
//     fmt.Printf("- Tipo de Conteúdo: %s\n", result.HTTPRequest.ContentType)
//     fmt.Println()

//     // Ping
//     fmt.Println(titleStyle.Render("📈 Teste de Conexão (Ping):"))
//     fmt.Printf("- Pacotes Enviados: %d\n", result.Ping.Sent)
//     fmt.Printf("- Pacotes Recebidos: %d\n", result.Ping.Received)
//     fmt.Printf("- Perda de Pacotes: %.0f%%\n", result.Ping.LossPercent)
//     fmt.Printf("- Tempo Médio de Resposta: %d ms\n", result.Ping.AvgLatency)
//     fmt.Println()

//     // Netstat
//     fmt.Println(titleStyle.Render("🖥️ Conexões Ativas (Netstat):"))
//     if len(result.Netstat.Connections) == 0 {
//         fmt.Println("- Nenhuma conexão ativa encontrada.")
//     } else {
//         fmt.Println("- Conexões Ativas:")
//         for _, conn := range result.Netstat.Connections {
//             fmt.Printf("  - %s %s → %s (%s)\n", conn.Protocol, conn.LocalAddress, conn.RemoteAddress, conn.Status)
//         }
//     }
//     fmt.Println()

//     // Iftop
//     fmt.Println(titleStyle.Render("📊 Uso de Rede Atual (Iftop - Interface: eth0):"))
//     fmt.Printf("- Tráfego Atual:\n")
//     fmt.Printf("  - Enviando: %s\n", result.Iftop.SendingKBps)
//     fmt.Printf("  - Recebendo: %s\n", result.Iftop.ReceivingKBps)
//     fmt.Println("- Top 3 Conexões Mais Ativas:")
//     for i, conn := range result.Iftop.TopConnections {
//         fmt.Printf("  %d. %s ↔ %s: Enviando %s | Recebendo %s\n", i+1, conn.Source, conn.Destination, conn.SentKBps, conn.ReceivedKBps)
//     }
// }
