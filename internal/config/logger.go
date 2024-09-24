package config

import (
	"go.uber.org/zap"
)

func InitLogger() (*zap.Logger, error) {
    config := zap.NewProductionConfig()

    l, err := config.Build()
    if err != nil {
        return nil, err
    }
	
    return l, nil
}
