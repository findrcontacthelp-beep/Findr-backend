package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func RecordProjectView(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		viewerUID := middleware.GetUserUID(c)

		_, err := pool.Exec(c.Request.Context(),
			`INSERT INTO project_post_views (project_id, viewer_uid)
			 VALUES ($1::uuid, $2)
			 ON CONFLICT (project_id, viewer_uid) DO NOTHING`,
			projectID, viewerUID,
		)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found"})
				return
			}
			log.Error("record project view failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to record view"})
			return
		}

		// Update views count
		_, _ = pool.Exec(c.Request.Context(),
			`UPDATE projects SET post_views_count = (
				SELECT COUNT(*) FROM project_post_views WHERE project_id = $1::uuid
			) WHERE id = $1::uuid`, projectID)

		c.JSON(http.StatusCreated, model.SuccessResponse{Message: "view recorded"})
	}
}

func GetProjectViewCount(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var count int
		err := pool.QueryRow(c.Request.Context(),
			`SELECT COUNT(*) FROM project_post_views WHERE project_id = $1::uuid`, projectID,
		).Scan(&count)
		if err != nil {
			log.Error("get project view count failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get view count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"views_count": count})
	}
}
