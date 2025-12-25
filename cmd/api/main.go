package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iagonc/jorge-cli/internal/config"
	"github.com/iagonc/jorge-cli/internal/handler"
	"github.com/iagonc/jorge-cli/internal/monitor"
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/router"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"github.com/iagonc/jorge-cli/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	log.Println("[API] Starting Jorge API Server...")

	// Initialize SQLite database
	db, err := config.InitializeSQLite()
	if err != nil {
		logger.Sugar().Errorf("Failed to initialize SQLite: %v", err)
		return
	}

	// Auto-migrate new schemas
	log.Println("[API] Running database migrations...")
	if err := db.AutoMigrate(
		&schemas.Resource{},
		&schemas.Monitor{},
		&schemas.CheckResult{},
		&schemas.Alert{},
	); err != nil {
		logger.Sugar().Errorf("Failed to migrate schemas: %v", err)
		return
	}

	// Initialize resource repository and use cases
	repo := repository.NewSQLiteResourceRepository(db)
	createUseCase := usecase.NewCreateResource(repo, logger)
	deleteUseCase := usecase.NewDeleteResource(repo, logger)
	getResourceByIDUseCase := usecase.NewGetResourceByID(repo, logger)
	listUseCase := usecase.NewListResources(repo, logger)
	listByNameUseCase := usecase.NewListResourcesByName(repo, logger)
	updateUseCase := usecase.NewUpdateResource(repo, logger)

	// Initialize resource handler
	h := handler.NewHandler(
		createUseCase,
		deleteUseCase,
		getResourceByIDUseCase,
		listUseCase,
		listByNameUseCase,
		updateUseCase,
		logger,
	)

	// Initialize monitoring scheduler
	scheduler := monitor.NewScheduler(db)

	// Initialize monitor handler
	mh := handler.NewMonitorHandler(db, scheduler)

	// Start scheduler automatically
	log.Println("[API] Starting monitoring scheduler...")
	scheduler.Start()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("[API] Shutting down...")
		scheduler.Stop()
		os.Exit(0)
	}()

	// Start API server
	log.Println("[API] Server running on http://localhost:8080")
	router.Initialize(h, mh)
}
