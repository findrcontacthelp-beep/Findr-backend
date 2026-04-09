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

func CreateComment(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		userUUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var req model.CreateCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var senderName, senderImage string
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT COALESCE(name, ''), COALESCE(profile_image_url, '') FROM users WHERE id = $1`, userID,
		).Scan(&senderName, &senderImage)

		isTopLevel := req.ParentCommentID == nil

		var comment model.ProjectComment
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO project_comments (project_id, sender_uuid, sender_id, sender_name, sender_image_url,
			  text, nesting_level, parent_comment_id, root_comment_id, is_top_level)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 RETURNING id, project_id, sender_uuid, sender_name, text, nesting_level, is_top_level, created_at`,
			projectID, userUUID, userID, senderName, senderImage,
			req.Text, req.NestingLevel, req.ParentCommentID, req.RootCommentID, isTopLevel,
		).Scan(&comment.ID, &comment.ProjectID, &comment.SenderUUID, &comment.SenderName,
			&comment.Text, &comment.NestingLevel, &comment.IsTopLevel, &comment.CreatedAt)
		if err != nil {
			log.Error("create comment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create comment"})
			return
		}

		// Increment comments_count in posts table
		_, _ = pool.Exec(c.Request.Context(),
			`UPDATE posts SET comments_count = comments_count + 1 WHERE id = $1::uuid`, projectID)

		// Increment parent reply_count if reply
		if req.ParentCommentID != nil {
			_, _ = pool.Exec(c.Request.Context(),
				`UPDATE project_comments SET reply_count = reply_count + 1 WHERE id = $1::uuid`, *req.ParentCommentID)
		}

		c.JSON(http.StatusCreated, comment)
	}
}

func GetComments(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var params model.PaginationParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, project_id, sender_uuid, sender_id, sender_name, sender_image_url,
			        text, nesting_level, parent_comment_id, root_comment_id, is_top_level,
			        reply_count, likes_count, created_at
			 FROM project_comments
			 WHERE project_id = $1::uuid AND is_top_level = true
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			projectID, params.Limit, params.Offset(),
		)
		if err != nil {
			log.Error("get comments failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get comments"})
			return
		}
		defer rows.Close()

		comments := scanComments(rows, log)
		c.JSON(http.StatusOK, comments)
	}
}

func GetReplies(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.Param("commentId")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, project_id, sender_uuid, sender_id, sender_name, sender_image_url,
			        text, nesting_level, parent_comment_id, root_comment_id, is_top_level,
			        reply_count, likes_count, created_at
			 FROM project_comments
			 WHERE parent_comment_id = $1 OR root_comment_id = $1
			 ORDER BY created_at ASC`,
			commentID,
		)
		if err != nil {
			log.Error("get replies failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get replies"})
			return
		}
		defer rows.Close()

		replies := scanComments(rows, log)
		c.JSON(http.StatusOK, replies)
	}
}

func UpdateComment(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.Param("commentId")
		userUUID := middleware.GetUserUID(c)

		var req struct {
			Text string `json:"text" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE project_comments SET text = $1 WHERE id = $2::uuid AND sender_uuid = $3`,
			req.Text, commentID, userUUID,
		)
		if err != nil {
			log.Error("update comment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to update comment"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "comment not found or not authorized"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "comment updated"})
	}
}

func DeleteComment(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.Param("commentId")
		userUUID := middleware.GetUserUID(c)

		// Get project_id before deleting to update count
		var projectID string
		err := pool.QueryRow(c.Request.Context(),
			`SELECT project_id FROM project_comments WHERE id = $1::uuid AND sender_uuid = $2`,
			commentID, userUUID,
		).Scan(&projectID)
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "comment not found or not authorized"})
			return
		}

		_, err = pool.Exec(c.Request.Context(),
			`DELETE FROM project_comments WHERE id = $1::uuid AND sender_uuid = $2`,
			commentID, userUUID,
		)
		if err != nil {
			log.Error("delete comment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete comment"})
			return
		}

		_, _ = pool.Exec(c.Request.Context(),
			`UPDATE posts SET comments_count = GREATEST(comments_count - 1, 0) WHERE id = $1::uuid`, projectID)

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "comment deleted"})
	}
}

func scanComments(rows pgx.Rows, log *zap.Logger) []model.ProjectComment {
	var comments []model.ProjectComment
	for rows.Next() {
		var cm model.ProjectComment
		if err := rows.Scan(
			&cm.ID, &cm.ProjectID, &cm.SenderUUID, &cm.SenderID, &cm.SenderName, &cm.SenderImageURL,
			&cm.Text, &cm.NestingLevel, &cm.ParentCommentID, &cm.RootCommentID, &cm.IsTopLevel,
			&cm.ReplyCount, &cm.LikesCount, &cm.CreatedAt,
		); err != nil {
			log.Error("scan comment failed", zap.Error(err))
			continue
		}
		comments = append(comments, cm)
	}
	return comments
}
