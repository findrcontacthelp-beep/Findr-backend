package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectComment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ProjectID       uuid.UUID  `json:"project_id" db:"project_id"`
	SenderUUID      *string    `json:"sender_uuid,omitempty" db:"sender_uuid"`
	SenderID        *uuid.UUID `json:"sender_id,omitempty" db:"sender_id"`
	SenderName      string     `json:"sender_name" db:"sender_name"`
	SenderImageURL  string     `json:"sender_image_url" db:"sender_image_url"`
	PostID          *string    `json:"post_id,omitempty" db:"post_id"`
	Text            string     `json:"text" db:"text"`
	NestingLevel    int        `json:"nesting_level" db:"nesting_level"`
	ParentCommentID *string    `json:"parent_comment_id,omitempty" db:"parent_comment_id"`
	RootCommentID   *string    `json:"root_comment_id,omitempty" db:"root_comment_id"`
	IsTopLevel      bool       `json:"is_top_level" db:"is_top_level"`
	ReplyCount      int        `json:"reply_count" db:"reply_count"`
	Likes           []string   `json:"likes" db:"likes"`
	LikesCount      int        `json:"likes_count" db:"likes_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type CreateCommentRequest struct {
	Text            string  `json:"text" binding:"required"`
	ParentCommentID *string `json:"parent_comment_id"`
	RootCommentID   *string `json:"root_comment_id"`
	NestingLevel    int     `json:"nesting_level"`
}
