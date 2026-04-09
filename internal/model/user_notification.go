package model

import (
	"time"

	"github.com/google/uuid"
)

type UserNotification struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	LastNotifiedAt *time.Time `json:"last_notified_at,omitempty" db:"last_notified_at"`
}
