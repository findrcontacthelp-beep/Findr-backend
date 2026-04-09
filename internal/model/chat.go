package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	LastMessage    *string         `json:"last_message,omitempty" db:"last_message"`
	LastMessageAt  *time.Time      `json:"last_message_at,omitempty" db:"last_message_at"`
	UnreadMessages json.RawMessage `json:"unread_messages" db:"unread_messages"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

type ChatParticipant struct {
	ChatID   uuid.UUID  `json:"chat_id" db:"chat_id"`
	UserUUID string     `json:"user_uuid" db:"user_uuid"`
	UserID   *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
}

type ChatWithParticipants struct {
	Chat
	Participants []ChatParticipant `json:"participants"`
}

type CreateChatRequest struct {
	ParticipantUUIDs []string `json:"participant_uuids" binding:"required,min=1"`
}
