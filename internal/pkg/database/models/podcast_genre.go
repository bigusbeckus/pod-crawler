package models

type PodcastGenre struct {
	Model

	PodcastID string `gorm:"not null;index:idx_podcast_genre,unique"`
	GenreID   string `gorm:"not null;index:idx_podcast_genre,unique"`

	Podcast Podcast
	Genre   Genre
}
