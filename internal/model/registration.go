package model

import (
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	FirebaseID          *string    `json:"firebase_id,omitempty" db:"firebase_id"`
	UserUID             *string    `json:"user_uid,omitempty" db:"user_uid"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	UserName            string     `json:"user_name" db:"user_name"`
	UserProfilePic      string     `json:"user_profile_pic" db:"user_profile_pic"`
	PostID              *string    `json:"post_id,omitempty" db:"post_id"`
	ProjectID           *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	Status              string     `json:"status" db:"status"`
	Registered          bool       `json:"registered" db:"registered"`
	Attended            bool       `json:"attended" db:"attended"`
	AttendanceConfirmed bool       `json:"attendance_confirmed" db:"attendance_confirmed"`
	Cancelled           bool       `json:"cancelled" db:"cancelled"`
	RegisteredAt        time.Time  `json:"registered_at" db:"registered_at"`
}
