package models

type PodcastGenre struct {
	PodcastID string `gorm:"primaryKey"`
	GenreID   string `gorm:"primaryKey"`

	Podcast Podcast
	Genre   Genre
}
