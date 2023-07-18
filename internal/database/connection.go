package database

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var once sync.Once

const dsn = "host=localhost user=postgres password=g7bXjBroyyCXyL92ZJT4mQu6hc3pTQAA port=5433 sslmode=disable"

func GetInstance() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return db, err
}
