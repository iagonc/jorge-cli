package config

import (
	"os"

	"github.com/iagonc/jorge-cli/internal/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ensureDbPathExists() (string, error) {
	logger := GetLogger()
	dbPath := "./db/main.db"
	
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		logger.Sugar().Infof("Database file not found at %v, creating...", dbPath)

		dirPath := "./db"
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			logger.Sugar().Errorf("Failed to create database directory at %v: %v", dirPath, err)
			return "", err
		}

		file, err := os.Create(dbPath)
		if err != nil {
			logger.Sugar().Errorf("Failed to create database file at %v: %v", dbPath, err)
			return "", err
		}

		err = file.Close()
		if err != nil {
			logger.Sugar().Errorf("Failed to close the file %v: %v", dbPath, err)
			return "", err
		}

		logger.Sugar().Infof("Database file successfully created at %v", dbPath)
	} else if err != nil {
		logger.Sugar().Errorf("Error accessing database file at %v: %v", dbPath, err)
		return "", err
	}

	return dbPath, nil
}

func InitializeSQLite() (*gorm.DB, error) {
	logger := GetLogger()

	validatedPath, err := ensureDbPathExists()
	if err != nil {
		logger.Sugar().Errorf("error validating database path: %w", err)
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(validatedPath), &gorm.Config{})
	if err != nil {
		logger.Sugar().Errorf("Failed to connect to the database at %v: %v", validatedPath, err)
		return nil, err
	}

	err = db.AutoMigrate(&schemas.Resource{})
	if err != nil {
		logger.Sugar().Errorf("Failed to auto-migrate the database: %v", err)
		return nil, err
	}

	logger.Info("Database initialized and migrated successfully")
	return db, nil
}
