package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
)

func ListTopics(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, firebase_id, topic, enabled FROM topics ORDER BY topic`)
		if err != nil {
			log.Error("list topics failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list topics"})
			return
		}
		defer rows.Close()

		var topics []model.Topic
		for rows.Next() {
			var t model.Topic
			if err := rows.Scan(&t.ID, &t.FirebaseID, &t.Topic, &t.Enabled); err != nil {
				continue
			}
			topics = append(topics, t)
		}

		c.JSON(http.StatusOK, topics)
	}
}

func CreateTopic(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateTopicRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var topic model.Topic
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO topics (topic, enabled) VALUES ($1, $2) RETURNING id, topic, enabled`,
			req.Topic, req.Enabled,
		).Scan(&topic.ID, &topic.Topic, &topic.Enabled)
		if err != nil {
			log.Error("create topic failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create topic"})
			return
		}

		c.JSON(http.StatusCreated, topic)
	}
}
