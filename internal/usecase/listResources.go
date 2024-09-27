package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type ListResourcesUseCase struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewListResourcesUseCase(repo repository.ResourceRepository, logger *zap.Logger) *ListResourcesUseCase {
    return &ListResourcesUseCase{repo: repo, logger: logger}
}

func (uc *ListResourcesUseCase) Execute() ([]*schemas.Resource, error) {
    resources, err := uc.repo.List()
    if err != nil {
        uc.logger.Sugar().Errorf("Failed to list resources: %v", err)
        return nil, err
    }

    uc.logger.Info("Resources listed successfully")
    return resources, nil
}
