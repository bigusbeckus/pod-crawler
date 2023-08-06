package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dsn = "host=localhost user=postgres password=g7bXjBroyyCXyL92ZJT4mQu6hc3pTQAA port=5433 sslmode=disable"

var db *gorm.DB

func GetInstance() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		db = nil
	}

	return db, err
}
