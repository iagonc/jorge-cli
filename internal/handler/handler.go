package handler

import (
	"github.com/iagonc/jorge-cli/internal/usecase"
	"go.uber.org/zap"
)

// Handler groups the use cases and the logger.
type Handler struct {
    CreateResource     *usecase.CreateResource
    DeleteResource     *usecase.DeleteResource
    GetResourceByID     *usecase.GetResourceByID
    ListResources      *usecase.ListResources
    ListResourcesByName *usecase.ListResourcesByName
    UpdateResource     *usecase.UpdateResource
    logger                    *zap.Logger
}

// NewHandler constructs a new Handler by injecting the use cases and the logger.
func NewHandler(
    createResource *usecase.CreateResource,
    deleteResource *usecase.DeleteResource,
    getResourceByID *usecase.GetResourceByID,
    listResources *usecase.ListResources,
    listResourcesByName *usecase.ListResourcesByName,
    updateResource *usecase.UpdateResource,
    logger *zap.Logger,
) *Handler {
    return &Handler{
        CreateResource:     createResource,
        DeleteResource:     deleteResource,
        GetResourceByID:    getResourceByID,
        ListResources:      listResources,
        ListResourcesByName: listResourcesByName,
        UpdateResource:     updateResource,
        logger:                    logger,
    }
}
