package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"go.uber.org/zap"
)

type DeleteResourceUseCase struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewDeleteResourceUseCase(repo repository.ResourceRepository, logger *zap.Logger) *DeleteResourceUseCase {
    return &DeleteResourceUseCase{repo: repo, logger: logger}
}

// Execute deletes a resource by ID and logs relevant information or errors
func (uc *DeleteResourceUseCase) Execute(id uint) error {
    resource, err := uc.repo.FindByID(id)
    if err != nil {
        if err == repository.ErrResourceNotFound {
            uc.logger.Sugar().Errorf("Resource with ID %d not found", id)
            return repository.ErrResourceNotFound
        }
        uc.logger.Sugar().Errorf("Failed to retrieve resource with ID %d: %v", id, err)
        return err
    }

    if err := uc.repo.Delete(resource.ID); err != nil {
        uc.logger.Sugar().Errorf("Failed to delete resource with ID %d: %v", id, err)
        return err
    }

    uc.logger.Info("Resource deleted successfully", zap.Uint("id", id))
    return nil
}
