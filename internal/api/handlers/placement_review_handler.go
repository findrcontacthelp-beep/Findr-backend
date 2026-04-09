package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func CreatePlacementReview(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var req model.CreatePlacementReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var submitterName string
		_ = pool.QueryRow(c.Request.Context(), `SELECT COALESCE(name,'') FROM users WHERE id = $1`, userID).Scan(&submitterName)

		var review model.PlacementReview
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO placement_reviews (
				submitted_by_uuid, submitted_by_id, submitted_by_name, company_name, company_logo,
				year, month, academic_year, visit_date, difficulty, overall_experience,
				package_type, package_min, package_max, package_list,
				students_shortlisted, students_selected,
				eligibility_branches, eligibility_cgpa, eligibility_max_backlogs, eligibility_other,
				tips, rounds
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
			 RETURNING id, company_name, submitted_at`,
			userUID, userID, submitterName, req.CompanyName, req.CompanyLogo,
			req.Year, req.Month, req.AcademicYear, req.VisitDate, req.Difficulty, req.OverallExperience,
			req.PackageType, req.PackageMin, req.PackageMax, req.PackageList,
			req.StudentsShortlisted, req.StudentsSelected,
			req.EligibilityBranches, req.EligibilityCGPA, req.EligibilityMaxBklogs, req.EligibilityOther,
			req.Tips, req.Rounds,
		).Scan(&review.ID, &review.CompanyName, &review.SubmittedAt)
		if err != nil {
			log.Error("create placement review failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create review"})
			return
		}

		c.JSON(http.StatusCreated, review)
	}
}

func ListPlacementReviews(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params model.PaginationParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		company := c.Query("company")
		year := c.Query("year")

		query := `SELECT id, submitted_by_name, submitted_at, company_name, company_logo,
		                 year, difficulty, overall_experience, package_type, package_min, package_max,
		                 students_shortlisted, students_selected, verification_status, upvotes
		          FROM placement_reviews WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM placement_reviews WHERE 1=1`
		args := []interface{}{}
		argIdx := 1

		if company != "" {
			query += fmt.Sprintf(` AND company_name ILIKE $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND company_name ILIKE $%d`, argIdx)
			args = append(args, "%"+company+"%")
			argIdx++
		}
		if year != "" {
			query += fmt.Sprintf(` AND year::text = $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND year::text = $%d`, argIdx)
			args = append(args, year)
			argIdx++
		}

		var totalCount int
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		_ = pool.QueryRow(c.Request.Context(), countQuery, countArgs...).Scan(&totalCount)

		query += fmt.Sprintf(` ORDER BY submitted_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, params.Limit, params.Offset())

		rows, err := pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			log.Error("list placement reviews failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list reviews"})
			return
		}
		defer rows.Close()

		var reviews []model.PlacementReview
		for rows.Next() {
			var r model.PlacementReview
			if err := rows.Scan(&r.ID, &r.SubmittedByName, &r.SubmittedAt, &r.CompanyName, &r.CompanyLogo,
				&r.Year, &r.Difficulty, &r.OverallExperience, &r.PackageType, &r.PackageMin, &r.PackageMax,
				&r.StudentsShortlisted, &r.StudentsSelected, &r.VerificationStatus, &r.Upvotes); err != nil {
				log.Error("scan placement review failed", zap.Error(err))
				continue
			}
			reviews = append(reviews, r)
		}

		c.JSON(http.StatusOK, model.ListResponse[model.PlacementReview]{
			Data:       reviews,
			TotalCount: totalCount,
			Page:       params.Page,
			Limit:      params.Limit,
			HasMore:    params.Page*params.Limit < totalCount,
		})
	}
}

func GetPlacementReview(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var r model.PlacementReview
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, submitted_by_uuid, submitted_by_id, submitted_by_name, submitted_at,
			        company_name, company_logo, year, month, academic_year, visit_date,
			        difficulty, overall_experience, package_type, package_min, package_max, package_list,
			        students_shortlisted, students_selected,
			        eligibility_branches, eligibility_cgpa, eligibility_max_backlogs, eligibility_other,
			        tips, rounds, verification_status, verified_at, upvotes
			 FROM placement_reviews WHERE id = $1`, id,
		).Scan(
			&r.ID, &r.SubmittedByUUID, &r.SubmittedByID, &r.SubmittedByName, &r.SubmittedAt,
			&r.CompanyName, &r.CompanyLogo, &r.Year, &r.Month, &r.AcademicYear, &r.VisitDate,
			&r.Difficulty, &r.OverallExperience, &r.PackageType, &r.PackageMin, &r.PackageMax, &r.PackageList,
			&r.StudentsShortlisted, &r.StudentsSelected,
			&r.EligibilityBranches, &r.EligibilityCGPA, &r.EligibilityMaxBklogs, &r.EligibilityOther,
			&r.Tips, &r.Rounds, &r.VerificationStatus, &r.VerifiedAt, &r.Upvotes,
		)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "review not found"})
			return
		}
		if err != nil {
			log.Error("get placement review failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get review"})
			return
		}

		c.JSON(http.StatusOK, r)
	}
}
