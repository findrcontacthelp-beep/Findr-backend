package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PlacementReview struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	SubmittedByUUID      *string         `json:"submitted_by_uuid,omitempty" db:"submitted_by_uuid"`
	SubmittedByID        *uuid.UUID      `json:"submitted_by_id,omitempty" db:"submitted_by_id"`
	SubmittedByName      string          `json:"submitted_by_name" db:"submitted_by_name"`
	SubmittedAt          time.Time       `json:"submitted_at" db:"submitted_at"`
	CompanyName          string          `json:"company_name" db:"company_name"`
	CompanyLogo          *string         `json:"company_logo,omitempty" db:"company_logo"`
	Year                 *int            `json:"year,omitempty" db:"year"`
	Month                *string         `json:"month,omitempty" db:"month"`
	AcademicYear         *string         `json:"academic_year,omitempty" db:"academic_year"`
	VisitDate            *time.Time      `json:"visit_date,omitempty" db:"visit_date"`
	Difficulty           *string         `json:"difficulty,omitempty" db:"difficulty"`
	OverallExperience    *string         `json:"overall_experience,omitempty" db:"overall_experience"`
	PackageType          *string         `json:"package_type,omitempty" db:"package_type"`
	PackageMin           *float32        `json:"package_min,omitempty" db:"package_min"`
	PackageMax           *float32        `json:"package_max,omitempty" db:"package_max"`
	PackageList          json.RawMessage `json:"package_list,omitempty" db:"package_list"`
	StudentsShortlisted  int             `json:"students_shortlisted" db:"students_shortlisted"`
	StudentsSelected     int             `json:"students_selected" db:"students_selected"`
	EligibilityBranches  []string        `json:"eligibility_branches" db:"eligibility_branches"`
	EligibilityCGPA      *float32        `json:"eligibility_cgpa,omitempty" db:"eligibility_cgpa"`
	EligibilityMaxBklogs *int            `json:"eligibility_max_backlogs,omitempty" db:"eligibility_max_backlogs"`
	EligibilityOther     *string         `json:"eligibility_other,omitempty" db:"eligibility_other"`
	Tips                 []string        `json:"tips" db:"tips"`
	Rounds               json.RawMessage `json:"rounds" db:"rounds"`
	VerificationStatus   string          `json:"verification_status" db:"verification_status"`
	VerifiedAt           *time.Time      `json:"verified_at,omitempty" db:"verified_at"`
	Upvotes              int             `json:"upvotes" db:"upvotes"`
}

type CreatePlacementReviewRequest struct {
	CompanyName          string          `json:"company_name" binding:"required"`
	CompanyLogo          *string         `json:"company_logo"`
	Year                 *int            `json:"year"`
	Month                *string         `json:"month"`
	AcademicYear         *string         `json:"academic_year"`
	VisitDate            *time.Time      `json:"visit_date"`
	Difficulty           *string         `json:"difficulty"`
	OverallExperience    *string         `json:"overall_experience"`
	PackageType          *string         `json:"package_type"`
	PackageMin           *float32        `json:"package_min"`
	PackageMax           *float32        `json:"package_max"`
	PackageList          json.RawMessage `json:"package_list"`
	StudentsShortlisted  int             `json:"students_shortlisted"`
	StudentsSelected     int             `json:"students_selected"`
	EligibilityBranches  []string        `json:"eligibility_branches"`
	EligibilityCGPA      *float32        `json:"eligibility_cgpa"`
	EligibilityMaxBklogs *int            `json:"eligibility_max_backlogs"`
	EligibilityOther     *string         `json:"eligibility_other"`
	Tips                 []string        `json:"tips"`
	Rounds               json.RawMessage `json:"rounds"`
}
