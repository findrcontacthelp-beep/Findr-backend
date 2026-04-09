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

func GetUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var u model.User
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, firebase_uid, name, email, profile_picture, headline, role_title,
			        is_student, college_name, company_name, experience, ctc, location,
			        lat, lng, profile_image_url, banner_image_url, skills, social_links,
			        college_year, college_stream, college_grade, college_start, college_end, college_institute,
			        exp_title, exp_company, exp_type, exp_location, exp_description,
			        exp_ctc, exp_start, exp_end, exp_currently_working,
			        about_text, activities, interests, user_list, stability, created_at, updated_at
			 FROM users WHERE id = $1`, id,
		).Scan(
			&u.ID, &u.FirebaseUID, &u.Name, &u.Email, &u.ProfilePicture, &u.Headline, &u.RoleTitle,
			&u.IsStudent, &u.CollegeName, &u.CompanyName, &u.Experience, &u.CTC, &u.Location,
			&u.Lat, &u.Lng, &u.ProfileImageURL, &u.BannerImageURL, &u.Skills, &u.SocialLinks,
			&u.CollegeYear, &u.CollegeStream, &u.CollegeGrade, &u.CollegeStart, &u.CollegeEnd, &u.CollegeInstitute,
			&u.ExpTitle, &u.ExpCompany, &u.ExpType, &u.ExpLocation, &u.ExpDescription,
			&u.ExpCTC, &u.ExpStart, &u.ExpEnd, &u.ExpCurrently,
			&u.AboutText, &u.Activities, &u.Interests, &u.UserList, &u.Stability, &u.CreatedAt, &u.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "user not found"})
			return
		}
		if err != nil {
			log.Error("get user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get user"})
			return
		}

		c.JSON(http.StatusOK, u)
	}
}

func GetCurrentUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		var u model.User
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, firebase_uid, name, email, profile_picture, headline, role_title,
			        is_student, college_name, company_name, experience, ctc, location,
			        lat, lng, profile_image_url, banner_image_url, skills, social_links,
			        college_year, college_stream, college_grade, college_start, college_end, college_institute,
			        exp_title, exp_company, exp_type, exp_location, exp_description,
			        exp_ctc, exp_start, exp_end, exp_currently_working,
			        about_text, activities, interests, user_list, stability, created_at, updated_at
			 FROM users WHERE firebase_uid = $1`, userUID,
		).Scan(
			&u.ID, &u.FirebaseUID, &u.Name, &u.Email, &u.ProfilePicture, &u.Headline, &u.RoleTitle,
			&u.IsStudent, &u.CollegeName, &u.CompanyName, &u.Experience, &u.CTC, &u.Location,
			&u.Lat, &u.Lng, &u.ProfileImageURL, &u.BannerImageURL, &u.Skills, &u.SocialLinks,
			&u.CollegeYear, &u.CollegeStream, &u.CollegeGrade, &u.CollegeStart, &u.CollegeEnd, &u.CollegeInstitute,
			&u.ExpTitle, &u.ExpCompany, &u.ExpType, &u.ExpLocation, &u.ExpDescription,
			&u.ExpCTC, &u.ExpStart, &u.ExpEnd, &u.ExpCurrently,
			&u.AboutText, &u.Activities, &u.Interests, &u.UserList, &u.Stability, &u.CreatedAt, &u.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "user not found"})
			return
		}
		if err != nil {
			log.Error("get current user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get user"})
			return
		}

		c.JSON(http.StatusOK, u)
	}
}

func UpdateUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		var req model.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`UPDATE users SET
				name = COALESCE($2, name),
				headline = COALESCE($3, headline),
				role_title = COALESCE($4, role_title),
				is_student = COALESCE($5, is_student),
				college_name = COALESCE($6, college_name),
				company_name = COALESCE($7, company_name),
				experience = COALESCE($8, experience),
				ctc = COALESCE($9, ctc),
				location = COALESCE($10, location),
				lat = COALESCE($11, lat),
				lng = COALESCE($12, lng),
				profile_image_url = COALESCE($13, profile_image_url),
				banner_image_url = COALESCE($14, banner_image_url),
				skills = COALESCE($15, skills),
				social_links = COALESCE($16, social_links),
				college_year = COALESCE($17, college_year),
				college_stream = COALESCE($18, college_stream),
				college_grade = COALESCE($19, college_grade),
				college_start = COALESCE($20, college_start),
				college_end = COALESCE($21, college_end),
				college_institute = COALESCE($22, college_institute),
				exp_title = COALESCE($23, exp_title),
				exp_company = COALESCE($24, exp_company),
				exp_type = COALESCE($25, exp_type),
				exp_location = COALESCE($26, exp_location),
				exp_description = COALESCE($27, exp_description),
				exp_ctc = COALESCE($28, exp_ctc),
				exp_start = COALESCE($29, exp_start),
				exp_end = COALESCE($30, exp_end),
				exp_currently_working = COALESCE($31, exp_currently_working),
				about_text = COALESCE($32, about_text),
				activities = COALESCE($33, activities),
				interests = COALESCE($34, interests)
			 WHERE firebase_uid = $1`,
			userUID, req.Name, req.Headline, req.RoleTitle, req.IsStudent,
			req.CollegeName, req.CompanyName, req.Experience, req.CTC,
			req.Location, req.Lat, req.Lng, req.ProfileImageURL, req.BannerImageURL,
			req.Skills, req.SocialLinks,
			req.CollegeYear, req.CollegeStream, req.CollegeGrade, req.CollegeStart,
			req.CollegeEnd, req.CollegeInstitute,
			req.ExpTitle, req.ExpCompany, req.ExpType, req.ExpLocation, req.ExpDescription,
			req.ExpCTC, req.ExpStart, req.ExpEnd, req.ExpCurrently,
			req.AboutText, req.Activities, req.Interests,
		)
		if err != nil {
			log.Error("update user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update user"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "user updated"})
	}
}

func ListUsers(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params model.PaginationParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		search := c.Query("search")
		skill := c.Query("skill")

		query := `SELECT id, firebase_uid, name, email, profile_picture, headline, role_title,
		                 is_student, college_name, company_name, location,
		                 profile_image_url, skills, interests, created_at
		          FROM users WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
		args := []interface{}{}
		argIdx := 1

		if search != "" {
			query += ` AND (name ILIKE $` + fmt.Sprint(argIdx) + ` OR headline ILIKE $` + fmt.Sprint(argIdx) + `)`
			countQuery += ` AND (name ILIKE $` + fmt.Sprint(argIdx) + ` OR headline ILIKE $` + fmt.Sprint(argIdx) + `)`
			args = append(args, "%"+search+"%")
			argIdx++
		}
		if skill != "" {
			query += ` AND $` + fmt.Sprint(argIdx) + ` = ANY(skills)`
			countQuery += ` AND $` + fmt.Sprint(argIdx) + ` = ANY(skills)`
			args = append(args, skill)
			argIdx++
		}

		var totalCount int
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		_ = pool.QueryRow(c.Request.Context(), countQuery, countArgs...).Scan(&totalCount)

		query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(argIdx) + ` OFFSET $` + fmt.Sprint(argIdx+1)
		args = append(args, params.Limit, params.Offset())

		rows, err := pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			log.Error("list users failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list users"})
			return
		}
		defer rows.Close()

		var users []model.User
		for rows.Next() {
			var u model.User
			if err := rows.Scan(
				&u.ID, &u.FirebaseUID, &u.Name, &u.Email, &u.ProfilePicture, &u.Headline, &u.RoleTitle,
				&u.IsStudent, &u.CollegeName, &u.CompanyName, &u.Location,
				&u.ProfileImageURL, &u.Skills, &u.Interests, &u.CreatedAt,
			); err != nil {
				log.Error("scan user failed", zap.Error(err))
				continue
			}
			users = append(users, u)
		}

		c.JSON(http.StatusOK, model.ListResponse[model.User]{
			Data:       users,
			TotalCount: totalCount,
			Page:       params.Page,
			Limit:      params.Limit,
			HasMore:    params.Page*params.Limit < totalCount,
		})
	}
}

func UpdateSkills(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		var req struct {
			Skills []string `json:"skills" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`UPDATE users SET skills = $1 WHERE firebase_uid = $2`,
			req.Skills, userUID,
		)
		if err != nil {
			log.Error("update skills failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update skills"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "skills updated"})
	}
}

func UpdateSocialLinks(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		var req struct {
			SocialLinks map[string]string `json:"social_links" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`UPDATE users SET social_links = $1 WHERE firebase_uid = $2`,
			req.SocialLinks, userUID,
		)
		if err != nil {
			log.Error("update social links failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update social links"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "social links updated"})
	}
}

func UpdateFCMToken(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		var req struct {
			FCMToken string `json:"fcm_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`UPDATE users SET fcm_token = $1 WHERE firebase_uid = $2`,
			req.FCMToken, userUID,
		)
		if err != nil {
			log.Error("update fcm token failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update fcm token"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "fcm token updated"})
	}
}

func DeleteUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		_, err := pool.Exec(c.Request.Context(),
			`DELETE FROM users WHERE firebase_uid = $1`, userUID,
		)
		if err != nil {
			log.Error("delete user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "user deleted"})
	}
}

