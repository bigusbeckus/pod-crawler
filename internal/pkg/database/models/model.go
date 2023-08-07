package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UpdatedAt time.Time      `gorm:"default:null"`
	DeletedAt gorm.DeletedAt `gorm:"index;default:null"`
}
