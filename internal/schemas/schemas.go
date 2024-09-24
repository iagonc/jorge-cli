package schemas

import (
	"gorm.io/gorm"
)

type Resource struct{
	gorm.Model
    Name      string    `json:"name" validate:"required"`
	Dns   	  string    `json:"dns" validate:"required"`
}
