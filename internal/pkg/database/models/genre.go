package models

type Genre struct {
	Model

	Name string `gorm:"not null;index:,unique,type:btree"`
}
