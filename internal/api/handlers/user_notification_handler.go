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

func GetNotifications(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_id, last_notified_at
			 FROM user_notifications WHERE user_id = $1 ORDER BY last_notified_at DESC`, userID,
		)
		if err != nil {
			log.Error("get notifications failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get notifications"})
			return
		}
		defer rows.Close()

		var notifications []model.UserNotification
		for rows.Next() {
			var n model.UserNotification
			if err := rows.Scan(&n.ID, &n.UserID, &n.LastNotifiedAt); err != nil {
				continue
			}
			notifications = append(notifications, n)
		}

		c.JSON(http.StatusOK, notifications)
	}
}

func UpdateNotification(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		notifID := c.Param("notifId")

		now := time.Now()
		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE user_notifications SET last_notified_at = $1
			 WHERE id = $2::uuid AND user_id = $3`,
			now, notifID, userID,
		)
		if err != nil {
			log.Error("update notification failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update notification"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "notification not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "notification updated"})
	}
}

func DeleteNotification(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		notifID := c.Param("notifId")

		tag, err := pool.Exec(c.Request.Context(),
			`DELETE FROM user_notifications WHERE id = $1::uuid AND user_id = $2`,
			notifID, userID,
		)
		if err != nil {
			log.Error("delete notification failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete notification"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "notification not found"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "notification deleted"})
	}
}
