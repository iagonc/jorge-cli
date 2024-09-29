package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Log is a global logger to be used throughout the project
var Log *zap.Logger

func InitializeLogger() (*zap.Logger, error) {
    logger, err := zap.NewProduction()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
        return nil, err
    }
    return logger, nil
}
