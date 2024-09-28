package usecase

import (
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type ListResourcesByName struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewListResourcesByName(repo repository.ResourceRepository, logger *zap.Logger) *ListResourcesByName {
    return &ListResourcesByName{repo: repo, logger: logger}
}

// Execute retrieves resources by name and logs relevant information or errors
func (uc *ListResourcesByName) Execute(name string) ([]*schemas.Resource, error) {
    resources, err := uc.repo.ListByName(name)
    if err != nil {
        uc.logger.Sugar().Errorf("Failed to list resources by name '%s': %v", name, err)
        return nil, err
    }

    uc.logger.Info("Resources listed successfully by name", zap.String("name", name))
    return resources, nil
}
