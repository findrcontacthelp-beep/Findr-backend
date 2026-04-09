package model

import "github.com/google/uuid"

type UserExtraActivity struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Domain      string    `json:"domain" db:"domain"`
	Link        string    `json:"link" db:"link"`
	Description string    `json:"description" db:"description"`
	Media       string    `json:"media" db:"media"`
}

type CreateExtraActivityRequest struct {
	Name        string `json:"name" binding:"required"`
	Domain      string `json:"domain"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Media       string `json:"media"`
}
