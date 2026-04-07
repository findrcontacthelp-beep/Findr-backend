package model

import "github.com/google/uuid"

type UserRating struct {
	ID       uuid.UUID `json:"id" db:"id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	RaterUID string    `json:"rater_uid" db:"rater_uid"`
	Rating   int       `json:"rating" db:"rating"`
}

type RateUserRequest struct {
	Rating int `json:"rating" binding:"required,min=1,max=5"`
}

type AverageRatingResponse struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
