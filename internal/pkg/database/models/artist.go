package models

type Artist struct {
	Model

	Name          string `gorm:"index:,type:btree"`
	ItunesID      uint64 `gorm:"unique"`
	ItunesViewUrl string `gorm:"unique"`
}
