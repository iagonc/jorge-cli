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
	// Inicializando o logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	
	// Configuração do banco de dados
	db, err := config.InitializeSQLite()
	if err != nil {
		logger.Sugar().Errorf("Failed to initialize SQLite: %v", err)
	}
	repo := repository.NewSQLiteResourceRepository(db)

	// Inicializando os casos de uso com o repositório e o logger
	createUseCase := usecase.NewCreateResourceUseCase(repo, logger)
	deleteUseCase := usecase.NewDeleteResourceUseCase(repo, logger)
	listUseCase := usecase.NewListResourcesUseCase(repo, logger)
	listByNameUseCase := usecase.NewListResourcesByNameUseCase(repo, logger)
	updateUseCase := usecase.NewUpdateResourceUseCase(repo, logger)

	// Inicializando o handler com os casos de uso e o logger
	h := handler.NewHandler(
		createUseCase, 
		deleteUseCase, 
		listUseCase, 
		listByNameUseCase, 
		updateUseCase, 
		logger,
	)

	// Passa o handler para o router e inicializa as rotas
	router.Initialize(h)
}
