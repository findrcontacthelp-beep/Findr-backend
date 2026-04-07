package model

import (
	"time"

	"github.com/google/uuid"
)

type UserViewer struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ViewerUID string    `json:"viewer_uid" db:"viewer_uid"`
	ViewedAt  time.Time `json:"viewed_at" db:"viewed_at"`
}
