package models

import "gorm.io/gorm"

type PodcastData struct {
	WrapperType            string  `json:"wrapperType"`
	Kind                   string  `json:"kind"`
	ArtistId               int     `json:"artistId"`
	TrackId                int     `json:"trackId"`
	ArtistName             string  `json:"artistName"`
	CollectionName         string  `json:"collectionName"`
	TrackName              string  `json:"trackName"`
	CollectionCensoredName string  `json:"collectionCensoredName"`
	TrackCensoredName      string  `json:"trackCensoredName"`
	ArtistViewUrl          string  `json:"artistViewUrl"`
	CollectionViewUrl      string  `json:"collectionViewUrl"`
	FeedUrl                string  `json:"feedUrl"`
	TrackViewUrl           string  `json:"trackViewUrl"`
	ArtworkUrl30           string  `json:"artworkUrl30"`
	ArtworkUrl60           string  `json:"artworkUrl60"`
	ArtworkUrl100          string  `json:"artworkUrl100"`
	CollectionPrice        float64 `json:"collectionPrice"`
	TrackPrice             float64 `json:"trackPrice"`
	TrackRentalPrice       float64 `json:"trackRentalPrice"`
	CollectionHdPrice      float64 `json:"collectionHdPrice"`
	TrackHdPrice           float64 `json:"trackHdPrice"`
	TrackHdRentalPrice     float64 `json:"trackHdRentalPrice"`
	ReleaseDate            string  `json:"releaseDate"`
	CollectionExplicitness string  `json:"collectionExplicitness"`
	TrackExplicitness      string  `json:"trackExplicitness"`
	TrackCount             int     `json:"trackCount"`
	Country                string  `json:"country"`
	Currency               string  `json:"currency"`
	PrimaryGenreName       string  `json:"primaryGenreName"`
	ContentAdvisoryRating  string  `json:"contentAdvisoryRating"`
	ArtworkUrl600          string  `json:"artworkUrl600"`
}

type Genre struct {
	gorm.Model
	ID   int
	Name string
}

// type PodcastGenres struct {
// 	gorm.Model
// 	PodcastId int
// 	Genre     string
// }

type PodcastModel struct {
	gorm.Model
	PodcastData
	ID     int
	Genres []Genre `gorm:"foreignKey:ID"`
}

type PodcastEntry struct {
	PodcastData
	CollectionId int      `json:"collectionId"`
	Genres       []string `json:"genres"`
}
