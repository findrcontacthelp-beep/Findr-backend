package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserUUID         string          `json:"user_uuid" db:"user_uuid"`
	Name             string          `json:"name" db:"name"`
	Email            *string         `json:"email,omitempty" db:"email"`
	ProfilePicture   string          `json:"profile_picture" db:"profile_picture"`
	FCMToken         *string         `json:"-" db:"fcm_token"`
	Headline         string          `json:"headline" db:"headline"`
	RoleTitle        string          `json:"role_title" db:"role_title"`
	IsStudent        bool            `json:"is_student" db:"is_student"`
	CollegeName      string          `json:"college_name" db:"college_name"`
	CompanyName      string          `json:"company_name" db:"company_name"`
	Experience       string          `json:"experience" db:"experience"`
	CTC              string          `json:"ctc" db:"ctc"`
	Location         string          `json:"location" db:"location"`
	Lat              float32         `json:"lat" db:"lat"`
	Lng              float32         `json:"lng" db:"lng"`
	ProfileImageURL  string          `json:"profile_image_url" db:"profile_image_url"`
	BannerImageURL   string          `json:"banner_image_url" db:"banner_image_url"`
	Skills           []string        `json:"skills" db:"skills"`
	SocialLinks      json.RawMessage `json:"social_links" db:"social_links"`
	CollegeYear      string          `json:"graduation_year" db:"college_year"`
	CollegeStream    string          `json:"branch" db:"college_stream"`
	CollegeGrade     string          `json:"college_grade" db:"college_grade"`
	CollegeStart     string          `json:"college_start" db:"college_start"`
	CollegeEnd       string          `json:"college_end" db:"college_end"`
	CollegeInstitute string          `json:"college_institute" db:"college_institute"`
	ExpTitle         string          `json:"exp_title" db:"exp_title"`
	ExpCompany       string          `json:"exp_company" db:"exp_company"`
	ExpType          string          `json:"exp_type" db:"exp_type"`
	ExpLocation      string          `json:"exp_location" db:"exp_location"`
	ExpDescription   string          `json:"exp_description" db:"exp_description"`
	ExpCTC           string          `json:"exp_ctc" db:"exp_ctc"`
	ExpStart         string          `json:"exp_start" db:"exp_start"`
	ExpEnd           string          `json:"exp_end" db:"exp_end"`
	ExpCurrently     bool            `json:"exp_currently_working" db:"exp_currently_working"`
	AboutText        string          `json:"about_text" db:"about_text"`
	Activities       json.RawMessage `json:"activities" db:"activities"`
	Interests        []string        `json:"interests" db:"interests"`
	UserList         []string        `json:"user_list" db:"user_list"`
	Stability        int             `json:"stability" db:"stability"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Name          string   `json:"name" binding:"required"`
	Email         string   `json:"email" binding:"required,email"`
	CollegeName   string   `json:"college_name" binding:"required"`
	CollegeStream string   `json:"branch" binding:"required"`
	CollegeYear   string   `json:"graduation_year" binding:"required"`
	Interests     []string `json:"interests" binding:"required,min=1"`
}

type UpdateUserRequest struct {
	Name             *string          `json:"name"`
	Headline         *string          `json:"headline"`
	RoleTitle        *string          `json:"role_title"`
	IsStudent        *bool            `json:"is_student"`
	CollegeName      *string          `json:"college_name"`
	CompanyName      *string          `json:"company_name"`
	Experience       *string          `json:"experience"`
	CTC              *string          `json:"ctc"`
	Location         *string          `json:"location"`
	Lat              *float32         `json:"lat"`
	Lng              *float32         `json:"lng"`
	ProfileImageURL  *string          `json:"profile_image_url"`
	BannerImageURL   *string          `json:"banner_image_url"`
	Skills           *[]string        `json:"skills"`
	SocialLinks      *json.RawMessage `json:"social_links"`
	CollegeYear      *string          `json:"graduation_year"`
	CollegeStream    *string          `json:"branch"`
	CollegeGrade     *string          `json:"college_grade"`
	CollegeStart     *string          `json:"college_start"`
	CollegeEnd       *string          `json:"college_end"`
	CollegeInstitute *string          `json:"college_institute"`
	ExpTitle         *string          `json:"exp_title"`
	ExpCompany       *string          `json:"exp_company"`
	ExpType          *string          `json:"exp_type"`
	ExpLocation      *string          `json:"exp_location"`
	ExpDescription   *string          `json:"exp_description"`
	ExpCTC           *string          `json:"exp_ctc"`
	ExpStart         *string          `json:"exp_start"`
	ExpEnd           *string          `json:"exp_end"`
	ExpCurrently     *bool            `json:"exp_currently_working"`
	AboutText        *string          `json:"about_text"`
	Activities       *json.RawMessage `json:"activities"`
	Interests        *[]string        `json:"interests"`
}
