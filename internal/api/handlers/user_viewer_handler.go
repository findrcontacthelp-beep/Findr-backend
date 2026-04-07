package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func RecordProfileView(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		viewerUID := middleware.GetUserUID(c)

		_, err := pool.Exec(c.Request.Context(),
			`INSERT INTO user_viewers (user_id, viewer_uid) VALUES ($1::uuid, $2)`,
			userID, viewerUID,
		)
		if err != nil {
			log.Error("record profile view failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to record view"})
			return
		}

		c.JSON(http.StatusCreated, model.SuccessResponse{Message: "view recorded"})
	}
}

func GetProfileViewers(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_id, viewer_uid, viewed_at FROM user_viewers
			 WHERE user_id = $1::uuid ORDER BY viewed_at DESC LIMIT 50`, userID,
		)
		if err != nil {
			log.Error("get profile viewers failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get viewers"})
			return
		}
		defer rows.Close()

		var viewers []model.UserViewer
		for rows.Next() {
			var v model.UserViewer
			if err := rows.Scan(&v.ID, &v.UserID, &v.ViewerUID, &v.ViewedAt); err != nil {
				continue
			}
			viewers = append(viewers, v)
		}

		c.JSON(http.StatusOK, viewers)
	}
}
