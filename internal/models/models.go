package models

import "gorm.io/gorm"

type ShortLink struct {
	gorm.Model
	OriginalURL string `gorm:"not null"`
	Hash        string `gorm:"uniqueIndex;not null;size:10"`
	Password    string `gorm:"size:255"`
	Clicks      int    `gorm:"default:0"`
}

type ContactMessage struct {
	gorm.Model
	Name    string `gorm:"not null"`
	Email   string `gorm:"not null"`
	Subject string `gorm:"not null"`
	Message string `gorm:"not null"`
}
