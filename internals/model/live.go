package model

import "time"

type LiveEventSummary struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Subtitle         string    `json:"subtitle"` // mapped from description
	ParticipantCount int       `json:"participant_count"` // mapped from views or similar
	CreatedAt        time.Time `json:"created_at"`
	ImageUrl         string    `json:"image_url"`
}
