package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internals/model"
)

func GetLiveEvents(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT
				id::text,
				coalesce(title, ''),
				coalesce(description, ''),
				coalesce(views_count, 0),
				created_at,
				coalesce(image_urls[1], '')
			FROM public.posts
			WHERE lower(type) = 'live'
			ORDER BY created_at DESC`)
		if err != nil {
			log.Error("get live events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live events"})
			return
		}
		defer rows.Close()

		events := make([]model.LiveEventSummary, 0)
		for rows.Next() {
			var item model.LiveEventSummary
			if err := rows.Scan(
				&item.ID,
				&item.Title,
				&item.Subtitle,
				&item.ParticipantCount,
				&item.CreatedAt,
				&item.ImageUrl,
			); err != nil {
				log.Error("scan live event item failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan live event"})
				return
			}
			events = append(events, item)
		}

		if err := rows.Err(); err != nil {
			log.Error("iterate live events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live events"})
			return
		}

		c.JSON(http.StatusOK, events)
	}
}

// GetLiveEventByID fetches full details WHERE id = $1 AND lower(type) = 'live', reusing FeedItem shape for full data.
func GetLiveEventByID(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		row := pool.QueryRow(c.Request.Context(),
			`SELECT
				id::text,
				type,
				coalesce(author_name, ''),
				coalesce(author_uuid, ''),
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
			FROM public.posts
			WHERE id = $1 AND lower(type) = 'live'`, id)

		var item model.FeedItem
		var projectRolesRaw []byte
		var enrolledPersonsRaw []byte
		var requestedPersonsRaw []byte
		if err := row.Scan(
			&item.ID,
			&item.Type,
			&item.AuthorName,
			&item.AuthorUUID,
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
			log.Error("scan live event detail failed", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "live event not found"})
			return
		}
		item.Type = "live"
		c.JSON(http.StatusOK, item)
	}
}
