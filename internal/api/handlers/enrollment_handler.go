package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func ApplyToProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		userUUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var req model.EnrollRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var userName, profilePic string
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT COALESCE(name,''), COALESCE(profile_image_url,'') FROM users WHERE id = $1`, userID,
		).Scan(&userName, &profilePic)

		var enrollment model.Enrollment
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO enrollments (user_uuid, user_id, user_name, user_profile_pic, project_id, role_id, role_name, message, status, pending)
			 VALUES ($1, $2, $3, $4, $5::uuid, $6, $7, $8, 'PENDING', true)
			 RETURNING id, status, requested_at`,
			userUUID, userID, userName, profilePic, projectID, req.RoleID, req.RoleName, req.Message,
		).Scan(&enrollment.ID, &enrollment.Status, &enrollment.RequestedAt)
		if err != nil {
			log.Error("apply to post failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to apply"})
			return
		}

		c.JSON(http.StatusCreated, enrollment)
	}
}

func GetProjectEnrollments(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_uuid, user_id, user_name, user_profile_pic, project_id,
			        role_id, role_name, message, status, pending, accepted, rejected, requested_at, responded_at
			 FROM enrollments WHERE project_id = $1::uuid ORDER BY requested_at DESC`, projectID,
		)
		if err != nil {
			log.Error("get post enrollments failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get enrollments"})
			return
		}
		defer rows.Close()

		var enrollments []model.Enrollment
		for rows.Next() {
			var e model.Enrollment
			if err := rows.Scan(&e.ID, &e.UserUUID, &e.UserID, &e.UserName, &e.UserProfilePic, &e.ProjectID,
				&e.RoleID, &e.RoleName, &e.Message, &e.Status, &e.Pending, &e.Accepted, &e.Rejected,
				&e.RequestedAt, &e.RespondedAt); err != nil {
				continue
			}
			enrollments = append(enrollments, e)
		}

		c.JSON(http.StatusOK, enrollments)
	}
}

func GetMyEnrollments(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_uuid, user_id, user_name, user_profile_pic, project_id,
			        role_id, role_name, message, status, pending, accepted, rejected, requested_at, responded_at
			 FROM enrollments WHERE user_id = $1 ORDER BY requested_at DESC`, userID,
		)
		if err != nil {
			log.Error("get my enrollments failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get enrollments"})
			return
		}
		defer rows.Close()

		var enrollments []model.Enrollment
		for rows.Next() {
			var e model.Enrollment
			if err := rows.Scan(&e.ID, &e.UserUUID, &e.UserID, &e.UserName, &e.UserProfilePic, &e.ProjectID,
				&e.RoleID, &e.RoleName, &e.Message, &e.Status, &e.Pending, &e.Accepted, &e.Rejected,
				&e.RequestedAt, &e.RespondedAt); err != nil {
				continue
			}
			enrollments = append(enrollments, e)
		}

		c.JSON(http.StatusOK, enrollments)
	}
}

func AcceptEnrollment(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		enrollmentID := c.Param("id")
		now := time.Now()

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE enrollments SET status = 'ACCEPTED', accepted = true, pending = false, responded_at = $1
			 WHERE id = $2::uuid`, now, enrollmentID,
		)
		if err != nil {
			log.Error("accept enrollment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to accept enrollment"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "enrollment not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "enrollment accepted"})
	}
}

func RejectEnrollment(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		enrollmentID := c.Param("id")
		now := time.Now()

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE enrollments SET status = 'REJECTED', rejected = true, pending = false, responded_at = $1
			 WHERE id = $2::uuid`, now, enrollmentID,
		)
		if err != nil {
			log.Error("reject enrollment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to reject enrollment"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "enrollment not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "enrollment rejected"})
	}
}
