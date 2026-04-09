package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internals/model"
)

func GetCampusBuzz(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT
				id::text,
				text,
				created_at
			FROM public.campus_buzz
			ORDER BY created_at DESC
			LIMIT 10`)
		if err != nil {
			log.Error("get campus buzz failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campus buzz"})
			return
		}
		defer rows.Close()

		buzzItems := make([]model.CampusBuzz, 0)
		for rows.Next() {
			var item model.CampusBuzz
			if err := rows.Scan(
				&item.ID,
				&item.Text,
				&item.CreatedAt,
			); err != nil {
				log.Error("scan campus buzz item failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan campus buzz"})
				return
			}
			buzzItems = append(buzzItems, item)
		}

		if err := rows.Err(); err != nil {
			log.Error("iterate campus buzz failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campus buzz"})
			return
		}

		c.JSON(http.StatusOK, buzzItems)
	}
}
