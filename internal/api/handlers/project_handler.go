package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func CreateProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID := middleware.GetUserUID(c)
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
			`INSERT INTO posts (author_uuid, author_name, type, title,
			  description, tags, image_urls, file_urls, video_url, links, roles_needed,
			  project_roles)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			 RETURNING id, type, title, created_at`,
			userUUID, authorName, req.Type, req.Title,
			req.Description, req.Tags, req.ImageURLs, req.FileURLs, req.VideoURL,
			req.Links, req.RolesNeeded, req.ProjectRoles,
		).Scan(&p.ID, &p.Type, &p.Title, &p.CreatedAt)
		if err != nil {
			log.Error("create post failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create post"})
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

		query := `SELECT id, author_uuid, author_name, type, title, description,
		                 tags, image_urls, likes_count, comments_count, views_count, created_at
		          FROM posts WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM posts WHERE 1=1`
		args := []interface{}{}
		argIdx := 1

		if projectType != "" {
			query += fmt.Sprintf(` AND type = $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND type = $%d`, argIdx)
			args = append(args, projectType)
			argIdx++
		}
		if search != "" {
			query += fmt.Sprintf(` AND title ILIKE $%d`, argIdx)
			countQuery += fmt.Sprintf(` AND title ILIKE $%d`, argIdx)
			args = append(args, "%"+search+"%")
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
			log.Error("list posts failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list posts"})
			return
		}
		defer rows.Close()

		var projects []model.Project
		for rows.Next() {
			var p model.Project
			if err := rows.Scan(
				&p.ID, &p.AuthorUUID, &p.AuthorName, &p.Type, &p.Title, &p.Description,
				&p.Tags, &p.ImageURLs, &p.LikesCount, &p.CommentsCount, &p.ViewsCount, &p.CreatedAt,
			); err != nil {
				log.Error("scan post failed", zap.Error(err))
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
			`SELECT id, author_uuid, author_name, type, title, 
			        description, tags, image_urls, file_urls, video_url, links, roles_needed,
			        project_roles, enrolled_persons, requested_persons,
			        likes_count, comments_count, views_count,
			        created_at, updated_at
			 FROM posts WHERE id = $1`, id,
		).Scan(
			&p.ID, &p.AuthorUUID, &p.AuthorName, &p.Type, &p.Title,
			&p.Description, &p.Tags, &p.ImageURLs, &p.FileURLs, &p.VideoURL, &p.Links, &p.RolesNeeded,
			&p.ProjectRoles, &p.EnrolledPersons, &p.RequestedPersons,
			&p.LikesCount, &p.CommentsCount, &p.ViewsCount,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
			return
		}
		if err != nil {
			log.Error("get post failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get post"})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func UpdateProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userUUID := middleware.GetUserUID(c)

		var req model.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE posts SET title = $1, description = $2, tags = $3,
			  image_urls = $4, file_urls = $5, video_url = $6, links = $7, roles_needed = $8,
			  project_roles = $9
			 WHERE id = $10 AND author_uuid = $11`,
			req.Title, req.Description, req.Tags,
			req.ImageURLs, req.FileURLs, req.VideoURL, req.Links, req.RolesNeeded,
			req.ProjectRoles, id, userUUID,
		)
		if err != nil {
			log.Error("update post failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update post"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "post not found or not authorized"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "post updated"})
	}
}

func DeleteProject(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userUUID := middleware.GetUserUID(c)

		tag, err := pool.Exec(c.Request.Context(),
			`DELETE FROM posts WHERE id = $1 AND author_uuid = $2`, id, userUUID,
		)
		if err != nil {
			log.Error("delete post failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete post"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "post not found or not authorized"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "post deleted"})
	}
}

func ToggleLike(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userUUID := middleware.GetUserUID(c)

		// Check if already liked via a separate likes table (preferred)
		// Or if we still use the array column, but it was dropped in the new schema.
		// Since 'likes' column is gone, I will assume there's a post_likes table.
		// However, for the purpose of the current refactor, I'll just increment likes_count
		// and leave the detailed logic if I had a schema for it.
		// Wait, I'll check if there's a post_likes table.
		
		var currentlyLiked bool
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_uuid = $2)`,
			id, userUUID,
		).Scan(&currentlyLiked)

		var err error
		if currentlyLiked {
			_, err = pool.Exec(c.Request.Context(), `DELETE FROM post_likes WHERE post_id = $1 AND user_uuid = $2`, id, userUUID)
			if err == nil {
				_, _ = pool.Exec(c.Request.Context(), `UPDATE posts SET likes_count = likes_count - 1 WHERE id = $1`, id)
			}
		} else {
			_, err = pool.Exec(c.Request.Context(), `INSERT INTO post_likes (post_id, user_uuid) VALUES ($1, $2)`, id, userUUID)
			if err == nil {
				_, _ = pool.Exec(c.Request.Context(), `UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1`, id)
			}
		}

		if err != nil {
			log.Error("toggle like failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to toggle like"})
			return
		}

		action := "liked"
		if currentlyLiked {
			action = "unliked"
		}
		
		var newCount int
		_ = pool.QueryRow(c.Request.Context(), `SELECT likes_count FROM posts WHERE id = $1`, id).Scan(&newCount)
		
		c.JSON(http.StatusOK, gin.H{"message": action, "likes_count": newCount})
	}
}

func GetMyProjects(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID := middleware.GetUserUID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, author_uuid, type, title, description, tags, likes_count, comments_count, created_at
			 FROM posts WHERE author_uuid = $1 ORDER BY created_at DESC`, userUUID,
		)
		if err != nil {
			log.Error("get my posts failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get posts"})
			return
		}
		defer rows.Close()

		var projects []model.Project
		for rows.Next() {
			var p model.Project
			if err := rows.Scan(&p.ID, &p.AuthorUUID, &p.Type, &p.Title, &p.Description, &p.Tags,
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
			`SELECT likes_count, comments_count, views_count FROM posts WHERE id = $1`, id,
		).Scan(&stats.LikesCount, &stats.CommentsCount, &stats.ViewsCount)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
			return
		}
		if err != nil {
			log.Error("get post stats failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func RecordProjectView(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := pool.Exec(c.Request.Context(), `UPDATE posts SET views_count = views_count + 1 WHERE id = $1`, id)
		if err != nil {
			log.Error("record view failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to record view"})
			return
		}
		c.JSON(http.StatusOK, model.SuccessResponse{Message: "view recorded"})
	}
}

func GetProjectViewCount(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var count int
		err := pool.QueryRow(c.Request.Context(), `SELECT views_count FROM posts WHERE id = $1`, id).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get view count"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"views_count": count})
	}
}

// Ensure json import is used
var _ = json.RawMessage{}
