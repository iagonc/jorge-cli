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

func (uc *DeleteResourceUseCase) Execute(id uint) error {
    // Verificar se o recurso existe
    resource, err := uc.repo.FindByID(id)
    if err != nil {
        // Verifica se o erro é que o recurso não foi encontrado
        if err == repository.ErrResourceNotFound {
            uc.logger.Sugar().Errorf("Resource with ID %d not found", id)
            return err // Retorna erro personalizado
        }
        // Loga outros erros
        uc.logger.Sugar().Errorf("Error finding resource with ID %d: %v", id, err)
        return err // Retorna erro genérico
    }

    // Deletar o recurso
    if err := uc.repo.Delete(resource.ID); err != nil {
        uc.logger.Sugar().Errorf("Failed to delete resource with ID %d: %v", id, err)
        return err // Retorna erro genérico
    }

    uc.logger.Info("Resource deleted successfully", zap.Uint("id", id))
    return nil
}
