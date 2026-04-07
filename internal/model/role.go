package model

import "github.com/google/uuid"

type AvailableRole struct {
	ID         uuid.UUID `json:"id" db:"id"`
	FirebaseID *string   `json:"firebase_id,omitempty" db:"firebase_id"`
	Name       string    `json:"name" db:"name"`
}

type RoleRequest struct {
	ID         uuid.UUID `json:"id" db:"id"`
	FirebaseID *string   `json:"firebase_id,omitempty" db:"firebase_id"`
	Name       string    `json:"name" db:"name"`
}

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required"`
}
