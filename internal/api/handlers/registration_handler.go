package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func RegisterForEvent(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		userUUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var userName, profilePic string
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT COALESCE(name,''), COALESCE(profile_image_url,'') FROM users WHERE id = $1`, userID,
		).Scan(&userName, &profilePic)

		var reg model.Registration
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO registrations (user_uuid, user_id, user_name, user_profile_pic, project_id, status, registered)
			 VALUES ($1, $2, $3, $4, $5::uuid, 'REGISTERED', true)
			 RETURNING id, status, registered_at`,
			userUUID, userID, userName, profilePic, projectID,
		).Scan(&reg.ID, &reg.Status, &reg.RegisteredAt)
		if err != nil {
			log.Error("register for event failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to register"})
			return
		}

		c.JSON(http.StatusCreated, reg)
	}
}

func GetEventRegistrations(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_uuid, user_id, user_name, user_profile_pic, project_id,
			        status, registered, attended, attendance_confirmed, cancelled, registered_at
			 FROM registrations WHERE project_id = $1::uuid ORDER BY registered_at DESC`, projectID,
		)
		if err != nil {
			log.Error("get event registrations failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get registrations"})
			return
		}
		defer rows.Close()

		var registrations []model.Registration
		for rows.Next() {
			var r model.Registration
			if err := rows.Scan(&r.ID, &r.UserUUID, &r.UserID, &r.UserName, &r.UserProfilePic, &r.ProjectID,
				&r.Status, &r.Registered, &r.Attended, &r.AttendanceConfirmed, &r.Cancelled, &r.RegisteredAt); err != nil {
				continue
			}
			registrations = append(registrations, r)
		}

		c.JSON(http.StatusOK, registrations)
	}
}

func GetMyRegistrations(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_uuid, user_id, user_name, user_profile_pic, project_id,
			        status, registered, attended, attendance_confirmed, cancelled, registered_at
			 FROM registrations WHERE user_id = $1 ORDER BY registered_at DESC`, userID,
		)
		if err != nil {
			log.Error("get my registrations failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get registrations"})
			return
		}
		defer rows.Close()

		var registrations []model.Registration
		for rows.Next() {
			var r model.Registration
			if err := rows.Scan(&r.ID, &r.UserUUID, &r.UserID, &r.UserName, &r.UserProfilePic, &r.ProjectID,
				&r.Status, &r.Registered, &r.Attended, &r.AttendanceConfirmed, &r.Cancelled, &r.RegisteredAt); err != nil {
				continue
			}
			registrations = append(registrations, r)
		}

		c.JSON(http.StatusOK, registrations)
	}
}

func CancelRegistration(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		regID := c.Param("id")
		userID := middleware.GetUserID(c)

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE registrations SET status = 'CANCELLED', cancelled = true, registered = false
			 WHERE id = $1::uuid AND user_id = $2`, regID, userID,
		)
		if err != nil {
			log.Error("cancel registration failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to cancel registration"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "registration not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "registration cancelled"})
	}
}
