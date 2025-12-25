package testutil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/handler"
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"github.com/iagonc/jorge-cli/internal/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TestServer represents a test server instance
type TestServer struct {
	Server  *http.Server
	DB      *gorm.DB
	BaseURL string
	Port    int
	t       *testing.T
}

// NewTestServer creates and starts a new test server
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Setup in-memory database
	db := SetupTestDB(t)

	// Create repository and usecases
	logger := zap.NewNop()
	repo := repository.NewSQLiteResourceRepository(db)

	createUseCase := usecase.NewCreateResource(repo, logger)
	deleteUseCase := usecase.NewDeleteResource(repo, logger)
	getResourceByIDUseCase := usecase.NewGetResourceByID(repo, logger)
	listUseCase := usecase.NewListResources(repo, logger)
	listByNameUseCase := usecase.NewListResourcesByName(repo, logger)
	updateUseCase := usecase.NewUpdateResource(repo, logger)

	h := handler.NewHandler(
		createUseCase,
		deleteUseCase,
		getResourceByIDUseCase,
		listUseCase,
		listByNameUseCase,
		updateUseCase,
		logger,
	)

	// Create router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	basePath := "/api/v1"
	v1 := router.Group(basePath)
	{
		v1.GET("/resources", h.ListResourcesHandler)
		v1.GET("/resources/name", h.ListResourcesByNameHandler)
		v1.GET("/resource", h.GetResourceByIDHandler)
		v1.POST("/resource", h.CreateResourceHandler)
		v1.PUT("/resource", h.UpdateResourceHandler)
		v1.DELETE("/resource", h.DeleteResourceHandler)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	ts := &TestServer{
		Server:  server,
		DB:      db,
		BaseURL: fmt.Sprintf("http://localhost:%d/api/v1", port),
		Port:    port,
		t:       t,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Server stopped
		}
	}()

	// Wait for server to be ready
	ts.WaitForReady()

	// Cleanup on test end
	t.Cleanup(func() {
		ts.Shutdown()
	})

	return ts
}

// WaitForReady waits for the server to be ready to accept connections
func (ts *TestServer) WaitForReady() {
	maxRetries := 50
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", ts.Port))
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ts.t.Fatalf("Server failed to start within timeout")
}

// Shutdown gracefully shuts down the test server
func (ts *TestServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ts.Server.Shutdown(ctx)
}

// SeedData seeds the test database with sample data
func (ts *TestServer) SeedData() []*schemas.Resource {
	return SeedTestData(ts.t, ts.DB)
}

// CleanupData removes all data from the test database
func (ts *TestServer) CleanupData() {
	CleanupTestData(ts.t, ts.DB)
}
