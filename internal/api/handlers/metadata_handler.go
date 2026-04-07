package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
)

func GetMetadata(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")

		var m model.Metadata
		err := pool.QueryRow(c.Request.Context(),
			`SELECT key, value, updated_at FROM metadata WHERE key = $1`, key,
		).Scan(&m.Key, &m.Value, &m.UpdatedAt)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "key not found"})
			return
		}
		if err != nil {
			log.Error("get metadata failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get metadata"})
			return
		}

		c.JSON(http.StatusOK, m)
	}
}

func SetMetadata(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")

		var req model.SetMetadataRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`INSERT INTO metadata (key, value) VALUES ($1, $2)
			 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = now()`,
			key, req.Value,
		)
		if err != nil {
			log.Error("set metadata failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to set metadata"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "metadata updated"})
	}
}
