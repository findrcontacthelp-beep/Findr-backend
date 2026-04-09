package model

import "github.com/google/uuid"

type Topic struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Topic      string    `json:"topic" db:"topic"`
	Enabled    bool      `json:"enabled" db:"enabled"`
}

type CreateTopicRequest struct {
	Topic   string `json:"topic" binding:"required"`
	Enabled bool   `json:"enabled"`
}
