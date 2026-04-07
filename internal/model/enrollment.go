package model

import (
	"time"

	"github.com/google/uuid"
)

type Enrollment struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	FirebaseID     *string    `json:"firebase_id,omitempty" db:"firebase_id"`
	UserUID        *string    `json:"user_uid,omitempty" db:"user_uid"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	UserName       string     `json:"user_name" db:"user_name"`
	UserProfilePic string     `json:"user_profile_pic" db:"user_profile_pic"`
	PostID         *string    `json:"post_id,omitempty" db:"post_id"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	RoleID         *string    `json:"role_id,omitempty" db:"role_id"`
	RoleName       string     `json:"role_name" db:"role_name"`
	Message        *string    `json:"message,omitempty" db:"message"`
	Status         string     `json:"status" db:"status"`
	Pending        bool       `json:"pending" db:"pending"`
	Accepted       bool       `json:"accepted" db:"accepted"`
	Rejected       bool       `json:"rejected" db:"rejected"`
	RequestedAt    time.Time  `json:"requested_at" db:"requested_at"`
	RespondedAt    *time.Time `json:"responded_at,omitempty" db:"responded_at"`
}

type EnrollRequest struct {
	RoleID   string  `json:"role_id"`
	RoleName string  `json:"role_name" binding:"required"`
	Message  *string `json:"message"`
}

type RespondEnrollmentRequest struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
}
