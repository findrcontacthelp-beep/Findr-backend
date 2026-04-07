package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	FirebaseID  *string         `json:"firebase_id,omitempty" db:"firebase_id"`
	MessageID   *string         `json:"message_id,omitempty" db:"message_id"`
	ChatID      uuid.UUID       `json:"chat_id" db:"chat_id"`
	SenderUID   *string         `json:"sender_uid,omitempty" db:"sender_uid"`
	SenderID    *uuid.UUID      `json:"sender_id,omitempty" db:"sender_id"`
	ReceiverUID *string         `json:"receiver_uid,omitempty" db:"receiver_uid"`
	Message     string          `json:"message" db:"message"`
	Status      string          `json:"status" db:"status"`
	ReplyTo     json.RawMessage `json:"reply_to,omitempty" db:"reply_to"`
	Media       json.RawMessage `json:"media,omitempty" db:"media"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type ReplyTo struct {
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
	SenderUID string `json:"sender_uid"`
}

type MessageMedia struct {
	Type string `json:"type"`
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type SendMessageRequest struct {
	Message     string        `json:"message" binding:"required"`
	ReceiverUID *string       `json:"receiver_uid"`
	ReplyTo     *ReplyTo      `json:"reply_to"`
	Media       *MessageMedia `json:"media"`
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSChatMessage struct {
	ChatID      string        `json:"chat_id"`
	Message     string        `json:"message"`
	ReceiverUID *string       `json:"receiver_uid,omitempty"`
	ReplyTo     *ReplyTo      `json:"reply_to,omitempty"`
	Media       *MessageMedia `json:"media,omitempty"`
}
