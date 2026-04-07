package model

import (
	"time"

	"github.com/google/uuid"
)

type UserVerification struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	VerifierUID string    `json:"verifier_uid" db:"verifier_uid"`
	VerifiedAt  time.Time `json:"verified_at" db:"verified_at"`
}
