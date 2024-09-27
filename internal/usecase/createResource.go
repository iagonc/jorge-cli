package usecase

import (
	"fmt"

	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"go.uber.org/zap"
)

type CreateResourceUseCase struct {
    repo   repository.ResourceRepository
    logger *zap.Logger
}

func NewCreateResourceUseCase(repo repository.ResourceRepository, logger *zap.Logger) *CreateResourceUseCase {
    return &CreateResourceUseCase{repo: repo, logger: logger}
}

func (uc *CreateResourceUseCase) Execute(resource *schemas.Resource) error {
    if _, err := uc.repo.FindByName(resource.Name); err == nil {
        uc.logger.Sugar().Errorf("Resource with name %s already exists", resource.Name)
        return fmt.Errorf("resource with name %s already exists", resource.Name)
    }

    if _, err := uc.repo.FindByDNS(resource.Dns); err == nil {
        uc.logger.Sugar().Errorf("Resource with DNS %s already exists", resource.Dns)
        return err
    }

    if err := uc.repo.Create(resource); err != nil {
        uc.logger.Sugar().Errorf("Failed to create resource: %v", err)
        return err
    }

    uc.logger.Info("Resource created successfully", zap.String("name", resource.Name))
    return nil
}
