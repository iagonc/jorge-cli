package config

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	logger *zap.Logger
	err error
)

func Init() error {
	logger = GetLogger()

	// Initialize SQLite
	db, err = InitializeSQLite()

	if err != nil {
		logger.Sugar().Errorf("Failed to initialize SQLite: %v", err)
		return err
	}

	return nil
}

func GetSQLite() *gorm.DB{
	return db
}

func GetLogger() *zap.Logger {
	logger, err := InitLogger()

	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

    return logger
}