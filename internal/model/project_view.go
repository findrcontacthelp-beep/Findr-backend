package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectPostView struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	ViewerUID string    `json:"viewer_uid" db:"viewer_uid"`
	ViewedAt  time.Time `json:"viewed_at" db:"viewed_at"`
}
