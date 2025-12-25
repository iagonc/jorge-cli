package mtr

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// MTRUsecase handles network path analysis
type MTRUsecase struct {
	logger *zap.Logger
}

// NewMTRUsecase creates a new MTRUsecase
func NewMTRUsecase(logger *zap.Logger) *MTRUsecase {
	return &MTRUsecase{logger: logger}
}

// RunMTR performs MTR-like network path analysis
func (u *MTRUsecase) RunMTR(ctx context.Context, target string, count int) (*models.MTRResult, error) {
	startTime := time.Now()

	if count <= 0 {
		count = 10
	}

	result := &models.MTRResult{
		Target:      target,
		PacketsSent: count,
		Hops:        []models.MTRHop{},
	}

	// Resolve target IP
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}
	if len(ips) > 0 {
		result.TargetIP = ips[0].String()
	}

	// Try system mtr/traceroute first
	hops, err := u.runSystemMTR(ctx, target, count)
	if err != nil {
		// Fallback to simple traceroute
		hops, err = u.runSimpleTrace(ctx, target, count)
		if err != nil {
			return nil, err
		}
	}

	result.Hops = hops
	result.TotalHops = len(hops)
	result.Duration = time.Since(startTime)

	// Check if destination was reached
	for _, hop := range hops {
		if hop.IP == result.TargetIP {
			result.Completed = true
			break
		}
	}

	// Generate summary
	u.generateSummary(result)

	return result, nil
}

func (u *MTRUsecase) runSystemMTR(ctx context.Context, target string, count int) ([]models.MTRHop, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Try mtr first
		if _, err := exec.LookPath("mtr"); err == nil {
			cmd = exec.CommandContext(ctx, "mtr", "-r", "-c", strconv.Itoa(count), "-n", target)
		} else {
			return nil, fmt.Errorf("mtr not found")
		}
	} else {
		return nil, fmt.Errorf("mtr not supported on %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return u.parseMTROutput(string(output))
}

func (u *MTRUsecase) parseMTROutput(output string) ([]models.MTRHop, error) {
	var hops []models.MTRHop
	lines := strings.Split(output, "\n")

	// Regex to parse mtr output
	// Example: " 1.|-- 192.168.1.1   0.0%    10    1.2   1.1   1.0   1.5   0.2"
	re := regexp.MustCompile(`\s*(\d+)\.\|[-]+\s+(\S+)\s+([\d.]+)%\s+(\d+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 10 {
			hopNum, _ := strconv.Atoi(matches[1])
			loss, _ := strconv.ParseFloat(matches[3], 64)
			sent, _ := strconv.Atoi(matches[4])
			last, _ := strconv.ParseFloat(matches[5], 64)
			avg, _ := strconv.ParseFloat(matches[6], 64)
			best, _ := strconv.ParseFloat(matches[7], 64)
			worst, _ := strconv.ParseFloat(matches[8], 64)
			stdDev, _ := strconv.ParseFloat(matches[9], 64)

			hop := models.MTRHop{
				Number:      hopNum,
				Host:        matches[2],
				IP:          matches[2],
				Loss:        loss,
				Sent:        sent,
				Received:    int(float64(sent) * (1 - loss/100)),
				AvgLatency:  time.Duration(avg * float64(time.Millisecond)),
				MinLatency:  time.Duration(best * float64(time.Millisecond)),
				MaxLatency:  time.Duration(worst * float64(time.Millisecond)),
				StdDev:      time.Duration(stdDev * float64(time.Millisecond)),
				LastLatency: time.Duration(last * float64(time.Millisecond)),
			}

			// Determine status
			if matches[2] == "???" {
				hop.Status = models.HopStatusTimeout
			} else if loss > 0 {
				hop.Status = models.HopStatusLoss
			} else if avg > 100 {
				hop.Status = models.HopStatusHighLat
			} else {
				hop.Status = models.HopStatusOK
			}

			hops = append(hops, hop)
		}
	}

	return hops, nil
}

func (u *MTRUsecase) runSimpleTrace(ctx context.Context, target string, count int) ([]models.MTRHop, error) {
	var hops []models.MTRHop
	maxHops := 30

	for ttl := 1; ttl <= maxHops; ttl++ {
		hop := models.MTRHop{
			Number: ttl,
			Sent:   count,
		}

		var latencies []time.Duration
		var received int
		var hopIP string

		for i := 0; i < count; i++ {
			select {
			case <-ctx.Done():
				return hops, ctx.Err()
			default:
			}

			ip, latency, err := u.probeHop(target, ttl, 2*time.Second)
			if err == nil {
				latencies = append(latencies, latency)
				received++
				hopIP = ip
			}
		}

		if hopIP != "" {
			hop.IP = hopIP
			hop.Host = hopIP
			// Try reverse DNS
			names, err := net.LookupAddr(hopIP)
			if err == nil && len(names) > 0 {
				hop.Host = strings.TrimSuffix(names[0], ".")
			}
		} else {
			hop.Host = "???"
			hop.Status = models.HopStatusTimeout
		}

		hop.Received = received
		hop.Loss = float64(count-received) / float64(count) * 100

		if len(latencies) > 0 {
			var total time.Duration
			hop.MinLatency = latencies[0]
			hop.MaxLatency = latencies[0]
			for _, l := range latencies {
				total += l
				if l < hop.MinLatency {
					hop.MinLatency = l
				}
				if l > hop.MaxLatency {
					hop.MaxLatency = l
				}
			}
			hop.AvgLatency = total / time.Duration(len(latencies))
			hop.LastLatency = latencies[len(latencies)-1]

			if hop.Loss > 0 {
				hop.Status = models.HopStatusLoss
			} else if hop.AvgLatency > 100*time.Millisecond {
				hop.Status = models.HopStatusHighLat
			} else {
				hop.Status = models.HopStatusOK
			}
		}

		hops = append(hops, hop)

		// Check if we reached destination
		if hopIP == target {
			break
		}
		ips, _ := net.LookupIP(target)
		for _, ip := range ips {
			if ip.String() == hopIP {
				return hops, nil
			}
		}
	}

	return hops, nil
}

func (u *MTRUsecase) probeHop(target string, ttl int, timeout time.Duration) (string, time.Duration, error) {
	// Use system ping with TTL
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("ping", "-c", "1", "-t", strconv.Itoa(ttl), "-W", "1", target)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("ping", "-c", "1", "-m", strconv.Itoa(ttl), "-W", "1000", target)
	} else {
		return "", 0, fmt.Errorf("unsupported OS")
	}

	start := time.Now()
	output, _ := cmd.CombinedOutput()
	latency := time.Since(start)

	// Parse output to find the responding IP
	outputStr := string(output)

	// Look for "From X.X.X.X" or "from X.X.X.X"
	re := regexp.MustCompile(`[Ff]rom\s+([\d.]+)`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) >= 2 {
		return matches[1], latency, nil
	}

	// Check if we got a reply (destination reached)
	if strings.Contains(outputStr, "bytes from") || strings.Contains(outputStr, "64 bytes") {
		re := regexp.MustCompile(`from\s+([\d.]+)`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) >= 2 {
			return matches[1], latency, nil
		}
	}

	return "", 0, fmt.Errorf("no response")
}

func (u *MTRUsecase) generateSummary(result *models.MTRResult) {
	summary := models.MTRSummary{
		DestinationReached: result.Completed,
	}

	var worstLoss float64
	var worstLatency time.Duration

	for _, hop := range result.Hops {
		summary.TotalLatency += hop.AvgLatency

		if hop.Loss > worstLoss {
			worstLoss = hop.Loss
			summary.WorstHop = hop.Number
			summary.WorstHopHost = hop.Host
			summary.WorstHopLoss = hop.Loss
		}

		if hop.AvgLatency > worstLatency && hop.Status != models.HopStatusTimeout {
			worstLatency = hop.AvgLatency
			if summary.WorstHopLoss == 0 { // Only if no loss issues
				summary.WorstHop = hop.Number
				summary.WorstHopHost = hop.Host
			}
			summary.WorstHopLatency = hop.AvgLatency
		}
	}

	// Generate recommendation
	if worstLoss > 10 {
		summary.Bottleneck = fmt.Sprintf("Hop %d (%s) com %.1f%% de perda", summary.WorstHop, summary.WorstHopHost, worstLoss)
		summary.Recommendation = "Alta perda de pacotes detectada. Pode indicar congestionamento ou problema no roteador."
	} else if worstLatency > 150*time.Millisecond {
		summary.Bottleneck = fmt.Sprintf("Hop %d (%s) com %v de latência", summary.WorstHop, summary.WorstHopHost, worstLatency.Round(time.Millisecond))
		summary.Recommendation = "Latência alta detectada. Pode ser distância geográfica ou congestionamento."
	} else if !result.Completed {
		summary.Recommendation = "Destino não alcançado. Pode haver firewall bloqueando ou rota inexistente."
	} else {
		summary.Recommendation = "Rota de rede saudável."
	}

	result.Summary = summary
}
