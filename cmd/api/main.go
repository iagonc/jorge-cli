package main

import (
	"github.com/iagonc/jorge-cli/internal/config"
	"github.com/iagonc/jorge-cli/internal/handler"
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/router"
	"github.com/iagonc/jorge-cli/internal/usecase"
	"go.uber.org/zap"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Initialize SQLite database
    db, err := config.InitializeSQLite()
    if err != nil {
        logger.Sugar().Errorf("Failed to initialize SQLite: %v", err)
    }
    repo := repository.NewSQLiteResourceRepository(db)

    // Initialize use cases with the repository and logger
    createUseCase := usecase.NewCreateResourceUseCase(repo, logger)
    deleteUseCase := usecase.NewDeleteResourceUseCase(repo, logger)
    listUseCase := usecase.NewListResourcesUseCase(repo, logger)
    listByNameUseCase := usecase.NewListResourcesByNameUseCase(repo, logger)
    updateUseCase := usecase.NewUpdateResourceUseCase(repo, logger)

    // Initialize handler with the use cases and logger
    h := handler.NewHandler(
        createUseCase,
        deleteUseCase,
        listUseCase,
        listByNameUseCase,
        updateUseCase,
        logger,
    )

    // Pass the handler to the router and initialize routes
    router.Initialize(h)
}
