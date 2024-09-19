package handler

import (
	"jorge-cli/config"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var(
	logger *zap.Logger
	db *gorm.DB
)

func InitializeHandler(){
	logger = config.GetLogger()
	db = config.GetSQLite()
}