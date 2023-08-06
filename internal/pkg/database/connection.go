package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const dsn = "host=localhost user=postgres password=g7bXjBroyyCXyL92ZJT4mQu6hc3pTQAA port=5433 dbname=crawled-podcasts sslmode=disable"

var db *gorm.DB

func GetInstance() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	var err error
	c := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
		// Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err = gorm.Open(postgres.Open(dsn), c)
	if err != nil {
		db = nil
	}

	return db, err
}
