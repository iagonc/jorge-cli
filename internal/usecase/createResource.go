package usecase

import (
	"fmt"

	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type CreateResource struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewCreateResource(repo repository.ResourceRepository, logger *zap.Logger) *CreateResource {
    return &CreateResource{repo: repo, logger: logger}
}

// Execute creates a new resource and logs relevant information or errors
func (uc *CreateResource) Execute(resource *schemas.Resource) error {
    if _, err := uc.repo.FindByName(resource.Name); err == nil {
        uc.logger.Sugar().Errorf("Resource with name '%s' already exists", resource.Name)
        return fmt.Errorf("resource with name '%s' already exists", resource.Name)
    }

    if _, err := uc.repo.FindByDNS(resource.Dns); err == nil {
        uc.logger.Sugar().Errorf("Resource with DNS '%s' already exists", resource.Dns)
        return err
    }

    if err := uc.repo.Create(resource); err != nil {
        uc.logger.Sugar().Errorf("Failed to create resource: %v", err)
        return err
    }

    uc.logger.Info("Resource created successfully", zap.String("name", resource.Name))
    return nil
}
