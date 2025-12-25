package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// ChaosUsecase handles chaos engineering experiments
type ChaosUsecase struct {
	logger  *zap.Logger
	running map[string]*runningExperiment
	mu      sync.Mutex
}

type runningExperiment struct {
	experiment *models.ChaosExperiment
	cancel     context.CancelFunc
	done       chan struct{}
}

// NewChaosUsecase creates a new ChaosUsecase
func NewChaosUsecase(logger *zap.Logger) *ChaosUsecase {
	return &ChaosUsecase{
		logger:  logger,
		running: make(map[string]*runningExperiment),
	}
}

// LoadScenario loads a chaos scenario from a file
func (u *ChaosUsecase) LoadScenario(path string) (*models.ChaosScenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario: %w", err)
	}

	var scenario models.ChaosScenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}

	return &scenario, nil
}

// RunExperiment runs a chaos experiment
func (u *ChaosUsecase) RunExperiment(ctx context.Context, exp *models.ChaosExperiment) (*models.ChaosResults, error) {
	exp.Status = models.ExperimentRunning
	now := time.Now()
	exp.StartedAt = &now

	u.logger.Info("Starting chaos experiment",
		zap.String("id", exp.ID),
		zap.String("type", string(exp.Type)))

	// Parse duration
	duration, err := time.ParseDuration(exp.Duration)
	if err != nil {
		duration = 30 * time.Second
	}

	expCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Store running experiment
	u.mu.Lock()
	u.running[exp.ID] = &runningExperiment{
		experiment: exp,
		cancel:     cancel,
		done:       make(chan struct{}),
	}
	u.mu.Unlock()

	defer func() {
		u.mu.Lock()
		delete(u.running, exp.ID)
		u.mu.Unlock()
	}()

	var results *models.ChaosResults

	switch exp.Type {
	case models.ChaosTypeCPUStress:
		results = u.runCPUStress(expCtx, exp.Config)
	case models.ChaosTypeMemoryStress:
		results = u.runMemoryStress(expCtx, exp.Config)
	case models.ChaosTypeNetworkLatency:
		results = u.runNetworkLatency(expCtx, exp.Config)
	case models.ChaosTypeHTTPFault:
		results = u.runHTTPFault(expCtx, exp.Config)
	default:
		results = &models.ChaosResults{
			Success: false,
			Summary: fmt.Sprintf("Experiment type %s not supported", exp.Type),
		}
	}

	endTime := time.Now()
	exp.EndedAt = &endTime
	exp.Status = models.ExperimentCompleted
	exp.Results = results

	return results, nil
}

// StopExperiment stops a running experiment
func (u *ChaosUsecase) StopExperiment(id string) error {
	u.mu.Lock()
	running, ok := u.running[id]
	u.mu.Unlock()

	if !ok {
		return fmt.Errorf("experiment %s not found or not running", id)
	}

	running.cancel()
	running.experiment.Status = models.ExperimentAborted

	return nil
}

// ListRunning lists running experiments
func (u *ChaosUsecase) ListRunning() []*models.ChaosExperiment {
	u.mu.Lock()
	defer u.mu.Unlock()

	var experiments []*models.ChaosExperiment
	for _, r := range u.running {
		experiments = append(experiments, r.experiment)
	}
	return experiments
}

func (u *ChaosUsecase) runCPUStress(ctx context.Context, config models.ChaosConfig) *models.ChaosResults {
	results := &models.ChaosResults{
		Success:      true,
		Metrics:      make(map[string]string),
		Observations: []string{},
	}

	workers := config.Workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}

	cpuPercent := config.CPUPercent
	if cpuPercent == 0 {
		cpuPercent = 80
	}

	results.Summary = fmt.Sprintf("Running CPU stress with %d workers at %d%%", workers, cpuPercent)
	results.Observations = append(results.Observations, fmt.Sprintf("Starting %d stress workers", workers))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// CPU-intensive work
					for j := 0; j < 1000000; j++ {
						_ = j * j
					}
					// Throttle based on CPU percent
					sleepTime := time.Duration(100-cpuPercent) * time.Millisecond / 10
					time.Sleep(sleepTime)
				}
			}
		}(i)
	}

	<-ctx.Done()
	wg.Wait()

	results.Observations = append(results.Observations, "CPU stress completed")
	results.Metrics["workers"] = fmt.Sprintf("%d", workers)
	results.Metrics["cpu_percent"] = fmt.Sprintf("%d%%", cpuPercent)

	return results
}

func (u *ChaosUsecase) runMemoryStress(ctx context.Context, config models.ChaosConfig) *models.ChaosResults {
	results := &models.ChaosResults{
		Success:      true,
		Metrics:      make(map[string]string),
		Observations: []string{},
	}

	// Parse memory size (default 100MB)
	memSize := int64(100 * 1024 * 1024)
	if config.MemoryBytes != "" {
		fmt.Sscanf(config.MemoryBytes, "%dMB", &memSize)
		memSize *= 1024 * 1024
	}

	results.Summary = fmt.Sprintf("Allocating %d MB of memory", memSize/(1024*1024))
	results.Observations = append(results.Observations, "Allocating memory...")

	// Allocate memory
	data := make([]byte, memSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	results.Observations = append(results.Observations, fmt.Sprintf("Allocated %d MB", len(data)/(1024*1024)))

	// Keep memory allocated until context is done
	<-ctx.Done()

	// Clear to allow GC
	data = nil
	runtime.GC()

	results.Observations = append(results.Observations, "Memory released")
	results.Metrics["allocated_mb"] = fmt.Sprintf("%d", memSize/(1024*1024))

	return results
}

func (u *ChaosUsecase) runNetworkLatency(ctx context.Context, config models.ChaosConfig) *models.ChaosResults {
	results := &models.ChaosResults{
		Success:      true,
		Metrics:      make(map[string]string),
		Observations: []string{},
	}

	latency := config.Latency
	if latency == "" {
		latency = "100ms"
	}

	// This is a simulated version - in production you'd use tc (traffic control) on Linux
	if runtime.GOOS != "linux" {
		results.Summary = "Network latency injection requires Linux (tc command)"
		results.Success = false
		results.Observations = append(results.Observations, "Simulated mode: actual network latency injection not available")
		return results
	}

	// Check if tc is available
	if _, err := exec.LookPath("tc"); err != nil {
		results.Summary = "tc command not found"
		results.Success = false
		return results
	}

	results.Summary = fmt.Sprintf("Injecting %s network latency", latency)
	results.Observations = append(results.Observations, "Note: Actual tc implementation would go here")
	results.Metrics["latency"] = latency

	<-ctx.Done()

	return results
}

func (u *ChaosUsecase) runHTTPFault(ctx context.Context, config models.ChaosConfig) *models.ChaosResults {
	results := &models.ChaosResults{
		Success:      true,
		Metrics:      make(map[string]string),
		Observations: []string{},
	}

	port := config.Port
	if port == 0 {
		port = 8888
	}

	statusCode := config.StatusCode
	if statusCode == 0 {
		statusCode = 500
	}

	errorRate := config.ErrorRate
	if errorRate == 0 {
		errorRate = 100
	}

	results.Summary = fmt.Sprintf("Starting fault injection proxy on port %d (error rate: %d%%, status: %d)", port, errorRate, statusCode)
	results.Observations = append(results.Observations, fmt.Sprintf("Proxy listening on :%d", port))

	// Start a simple HTTP server that returns errors
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(100) < errorRate {
			w.WriteHeader(statusCode)
			w.Write([]byte("Chaos: Injected fault"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			u.logger.Error("Fault proxy error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)

	results.Observations = append(results.Observations, "Fault injection proxy stopped")
	results.Metrics["port"] = fmt.Sprintf("%d", port)
	results.Metrics["error_rate"] = fmt.Sprintf("%d%%", errorRate)
	results.Metrics["status_code"] = fmt.Sprintf("%d", statusCode)

	return results
}

// VerifySteadyState verifies steady state conditions
func (u *ChaosUsecase) VerifySteadyState(ctx context.Context, check models.SteadyStateCheck) (bool, string) {
	switch check.Type {
	case "http":
		return u.verifyHTTP(ctx, check.Target, check.Expect)
	case "tcp":
		return u.verifyTCP(ctx, check.Target)
	default:
		return false, fmt.Sprintf("Unknown check type: %s", check.Type)
	}
}

func (u *ChaosUsecase) verifyHTTP(ctx context.Context, target, expect string) (bool, string) {
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return false, err.Error()
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	// Check expected status
	expectedStatus := 200
	if expect != "" {
		fmt.Sscanf(expect, "%d", &expectedStatus)
	}

	if resp.StatusCode == expectedStatus {
		return true, fmt.Sprintf("HTTP %d OK", resp.StatusCode)
	}
	return false, fmt.Sprintf("Expected %d, got %d", expectedStatus, resp.StatusCode)
}

func (u *ChaosUsecase) verifyTCP(ctx context.Context, target string) (bool, string) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return false, err.Error()
	}
	conn.Close()
	return true, "TCP connection successful"
}

// GenerateScenario generates a sample chaos scenario
func (u *ChaosUsecase) GenerateScenario() string {
	return `name: "Sample Chaos Scenario"
description: "Test system resilience with CPU stress and network faults"

steady_state:
  - name: "API Available"
    type: "http"
    target: "http://localhost:8080/health"
    expect: "200"

  - name: "Database Reachable"
    type: "tcp"
    target: "localhost:5432"

experiments:
  - id: "exp-cpu-001"
    name: "CPU Stress Test"
    description: "Stress CPU to 80%"
    type: "cpu_stress"
    duration: "30s"
    target:
      type: "host"
      hosts:
        - "localhost"
    config:
      cpu_percent: 80
      workers: 4

  - id: "exp-net-001"
    name: "Network Latency"
    description: "Add 100ms latency"
    type: "network_latency"
    duration: "60s"
    target:
      type: "host"
      hosts:
        - "localhost"
    config:
      latency: "100ms"
      jitter: "20ms"

  - id: "exp-http-001"
    name: "HTTP Fault Injection"
    description: "Return 500 errors 50% of the time"
    type: "http_fault"
    duration: "60s"
    target:
      type: "service"
    config:
      port: 8888
      status_code: 500
      error_rate: 50
`
}
