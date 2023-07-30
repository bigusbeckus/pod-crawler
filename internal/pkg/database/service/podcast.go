package service

import (
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
)

func PodcastFromItunesResult(result podcast.ItunesResult) (*models.Podcast, error) {
	p := models.Podcast{
		Title:         result.CollectionName,
		CensoredTitle: result.CollectionCensoredName,
		FeedUrl:       result.FeedUrl,
		ReleaseDate:   result.ReleaseDate,
		Description:   "",

		Country:               result.Country,
		EpisodeCount:          0,
		ContentAdvisoryRating: result.ContentAdvisoryRating,

		ItunesID:            uint64(result.CollectionId),
		ItunesViewUrl:       result.CollectionViewUrl,
		ItunesArtworkUrl30:  result.ArtworkUrl30,
		ItunesArtworkUrl60:  result.ArtworkUrl60,
		ItunesArtworkUrl100: result.ArtworkUrl100,
		ItunesArtworkUrl600: result.ArtworkUrl600,
	}

	db, err := database.GetInstance()
	if err != nil {
		return nil, err
	}

	var primaryGenre *models.Genre
	db.Where("Name = ?", result.PrimaryGenreName).First(primaryGenre)
	if primaryGenre != nil {
		p.PrimaryGenreID = primaryGenre.ID
	} else {
		p.PrimaryGenre = models.Genre{
			Name: result.PrimaryGenreName,
		}
	}

	var artist *models.Artist
	db.Where("ItunesID = ?", result.ArtistId).First(artist)
	if artist != nil {
		p.ArtistID = artist.ID
	} else {
		p.Artist = models.Artist{
			Name:          result.ArtistName,
			ItunesID:      uint64(result.ArtistId),
			ItunesViewUrl: result.ArtistViewUrl,
		}
	}

	var genres []models.Genre
	db.Where("name IN ?", result.Genres).Find(&genres)
	for _, genreName := range result.Genres {
		found := false
		for _, genre := range genres {
			if genre.Name == genreName {
				p.PodcastGenres = append(
					p.PodcastGenres,
					models.PodcastGenre{
						GenreID: genre.ID,
					},
				)
				found = true
				break
			}
		}
		if !found {
			p.PodcastGenres = append(
				p.PodcastGenres,
				models.PodcastGenre{
					Genre: models.Genre{
						Name: genreName,
					},
				},
			)
		}
	}

	return &p, err
}
