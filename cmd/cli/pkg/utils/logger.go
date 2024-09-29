package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func InitializeLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		return nil, err
	}
	return logger, nil
}
