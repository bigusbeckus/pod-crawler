package models

type Artist struct {
	Model

  Name          string `gorm:"index:,type:btree"`
	ItunesId      int    `gorm:"unique"`
	ItunesViewUrl string `gorm:"unique"`
}
