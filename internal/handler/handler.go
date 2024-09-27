package handler

import (
	"github.com/iagonc/jorge-cli/internal/usecase"
	"go.uber.org/zap"
)

// Handler groups the use cases and the logger.
type Handler struct {
    CreateResourceUseCase     *usecase.CreateResourceUseCase
    DeleteResourceUseCase     *usecase.DeleteResourceUseCase
    ListResourcesUseCase      *usecase.ListResourcesUseCase
    ListResourcesByNameUseCase *usecase.ListResourcesByNameUseCase
    UpdateResourceUseCase     *usecase.UpdateResourceUseCase
    logger                    *zap.Logger
}

// NewHandler constructs a new Handler by injecting the use cases and the logger.
func NewHandler(
    createResourceUseCase *usecase.CreateResourceUseCase,
    deleteResourceUseCase *usecase.DeleteResourceUseCase,
    listResourcesUseCase *usecase.ListResourcesUseCase,
    listResourcesByNameUseCase *usecase.ListResourcesByNameUseCase,
    updateResourceUseCase *usecase.UpdateResourceUseCase,
    logger *zap.Logger,
) *Handler {
    return &Handler{
        CreateResourceUseCase:     createResourceUseCase,
        DeleteResourceUseCase:     deleteResourceUseCase,
        ListResourcesUseCase:      listResourcesUseCase,
        ListResourcesByNameUseCase: listResourcesByNameUseCase,
        UpdateResourceUseCase:     updateResourceUseCase,
        logger:                    logger,
    }
}
