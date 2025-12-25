package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockUsecase handles mock server operations
type MockUsecase struct {
	logger   *zap.Logger
	server   *http.Server
	requests []models.MockRequest
	stats    models.MockStats
	mu       sync.RWMutex
	config   models.MockConfig
}

// NewMockUsecase creates a new MockUsecase
func NewMockUsecase(logger *zap.Logger) *MockUsecase {
	return &MockUsecase{
		logger:   logger,
		requests: []models.MockRequest{},
		stats: models.MockStats{
			RequestsByPath:   make(map[string]int),
			RequestsByMethod: make(map[string]int),
		},
	}
}

// RequestCallback is called for each incoming request
type RequestCallback func(req models.MockRequest, resp models.MockResponse)

// Start starts the mock server
func (u *MockUsecase) Start(ctx context.Context, config models.MockConfig, callback RequestCallback) error {
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	if config.DefaultStatus == 0 {
		config.DefaultStatus = 200
	}

	u.config = config
	u.stats.StartTime = time.Now()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u.handleRequest(w, r, callback)
	})

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	u.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	u.logger.Info("Starting mock server", zap.String("addr", addr))

	errCh := make(chan error, 1)
	go func() {
		if err := u.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		u.Stop()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// Stop stops the mock server
func (u *MockUsecase) Stop() error {
	if u.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return u.server.Shutdown(ctx)
	}
	return nil
}

func (u *MockUsecase) handleRequest(w http.ResponseWriter, r *http.Request, callback RequestCallback) {
	start := time.Now()

	// Read body
	bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	defer r.Body.Close()

	// Create request record
	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}

	mockReq := models.MockRequest{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		Method:      r.Method,
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
		Headers:     headers,
		Body:        string(bodyBytes),
		RemoteAddr:  r.RemoteAddr,
		ContentType: r.Header.Get("Content-Type"),
		BodySize:    int64(len(bodyBytes)),
	}

	// Store request
	u.mu.Lock()
	u.requests = append(u.requests, mockReq)
	u.stats.TotalRequests++
	u.stats.RequestsByPath[r.URL.Path]++
	u.stats.RequestsByMethod[r.Method]++
	u.mu.Unlock()

	// Find matching route
	var matchedRoute *models.MockRoute
	for _, route := range u.config.Routes {
		if route.Path == r.URL.Path && (route.Method == "" || route.Method == r.Method) {
			matchedRoute = &route
			break
		}
	}

	// Prepare response
	mockResp := models.MockResponse{
		StatusCode: u.config.DefaultStatus,
		Headers:    make(map[string]string),
	}

	if matchedRoute != nil {
		mockResp.StatusCode = matchedRoute.StatusCode
		mockResp.Body = matchedRoute.Body
		for k, v := range matchedRoute.Headers {
			mockResp.Headers[k] = v
		}
		if matchedRoute.ContentType != "" {
			w.Header().Set("Content-Type", matchedRoute.ContentType)
		}
		if matchedRoute.Delay > 0 {
			time.Sleep(matchedRoute.Delay)
		}
	}

	// Apply global delay
	if u.config.Delay > 0 {
		time.Sleep(u.config.Delay)
	}

	// CORS headers if enabled
	if u.config.CORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Apply response headers
	for k, v := range mockResp.Headers {
		w.Header().Set(k, v)
	}

	// Write response
	w.WriteHeader(mockResp.StatusCode)
	if mockResp.Body != "" {
		w.Write([]byte(mockResp.Body))
	} else {
		// Default response
		response := map[string]interface{}{
			"message":   "Mock response",
			"path":      r.URL.Path,
			"method":    r.Method,
			"timestamp": time.Now().Format(time.RFC3339),
			"request_id": mockReq.ID,
		}
		if len(bodyBytes) > 0 {
			response["received_body"] = string(bodyBytes)
		}
		json.NewEncoder(w).Encode(response)
	}

	mockResp.Duration = time.Since(start)

	// Callback
	if callback != nil {
		callback(mockReq, mockResp)
	}
}

// GetRequests returns all recorded requests
func (u *MockUsecase) GetRequests() []models.MockRequest {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return append([]models.MockRequest{}, u.requests...)
}

// GetStats returns current statistics
func (u *MockUsecase) GetStats() models.MockStats {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.stats
}

// ClearRequests clears all recorded requests
func (u *MockUsecase) ClearRequests() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.requests = []models.MockRequest{}
}
