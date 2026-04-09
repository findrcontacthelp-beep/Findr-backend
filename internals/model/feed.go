package model

import "time"

type ProjectRole struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Openings    int      `json:"openings,omitempty"`
}

type FeedItem struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	AuthorName       string            `json:"author_name"`
	AuthorUID        string            `json:"author_uid,omitempty"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Tags             []string          `json:"tags"`
	ImageURLs        []string          `json:"image_urls"`
	FileURLs         []string          `json:"file_urls"`
	VideoURL         string            `json:"video_url"`
	Links            []string          `json:"links"`
	RolesNeeded      []string          `json:"roles_needed"`
	ProjectRoles     []ProjectRole     `json:"project_roles,omitempty"`
	EnrolledPersons  map[string]string `json:"enrolled_persons,omitempty"`
	RequestedPersons map[string]string `json:"requested_persons,omitempty"`
	LikesCount       int               `json:"likes_count"`
	CommentsCount    int               `json:"comments_count"`
	ViewsCount       int               `json:"views_count"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
