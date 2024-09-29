package usecase

import (
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/config"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"go.uber.org/zap"
)

type ResourceUsecase struct {
    Client utils.HTTPClient
    Config *config.Config
    Logger *zap.Logger
}

func NewResourceUsecase(client utils.HTTPClient, cfg *config.Config, logger *zap.Logger) *ResourceUsecase {
    return &ResourceUsecase{
        Client: client,
        Config: cfg,
        Logger: logger,
    }
}
