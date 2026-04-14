package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	commonmodel "github.com/findr-app/findr-backend/internal/model"
	"github.com/findr-app/findr-backend/internals/model"
)

// GetLiveEvents returns a list of all live events, ordered by most recent first.
//
//
// GetLiveEvents returns a paginated list of live events, ordered by most recent first.
//
// Query params:
//   - page:  page number (default: 1, min: 3ß)
//   - limit: items per page (default: 20, min: 1, max: 100)
//
// Response: ListResponse[LiveEventSummary] with total_count, page, limit, has_more.
//
// Route: GET /api/v1/live?page=1&limit=20
func GetLiveEvents(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Parse and apply default pagination params from query string
		var params commonmodel.PaginationParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		ctx := c.Request.Context()

		// Step 2: Get total count of live events for pagination metadata
		var totalCount int
		err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM public.posts WHERE lower(type) = 'live'`).Scan(&totalCount)
		if err != nil {
			log.Error("count live events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live events"})
			return
		}

		// Step 3: Query the current page of live events, newest first
		rows, err := pool.Query(ctx,
			`SELECT
				id::text,
				coalesce(title, ''),
				coalesce(description, ''),
				coalesce(views_count, 0),
				created_at,
				coalesce(image_urls[1], '')
			FROM public.posts
			WHERE lower(type) = 'live'
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`,
			params.Limit, params.Offset())
		if err != nil {
			log.Error("get live events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live events"})
			return
		}
		defer rows.Close()

		// Step 4: Scan each row into a LiveEventSummary struct.
		// Initialized as empty slice (not nil) so JSON serializes to [] instead of null.
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

		// Step 5: Check for errors encountered during row iteration (e.g. connection drop)
		if err := rows.Err(); err != nil {
			log.Error("iterate live events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live events"})
			return
		}

		// Step 6: Return paginated response with metadata
		c.JSON(http.StatusOK, commonmodel.ListResponse[model.LiveEventSummary]{
			Data:       events,
			TotalCount: totalCount,
			Page:       params.Page,
			Limit:      params.Limit,
			HasMore:    params.Page*params.Limit < totalCount,
		})
	}
}

// GetLiveEventByID fetches full details of a single live event by its ID.
//
// Error cases:
//   - Event not found or not of type 'live' -> 404
//   - Database error -> 500 (logged as 404 to client to avoid leaking internals)
//
func GetLiveEventByID(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Step 1: Query a single live event by ID from the posts table
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

		// Step 2: Scan the result into a FeedItem.
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
			log.Error("scan live event detail failed", zap.String("id", id), zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "live event not found"})
			return
		}

		// Step 3: Ensure type is set explicitly in case the DB value differs in casing
		item.Type = "live"
		c.JSON(http.StatusOK, item)
	}
}
