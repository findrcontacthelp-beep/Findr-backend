package model

type RegisterRequest struct {
	Name           string   `json:"name" binding:"required"`
	Email          string   `json:"email" binding:"required,email"`
	Password       string   `json:"password,omitempty" binding:"omitempty,min=8"`
	CollegeName    string   `json:"college_name" binding:"required"`
	Branch         string   `json:"branch" binding:"required"`
	GraduationYear string   `json:"graduation_year" binding:"required"`
	Interests      []string `json:"interests" binding:"required,min=1"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        interface{} `json:"user"`
}
