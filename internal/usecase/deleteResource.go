package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type DeleteResource struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewDeleteResource(repo repository.ResourceRepository, logger *zap.Logger) *DeleteResource {
    return &DeleteResource{repo: repo, logger: logger}
}

// Execute deletes a resource by ID and returns the deleted resource or an error
func (uc *DeleteResource) Execute(id uint) (*schemas.Resource, error) {
    // Retrieve the resource before deleting it
    resource, err := uc.repo.FindByID(id)
    if err != nil {
        if err == repository.ErrResourceNotFound {
            uc.logger.Sugar().Errorf("Resource with ID %d not found", id)
            return nil, repository.ErrResourceNotFound
        }
        uc.logger.Sugar().Errorf("Failed to retrieve resource with ID %d: %v", id, err)
        return nil, err
    }

    // Delete the resource
    if err := uc.repo.Delete(resource.ID); err != nil {
        uc.logger.Sugar().Errorf("Failed to delete resource with ID %d: %v", id, err)
        return nil, err
    }

    uc.logger.Info("Resource deleted successfully", zap.Uint("id", id))

    // Return the deleted resource
    return resource, nil
}
