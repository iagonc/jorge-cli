package resource

import (
	"github.com/iagonc/jorge-cli/cmd/cli/internal/config"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"

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
