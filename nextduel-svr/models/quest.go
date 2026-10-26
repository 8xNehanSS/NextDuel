package models

import "time"

// Quest is the local question bank. It is seeded from questions/seed.json and
// grows with every question fetched from OpenTDB, so matches keep working when
// the external API is rate-limited or down.
type Quest struct {
	// ID is a content hash of the question text, which de-duplicates entries.
	ID       string `gorm:"primaryKey;size:64" json:"id"`
	Question string `gorm:"not null" json:"question"`
	Answer   string `gorm:"not null" json:"answer"`
	// IncorrectAnswers is a JSON-encoded []string.
	IncorrectAnswers string    `gorm:"type:text" json:"incorrect_answers"`
	Difficulty       string    `gorm:"size:20;index" json:"difficulty"`
	Category         string    `gorm:"size:100" json:"category"`
	Source           string    `gorm:"size:20" json:"source"`
	CreatedAt        time.Time `json:"created_at"`
}
