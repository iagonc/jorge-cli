package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type ListResourcesByNameUseCase struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewListResourcesByNameUseCase(repo repository.ResourceRepository, logger *zap.Logger) *ListResourcesByNameUseCase {
    return &ListResourcesByNameUseCase{repo: repo, logger: logger}
}

func (uc *ListResourcesByNameUseCase) Execute(name string) ([]*schemas.Resource, error) {
    resources, err := uc.repo.ListByName(name)
    if err != nil {
        uc.logger.Sugar().Errorf("Failed to list resources by name %s: %v", name, err)
        return nil, err
    }

    uc.logger.Info("Resources listed successfully by name", zap.String("name", name))
    return resources, nil
}
