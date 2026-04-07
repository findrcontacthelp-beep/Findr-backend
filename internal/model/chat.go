package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	FirebaseID     *string         `json:"firebase_id,omitempty" db:"firebase_id"`
	LastMessage    *string         `json:"last_message,omitempty" db:"last_message"`
	LastMessageAt  *time.Time      `json:"last_message_at,omitempty" db:"last_message_at"`
	UnreadMessages json.RawMessage `json:"unread_messages" db:"unread_messages"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

type ChatParticipant struct {
	ChatID  uuid.UUID  `json:"chat_id" db:"chat_id"`
	UserUID string     `json:"user_uid" db:"user_uid"`
	UserID  *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
}

type ChatWithParticipants struct {
	Chat
	Participants []ChatParticipant `json:"participants"`
}

type CreateChatRequest struct {
	ParticipantUIDs []string `json:"participant_uids" binding:"required,min=1"`
}
