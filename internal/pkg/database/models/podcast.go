package models

type Podcast struct {
	Model

	Title         string `gorm:"not null;index:,type:btree"`
	CensoredTitle string `gorm:"not null;index:,type:btree"`
	FeedUrl       string `gorm:"unique"`
	ReleaseDate   string
	Description   string

	Country               string
	EpisodeCount          int
	ContentAdvisoryRating string

	ItunesID            uint64 `gorm:"unique"`
	ItunesViewUrl       string `gorm:"unique"`
	ItunesArtworkUrl30  string
	ItunesArtworkUrl60  string
	ItunesArtworkUrl100 string
	ItunesArtworkUrl600 string

	PrimaryGenreID string
	ArtistID       string `gorm:"not null"`

	PrimaryGenre  Genre `gorm:"foreignKey:PrimaryGenreID"`
	Artist        Artist
	PodcastGenres []PodcastGenre
}
