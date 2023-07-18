package database

import (
	"fmt"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
)

func RunMigrations() error {
	db, err := GetInstance()
	if err != nil {
		fmt.Println("Unable to get database instance")
		return err
	}

	podcastModelErr := db.AutoMigrate(&models.PodcastModel{})

	if podcastModelErr != nil {
		fmt.Println("Migrations failed")
		return podcastModelErr
	}

	return nil
}
