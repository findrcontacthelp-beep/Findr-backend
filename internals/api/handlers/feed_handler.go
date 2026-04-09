package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internals/model"
)

func GetFeed(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT
				id::text,
				type,
				coalesce(author_name, ''),
				coalesce(author_uid, ''),
				coalesce(title, ''),
				coalesce(description, ''),
				coalesce(tags, '{}'),
				coalesce(image_urls, '{}'),
				coalesce(file_urls, '{}'),
				coalesce(video_url, ''),
				coalesce(links, '{}'),
				coalesce(roles_needed, '{}'),
				coalesce(project_roles, '[]'::jsonb),
				coalesce(enrolled_persons, '{}'::jsonb),
				coalesce(requested_persons, '{}'::jsonb),
				coalesce(likes_count, 0),
				coalesce(comments_count, 0),
				coalesce(views_count, 0),
				created_at,
				updated_at
			FROM public.projects
			WHERE lower(type) <> 'live'
			ORDER BY created_at DESC`)
		if err != nil {
			log.Error("get feed failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
			return
		}
		defer rows.Close()

		feed := make([]model.FeedItem, 0)
		for rows.Next() {
			var item model.FeedItem
			var projectRolesRaw []byte
			var enrolledPersonsRaw []byte
			var requestedPersonsRaw []byte
			if err := rows.Scan(
				&item.ID,
				&item.Type,
				&item.AuthorName,
				&item.AuthorUID,
				&item.Title,
				&item.Description,
				&item.Tags,
				&item.ImageURLs,
				&item.FileURLs,
				&item.VideoURL,
				&item.Links,
				&item.RolesNeeded,
				&projectRolesRaw,
				&enrolledPersonsRaw,
				&requestedPersonsRaw,
				&item.LikesCount,
				&item.CommentsCount,
				&item.ViewsCount,
				&item.CreatedAt,
				&item.UpdatedAt,
			); err != nil {
				log.Error("scan feed item failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
				return
			}

			item.Type = normalizeFeedType(item.Type)
			if item.Type == "project" {
				if err := json.Unmarshal(projectRolesRaw, &item.ProjectRoles); err != nil {
					log.Error("unmarshal project roles failed", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
					return
				}
				if err := json.Unmarshal(enrolledPersonsRaw, &item.EnrolledPersons); err != nil {
					log.Error("unmarshal enrolled persons failed", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
					return
				}
				if err := json.Unmarshal(requestedPersonsRaw, &item.RequestedPersons); err != nil {
					log.Error("unmarshal requested persons failed", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
					return
				}
			}
			feed = append(feed, item)
		}

		if err := rows.Err(); err != nil {
			log.Error("iterate feed failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get feed"})
			return
		}

		c.JSON(http.StatusOK, feed)
	}
}

func normalizeFeedType(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "post", "project", "event", "bug", "announcement", "poll":
		return normalized
	default:
		return "post"
	}
}
