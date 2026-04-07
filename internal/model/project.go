package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	FirebaseID       *string         `json:"firebase_id,omitempty" db:"firebase_id"`
	AuthorUID        *string         `json:"author_uid,omitempty" db:"author_uid"`
	AuthorID         *uuid.UUID      `json:"author_id,omitempty" db:"author_id"`
	AuthorName       string          `json:"author_name" db:"author_name"`
	Type             string          `json:"type" db:"type"`
	Title            string          `json:"title" db:"title"`
	TitleLower       string          `json:"title_lower" db:"title_lower"`
	Description      string          `json:"description" db:"description"`
	Tags             []string        `json:"tags" db:"tags"`
	ImageURLs        []string        `json:"image_urls" db:"image_urls"`
	FileURLs         []string        `json:"file_urls" db:"file_urls"`
	VideoURL         string          `json:"video_url" db:"video_url"`
	Links            []string        `json:"links" db:"links"`
	RolesNeeded      []string        `json:"roles_needed" db:"roles_needed"`
	ProjectRoles     json.RawMessage `json:"project_roles" db:"project_roles"`
	EventDetails     json.RawMessage `json:"event_details,omitempty" db:"event_details"`
	EnrolledPersons  json.RawMessage `json:"enrolled_persons" db:"enrolled_persons"`
	RequestedPersons json.RawMessage `json:"requested_persons" db:"requested_persons"`
	Likes            []string        `json:"likes" db:"likes"`
	LikesCount       int             `json:"likes_count" db:"likes_count"`
	CommentsCount    int             `json:"comments_count" db:"comments_count"`
	ViewsCount       int             `json:"views_count" db:"views_count"`
	PostViewsCount   int             `json:"post_views_count" db:"post_views_count"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

type CreateProjectRequest struct {
	Type         string          `json:"type" binding:"required,oneof=Post Project Event"`
	Title        string          `json:"title" binding:"required"`
	Description  string          `json:"description"`
	Tags         []string        `json:"tags"`
	ImageURLs    []string        `json:"image_urls"`
	FileURLs     []string        `json:"file_urls"`
	VideoURL     string          `json:"video_url"`
	Links        []string        `json:"links"`
	RolesNeeded  []string        `json:"roles_needed"`
	ProjectRoles json.RawMessage `json:"project_roles"`
	EventDetails json.RawMessage `json:"event_details"`
}

type ProjectRole struct {
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
	Openings int    `json:"openings"`
	Filled   int    `json:"filled"`
}

type EventDetails struct {
	EventType       string `json:"event_type"`
	Date            string `json:"date"`
	Time            string `json:"time"`
	Location        string `json:"location,omitempty"`
	MeetingLink     string `json:"meeting_link,omitempty"`
	MaxParticipants int    `json:"max_participants"`
}
