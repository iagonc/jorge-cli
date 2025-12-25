package portscan

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// PortScanner interface for testability
type PortScanner interface {
	ScanPort(ctx context.Context, host string, port int, timeout time.Duration) models.PortResult
}

// DefaultPortScanner implements PortScanner using TCP connect
type DefaultPortScanner struct{}

func (s *DefaultPortScanner) ScanPort(ctx context.Context, host string, port int, timeout time.Duration) models.PortResult {
	result := models.PortResult{
		Port:    port,
		Service: models.CommonPorts[port],
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{Timeout: timeout}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.Status = models.PortFiltered
		} else {
			result.Status = models.PortClosed
		}
		return result
	}
	defer conn.Close()

	result.Status = models.PortOpen
	return result
}

// PortScanUsecase handles port scanning operations
type PortScanUsecase struct {
	Logger      *zap.Logger
	Scanner     PortScanner
	Concurrency int
	Timeout     time.Duration
}

// NewPortScanUsecase creates a new PortScanUsecase
func NewPortScanUsecase(logger *zap.Logger) *PortScanUsecase {
	return &PortScanUsecase{
		Logger:      logger,
		Scanner:     &DefaultPortScanner{},
		Concurrency: 100,
		Timeout:     2 * time.Second,
	}
}

// ScanPorts scans the specified ports on the host
func (u *PortScanUsecase) ScanPorts(ctx context.Context, host string, ports []int) (*models.PortScanResult, []error) {
	u.Logger.Info("Starting port scan", zap.String("host", host), zap.Int("ports", len(ports)))

	result := &models.PortScanResult{
		Host:       host,
		StartTime:  time.Now(),
		TotalPorts: len(ports),
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		errorsList []error
		openPorts  []models.PortResult
		closed     int
		filtered   int
	)

	// Semaphore for concurrency control
	sem := make(chan struct{}, u.Concurrency)

	for _, port := range ports {
		select {
		case <-ctx.Done():
			errorsList = append(errorsList, ctx.Err())
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, errorsList
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			portResult := u.Scanner.ScanPort(ctx, host, p, u.Timeout)

			mu.Lock()
			switch portResult.Status {
			case models.PortOpen:
				openPorts = append(openPorts, portResult)
			case models.PortClosed:
				closed++
			case models.PortFiltered:
				filtered++
			}
			mu.Unlock()
		}(port)
	}

	wg.Wait()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.OpenPorts = openPorts
	result.ClosedPorts = closed
	result.FilteredPorts = filtered

	u.Logger.Info("Port scan completed",
		zap.String("host", host),
		zap.Int("open", len(openPorts)),
		zap.Int("closed", closed),
		zap.Int("filtered", filtered),
		zap.Duration("duration", result.Duration))

	return result, errorsList
}

// ParsePortRange parses a port range string like "80,443,8080-8090"
func ParsePortRange(portRange string) ([]int, error) {
	var ports []int
	var start, end int

	// Split by comma
	for _, part := range splitPorts(portRange) {
		// Check if it's a range
		n, err := fmt.Sscanf(part, "%d-%d", &start, &end)
		if err == nil && n == 2 {
			if start > end {
				start, end = end, start
			}
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			// Single port
			var port int
			_, err := fmt.Sscanf(part, "%d", &port)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			ports = append(ports, port)
		}
	}

	return ports, nil
}

func splitPorts(s string) []string {
	var result []string
	var current string

	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if c != ' ' {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
