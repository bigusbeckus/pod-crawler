package database

import (
	"errors"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
)

func RunMigrations() error {
	db, err := GetInstance()
	if err != nil {
		logger.Error.Println("Unable to get database instance")
		return err
	}

	artistModelErr := db.AutoMigrate(&models.Artist{})
	genreModelErr := db.AutoMigrate(&models.Genre{})
	podcastModelErr := db.AutoMigrate(&models.Podcast{})
	podcastGenreModelErr := db.AutoMigrate(&models.PodcastGenre{})

	err = errors.Join(
		artistModelErr,
		genreModelErr,
		podcastModelErr,
		podcastGenreModelErr,
	)

	if err != nil {
		logger.Error.Println("Migration queries failed")
		return podcastModelErr
	}

	return nil
}
