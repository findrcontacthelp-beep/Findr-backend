package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func CreateProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var req model.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var authorName string
		_ = pool.QueryRow(c.Request.Context(), `SELECT name FROM users WHERE id = $1`, userID).Scan(&authorName)

		var p model.Project
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO projects (author_uid, author_id, author_name, type, title, title_lower,
			  description, tags, image_urls, file_urls, video_url, links, roles_needed,
			  project_roles, event_details)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			 RETURNING id, type, title, created_at`,
			userUID, userID, authorName, req.Type, req.Title, strings.ToLower(req.Title),
			req.Description, req.Tags, req.ImageURLs, req.FileURLs, req.VideoURL,
			req.Links, req.RolesNeeded, req.ProjectRoles, req.EventDetails,
		).Scan(&p.ID, &p.Type, &p.Title, &p.CreatedAt)
		if err != nil {
			log.Error("create project failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create project"})
			return
		}

		c.JSON(http.StatusCreated, p)
	}
}

func ListProjects(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params model.PaginationParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		projectType := c.Query("type")
		search := c.Query("search")
		tag := c.Query("tag")

		query := `SELECT id, author_uid, author_id, author_name, type, title, description,
		                 tags, image_urls, likes_count, comments_count, views_count, created_at
		          FROM projects WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM projects WHERE 1=1`
		args := []interface{}{}
		argIdx := 1

		if projectType != "" {
			query += fmt.Sprintf(` AND type = $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND type = $%d`, argIdx)
			args = append(args, projectType)
			argIdx++
		}
		if search != "" {
			query += fmt.Sprintf(` AND title_lower LIKE $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND title_lower LIKE $%d`, argIdx)
			args = append(args, "%"+strings.ToLower(search)+"%")
			argIdx++
		}
		if tag != "" {
			query += fmt.Sprintf(` AND $%d = ANY(tags)`, argIdx)
			countQuery += fmt.Sprintf(` AND $%d = ANY(tags)`, argIdx)
			args = append(args, tag)
			argIdx++
		}

		var totalCount int
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		_ = pool.QueryRow(c.Request.Context(), countQuery, countArgs...).Scan(&totalCount)

		query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, params.Limit, params.Offset())

		rows, err := pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			log.Error("list projects failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list projects"})
			return
		}
		defer rows.Close()

		var projects []model.Project
		for rows.Next() {
			var p model.Project
			if err := rows.Scan(
				&p.ID, &p.AuthorUID, &p.AuthorID, &p.AuthorName, &p.Type, &p.Title, &p.Description,
				&p.Tags, &p.ImageURLs, &p.LikesCount, &p.CommentsCount, &p.ViewsCount, &p.CreatedAt,
			); err != nil {
				log.Error("scan project failed", zap.Error(err))
				continue
			}
			projects = append(projects, p)
		}

		c.JSON(http.StatusOK, model.ListResponse[model.Project]{
			Data:       projects,
			TotalCount: totalCount,
			Page:       params.Page,
			Limit:      params.Limit,
			HasMore:    params.Page*params.Limit < totalCount,
		})
	}
}

func GetProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var p model.Project
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, firebase_id, author_uid, author_id, author_name, type, title, title_lower,
			        description, tags, image_urls, file_urls, video_url, links, roles_needed,
			        project_roles, event_details, enrolled_persons, requested_persons,
			        likes, likes_count, comments_count, views_count, post_views_count,
			        created_at, updated_at
			 FROM projects WHERE id = $1`, id,
		).Scan(
			&p.ID, &p.FirebaseID, &p.AuthorUID, &p.AuthorID, &p.AuthorName, &p.Type, &p.Title, &p.TitleLower,
			&p.Description, &p.Tags, &p.ImageURLs, &p.FileURLs, &p.VideoURL, &p.Links, &p.RolesNeeded,
			&p.ProjectRoles, &p.EventDetails, &p.EnrolledPersons, &p.RequestedPersons,
			&p.Likes, &p.LikesCount, &p.CommentsCount, &p.ViewsCount, &p.PostViewsCount,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found"})
			return
		}
		if err != nil {
			log.Error("get project failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get project"})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func UpdateProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		var req model.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE projects SET title = $1, title_lower = $2, description = $3, tags = $4,
			  image_urls = $5, file_urls = $6, video_url = $7, links = $8, roles_needed = $9,
			  project_roles = $10, event_details = $11
			 WHERE id = $12 AND author_id = $13`,
			req.Title, strings.ToLower(req.Title), req.Description, req.Tags,
			req.ImageURLs, req.FileURLs, req.VideoURL, req.Links, req.RolesNeeded,
			req.ProjectRoles, req.EventDetails, id, userID,
		)
		if err != nil {
			log.Error("update project failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update project"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found or not authorized"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "project updated"})
	}
}

func DeleteProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		tag, err := pool.Exec(c.Request.Context(),
			`DELETE FROM projects WHERE id = $1 AND author_id = $2`, id, userID,
		)
		if err != nil {
			log.Error("delete project failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete project"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found or not authorized"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "project deleted"})
	}
}

func ToggleLike(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userUID := middleware.GetUserUID(c)

		var likes []string
		err := pool.QueryRow(c.Request.Context(),
			`SELECT likes FROM projects WHERE id = $1`, id,
		).Scan(&likes)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found"})
			return
		}
		if err != nil {
			log.Error("get likes failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to toggle like"})
			return
		}

		found := false
		newLikes := []string{}
		for _, uid := range likes {
			if uid == userUID {
				found = true
				continue
			}
			newLikes = append(newLikes, uid)
		}
		if !found {
			newLikes = append(newLikes, userUID)
		}

		_, err = pool.Exec(c.Request.Context(),
			`UPDATE projects SET likes = $1, likes_count = $2 WHERE id = $3`,
			newLikes, len(newLikes), id,
		)
		if err != nil {
			log.Error("update likes failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to toggle like"})
			return
		}

		action := "liked"
		if found {
			action = "unliked"
		}
		c.JSON(http.StatusOK, gin.H{"message": action, "likes_count": len(newLikes)})
	}
}

func GetMyProjects(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, type, title, description, tags, likes_count, comments_count, created_at
			 FROM projects WHERE author_id = $1 ORDER BY created_at DESC`, userID,
		)
		if err != nil {
			log.Error("get my projects failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get projects"})
			return
		}
		defer rows.Close()

		var projects []model.Project
		for rows.Next() {
			var p model.Project
			if err := rows.Scan(&p.ID, &p.Type, &p.Title, &p.Description, &p.Tags,
				&p.LikesCount, &p.CommentsCount, &p.CreatedAt); err != nil {
				continue
			}
			projects = append(projects, p)
		}

		c.JSON(http.StatusOK, projects)
	}
}

func GetProjectStats(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var stats struct {
			LikesCount    int `json:"likes_count"`
			CommentsCount int `json:"comments_count"`
			ViewsCount    int `json:"views_count"`
		}
		err := pool.QueryRow(c.Request.Context(),
			`SELECT likes_count, comments_count, views_count FROM projects WHERE id = $1`, id,
		).Scan(&stats.LikesCount, &stats.CommentsCount, &stats.ViewsCount)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "project not found"})
			return
		}
		if err != nil {
			log.Error("get project stats failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// Ensure json import is used
var _ = json.RawMessage{}
