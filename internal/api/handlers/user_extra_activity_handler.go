package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func CreateExtraActivity(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req model.CreateExtraActivityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var activity model.UserExtraActivity
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO user_extra_activities (user_id, name, domain, link, description, media)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, user_id, name, domain, link, description, media`,
			userID, req.Name, req.Domain, req.Link, req.Description, req.Media,
		).Scan(&activity.ID, &activity.UserID, &activity.Name, &activity.Domain,
			&activity.Link, &activity.Description, &activity.Media)
		if err != nil {
			log.Error("create extra activity failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create activity"})
			return
		}

		c.JSON(http.StatusCreated, activity)
	}
}

func GetExtraActivities(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, firebase_id, user_id, name, domain, link, description, media
			 FROM user_extra_activities WHERE user_id = $1::uuid`, userID,
		)
		if err != nil {
			log.Error("get extra activities failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get activities"})
			return
		}
		defer rows.Close()

		var activities []model.UserExtraActivity
		for rows.Next() {
			var a model.UserExtraActivity
			if err := rows.Scan(&a.ID, &a.FirebaseID, &a.UserID, &a.Name, &a.Domain,
				&a.Link, &a.Description, &a.Media); err != nil {
				continue
			}
			activities = append(activities, a)
		}

		c.JSON(http.StatusOK, activities)
	}
}

func UpdateExtraActivity(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		activityID := c.Param("activityId")

		var req model.CreateExtraActivityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE user_extra_activities
			 SET name = $1, domain = $2, link = $3, description = $4, media = $5
			 WHERE id = $6::uuid AND user_id = $7`,
			req.Name, req.Domain, req.Link, req.Description, req.Media, activityID, userID,
		)
		if err != nil {
			log.Error("update extra activity failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update activity"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "activity not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "activity updated"})
	}
}

func DeleteExtraActivity(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		activityID := c.Param("activityId")

		tag, err := pool.Exec(c.Request.Context(),
			`DELETE FROM user_extra_activities WHERE id = $1::uuid AND user_id = $2`,
			activityID, userID,
		)
		if err != nil {
			log.Error("delete extra activity failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete activity"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "activity not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "activity deleted"})
	}
}

// ensure pgx is used
var _ = pgx.ErrNoRows
