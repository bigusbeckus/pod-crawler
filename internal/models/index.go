package models

import (
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/db"
)

func RunMigrations() {

	dbInstance, err := db.GetInstance()
	if err != nil {
		panic("Unable to get database instance")
	}

	err = dbInstance.AutoMigrate(&PodcastModel{})

	if err != nil {
		panic("Migrations failed")
	}
}
