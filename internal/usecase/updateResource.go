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
    if err := uc.repo.Update(resource); err != nil {
        uc.logger.Sugar().Errorf("Failed to update resource with ID %d: %v", resource.ID, err)
        return err
    }

    uc.logger.Info("Resource updated successfully", zap.Uint("id", resource.ID))
    return nil
}
