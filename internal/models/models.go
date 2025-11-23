package models

import "gorm.io/gorm"

type ShortLink struct {
	gorm.Model
	OriginalURL string `gorm:"not null"`
	Hash        string `gorm:"uniqueIndex;not null;size:10"`
	Password    string `gorm:"size:255"`
	Clicks      int    `gorm:"default:0"`
}
