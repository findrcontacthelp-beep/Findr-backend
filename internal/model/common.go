package model

import "time"

type PaginationParams struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

func (p *PaginationParams) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 20
	}
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type CursorParams struct {
	Before time.Time `form:"before"`
	Limit  int       `form:"limit" binding:"min=1,max=100"`
}

func (c *CursorParams) SetDefaults() {
	if c.Before.IsZero() {
		c.Before = time.Now()
	}
	if c.Limit == 0 {
		c.Limit = 50
	}
}

type ListResponse[T any] struct {
	Data       []T  `json:"data"`
	TotalCount int  `json:"total_count"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	HasMore    bool `json:"has_more"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
