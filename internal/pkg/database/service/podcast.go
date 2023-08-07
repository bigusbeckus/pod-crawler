package service

import (
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
)

func PodcastFromItunesResult(result podcast.ItunesResult) (*models.Podcast, error) {
	p := models.Podcast{
		Title:         *result.CollectionName,
		CensoredTitle: *result.CollectionCensoredName,
		FeedUrl:       result.FeedUrl,
		ArtistName:    &result.ArtistName,
		ReleaseDate:   &result.ReleaseDate,

		Country:               &result.Country,
		EpisodeCount:          &result.TrackCount,
		ContentAdvisoryRating: &result.ContentAdvisoryRating,

		ItunesID:            &result.CollectionId,
		ItunesViewUrl:       &result.CollectionViewUrl,
		ItunesArtworkUrl30:  &result.ArtworkUrl30,
		ItunesArtworkUrl60:  &result.ArtworkUrl60,
		ItunesArtworkUrl100: &result.ArtworkUrl100,
		ItunesArtworkUrl600: &result.ArtworkUrl600,

		ItunesArtistId:      result.ArtistId,
		ItunesArtistViewUrl: result.ArtistViewUrl,
	}

	db, _ := database.GetInstance()

	// Cleanup genres
	resultGenres := make([]string, 0, len(result.Genres))
	for _, genre := range result.Genres {
		g := strings.TrimSpace(genre)
		if len(g) > 0 {
			resultGenres = append(resultGenres, g)
		}
	}

	// // Cleanup primary genre
	// var primaryGenre models.Genre
	// if result.PrimaryGenreName != nil {
	// 	g := strings.TrimSpace(*result.PrimaryGenreName)
	// 	if len(g) > 0 {
	// 		db.Where("name = ?", result.PrimaryGenreName).FirstOrCreate(&primaryGenre)
	// 		if db.Error != nil {
	// 			// db.Rollback()
	// 			return nil, db.Error
	// 		}
	// 		p.PrimaryGenreID = &primaryGenre.ID
	// 	}
	// }
	hasPrimaryGenre := result.PrimaryGenreName != nil && len(*result.PrimaryGenreName) > 0

	// Use genres already in db, create the ones that aren't
	if len(resultGenres) > 0 {
		var genres []models.Genre
		db.Where("name IN ?", resultGenres).Find(&genres)
		if db.Error != nil {
			return nil, db.Error
		}

		unsavedGenres := make([]models.Genre, 0)
		for _, genreName := range resultGenres {
			foundInDb := false
			for _, genre := range genres {
				if *genre.Name == genreName {
					p.PodcastGenres = append(
						p.PodcastGenres,
						models.PodcastGenre{
							GenreID: genre.ID,
						},
					)
					foundInDb = true

					if hasPrimaryGenre {
						isPrimaryGenre := *genre.Name == *result.PrimaryGenreName
						if isPrimaryGenre {

							p.PrimaryGenreID = &genre.ID
						}
					}

					break
				}
			}

			if foundInDb {
				continue
			}

			g := genreName
			unsavedGenres = append(unsavedGenres, models.Genre{
				Name: &g,
			})
		}

		if len(unsavedGenres) > 0 {
			db.Create(unsavedGenres)
			if db.Error != nil {
				return nil, db.Error
			}

			for _, createdGenre := range unsavedGenres {
				p.PodcastGenres = append(
					p.PodcastGenres,
					models.PodcastGenre{
						GenreID: createdGenre.ID,
					},
				)

				if hasPrimaryGenre {
					isPrimaryGenre := *createdGenre.Name == *result.PrimaryGenreName
					if isPrimaryGenre {
						p.PrimaryGenreID = &createdGenre.ID
					}
				}
			}
		}
	}

	return &p, nil
}
