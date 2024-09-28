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
    createUseCase := usecase.NewCreateResource(repo, logger)
    deleteUseCase := usecase.NewDeleteResource(repo, logger)
    getResourceByIDUseCase := usecase.NewGetResourceByID(repo, logger)
    listUseCase := usecase.NewListResources(repo, logger)
    listByNameUseCase := usecase.NewListResourcesByName(repo, logger)
    updateUseCase := usecase.NewUpdateResource(repo, logger)

    // Initialize handler with the use cases and logger
    h := handler.NewHandler(
        createUseCase,
        deleteUseCase,
        getResourceByIDUseCase,
        listUseCase,
        listByNameUseCase,
        updateUseCase,
        logger,
    )

    // Pass the handler to the router and initialize routes
    router.Initialize(h)
}
