package models

type Podcast struct {
	Model

	Title         string  `gorm:"not null;index:,type:btree"`
	CensoredTitle string  `gorm:"not null;index:,type:btree"`
	FeedUrl       *string // `gorm:"not null;unique"`
	ArtistName    *string
	ReleaseDate   *string
	Description   *string

	Country               *string
	EpisodeCount          *uint32
	ContentAdvisoryRating *string

	ItunesID            *uint32 // `gorm:"unique"`
	ItunesViewUrl       *string // `gorm:"unique"`
	ItunesArtworkUrl30  *string
	ItunesArtworkUrl60  *string
	ItunesArtworkUrl100 *string
	ItunesArtworkUrl600 *string

	ItunesArtistId      *uint32 `gorm:"default:null"`
	ItunesArtistViewUrl *string `gorm:"default:null"`

	PrimaryGenreID *string

	PrimaryGenre  *Genre `gorm:"foreignKey:PrimaryGenreID"`
	PodcastGenres []PodcastGenre
}
