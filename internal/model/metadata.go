package model

import (
	"encoding/json"
	"time"
)

type Metadata struct {
	Key       string          `json:"key" db:"key"`
	Value     json.RawMessage `json:"value" db:"value"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type SetMetadataRequest struct {
	Value json.RawMessage `json:"value" binding:"required"`
}
