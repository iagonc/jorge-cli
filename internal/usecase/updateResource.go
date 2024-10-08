package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type UpdateResource struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewUpdateResource(repo repository.ResourceRepository, logger *zap.Logger) *UpdateResource {
    return &UpdateResource{repo: repo, logger: logger}
}

// Execute updates a resource and logs relevant information or errors
func (uc *UpdateResource) Execute(resource *schemas.Resource) error {
    existingResource, err := uc.repo.FindByID(resource.ID)
    if err != nil {
        uc.logger.Sugar().Errorf("Resource with ID %d not found: %v", resource.ID, err)
        return err
    }

    // Update fields if they are provided
    if resource.Name != "" {
        existingResource.Name = resource.Name
    }
    if resource.Dns != "" {
        existingResource.Dns = resource.Dns
    }

    if err := uc.repo.Update(existingResource); err != nil {
        uc.logger.Sugar().Errorf("Failed to update resource with ID %d: %v", resource.ID, err)
        return err
    }

    uc.logger.Info("Resource updated successfully", zap.Uint("id", resource.ID))
    return nil
}
