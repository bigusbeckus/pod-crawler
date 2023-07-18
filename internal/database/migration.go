package database

import "github.com/bigusbeckus/podcast-feed-fetcher/internal/database/models"

func RunMigrations() {
	db, err := GetInstance()
	if err != nil {
		panic("Unable to get database instance")
	}

	err = db.AutoMigrate(&models.PodcastModel{})

	if err != nil {
		panic("Migrations failed")
	}
}
