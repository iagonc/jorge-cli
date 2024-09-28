package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type GetResourceByID struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewGetResourceByID(repo repository.ResourceRepository, logger *zap.Logger) *GetResourceByID {
    return &GetResourceByID{repo: repo, logger: logger}
}

// Execute deletes a resource by ID and returns the deleted resource or an error
func (uc *GetResourceByID) Execute(id uint) (*schemas.Resource, error) {
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

    uc.logger.Info("Resource found successfully", zap.Uint("id", id))

    // Return the deleted resource
    return resource, nil
}
