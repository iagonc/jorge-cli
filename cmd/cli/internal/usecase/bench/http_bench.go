package bench

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// HTTPBenchmarkUsecase handles HTTP benchmarking operations
type HTTPBenchmarkUsecase struct {
	Logger *zap.Logger
}

// NewHTTPBenchmarkUsecase creates a new HTTPBenchmarkUsecase
func NewHTTPBenchmarkUsecase(logger *zap.Logger) *HTTPBenchmarkUsecase {
	return &HTTPBenchmarkUsecase{
		Logger: logger,
	}
}

// RunBenchmark runs an HTTP benchmark with the given configuration
func (u *HTTPBenchmarkUsecase) RunBenchmark(ctx context.Context, config models.BenchmarkConfig) (*models.BenchmarkResult, error) {
	u.Logger.Info("Starting HTTP benchmark",
		zap.String("url", config.URL),
		zap.Int("requests", config.Requests),
		zap.Int("concurrency", config.Concurrency))

	if config.Concurrency <= 0 {
		config.Concurrency = 10
	}
	if config.Requests <= 0 {
		config.Requests = 100
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Method == "" {
		config.Method = http.MethodGet
	}

	result := &models.BenchmarkResult{
		URL:         config.URL,
		StatusCodes: make(map[int]int),
		Errors:      make(map[string]int),
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.Concurrency,
			MaxIdleConnsPerHost: config.Concurrency,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Channel for request distribution
	requestChan := make(chan int, config.Requests)
	resultsChan := make(chan models.RequestResult, config.Requests)

	// Fill request channel
	for i := 0; i < config.Requests; i++ {
		requestChan <- i
	}
	close(requestChan)

	var wg sync.WaitGroup
	var completed int64

	startTime := time.Now()

	// Start workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				reqResult := u.doRequest(ctx, client, config)
				resultsChan <- reqResult
				atomic.AddInt64(&completed, 1)
			}
		}()
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var latencies []time.Duration
	var totalBytes int64

	for reqResult := range resultsChan {
		result.TotalRequests++

		if reqResult.Error != nil {
			result.FailedRequests++
			errMsg := reqResult.Error.Error()
			result.Errors[errMsg]++
		} else {
			result.SuccessfulRequests++
			result.StatusCodes[reqResult.StatusCode]++
			latencies = append(latencies, reqResult.Duration)
			totalBytes += reqResult.Size
		}
	}

	result.TotalDuration = time.Since(startTime)
	result.BytesReceived = totalBytes

	// Calculate statistics
	if len(latencies) > 0 {
		result.Latency = calculateLatencyStats(latencies)
	}

	if result.TotalDuration.Seconds() > 0 {
		result.RequestsPerSecond = float64(result.TotalRequests) / result.TotalDuration.Seconds()
	}

	u.Logger.Info("Benchmark completed",
		zap.Int("total", result.TotalRequests),
		zap.Int("successful", result.SuccessfulRequests),
		zap.Int("failed", result.FailedRequests),
		zap.Float64("rps", result.RequestsPerSecond))

	return result, nil
}

func (u *HTTPBenchmarkUsecase) doRequest(ctx context.Context, client *http.Client, config models.BenchmarkConfig) models.RequestResult {
	result := models.RequestResult{}

	var bodyReader io.Reader
	if config.Body != "" {
		bodyReader = strings.NewReader(config.Body)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		result.Error = err
		return result
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Read and discard body to measure full response time
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}
	result.Size = int64(len(body))

	return result
}

func calculateLatencyStats(latencies []time.Duration) models.LatencyStats {
	if len(latencies) == 0 {
		return models.LatencyStats{}
	}

	// Sort for percentile calculations
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var total time.Duration
	for _, l := range latencies {
		total += l
	}

	return models.LatencyStats{
		Min:  latencies[0],
		Max:  latencies[len(latencies)-1],
		Mean: total / time.Duration(len(latencies)),
		P50:  percentile(latencies, 50),
		P95:  percentile(latencies, 95),
		P99:  percentile(latencies, 99),
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	index := int(float64(len(sorted)-1) * p / 100)
	return sorted[index]
}
