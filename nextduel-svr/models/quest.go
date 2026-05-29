package models

import "time"

type Quest struct {
	ID         string    `gorm:"primaryKey;size:64" json:"id"`
	Question   string    `gorm:"not null" json:"question"`
	Answer     string    `gorm:"not null" json:"answer"`
	Difficulty int       `gorm:"default:1" json:"difficulty"`
	Category   string    `json:"category"`
	CreatedAt  time.Time `json:"created_at"`
}
