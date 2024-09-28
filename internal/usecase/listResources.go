package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type ListResources struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewListResources(repo repository.ResourceRepository, logger *zap.Logger) *ListResources {
    return &ListResources{repo: repo, logger: logger}
}

// Execute retrieves and logs all available resources
func (uc *ListResources) Execute() ([]*schemas.Resource, error) {
    resources, err := uc.repo.List()
    if err != nil {
        uc.logger.Sugar().Errorf("Failed to list resources: %v", err)
        return nil, err
    }

    uc.logger.Info("Resources listed successfully")
    return resources, nil
}
