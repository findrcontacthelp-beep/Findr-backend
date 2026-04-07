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

func RateUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		raterUID := middleware.GetUserUID(c)

		var req model.RateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		_, err := pool.Exec(c.Request.Context(),
			`INSERT INTO user_ratings (user_id, rater_uid, rating) VALUES ($1::uuid, $2, $3)
			 ON CONFLICT (user_id, rater_uid) DO UPDATE SET rating = $3`,
			userID, raterUID, req.Rating,
		)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "user not found"})
				return
			}
			log.Error("rate user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to rate user"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "rating saved"})
	}
}

func GetUserRatings(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_id, rater_uid, rating FROM user_ratings WHERE user_id = $1::uuid`, userID,
		)
		if err != nil {
			log.Error("get user ratings failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get ratings"})
			return
		}
		defer rows.Close()

		var ratings []model.UserRating
		for rows.Next() {
			var r model.UserRating
			if err := rows.Scan(&r.ID, &r.UserID, &r.RaterUID, &r.Rating); err != nil {
				continue
			}
			ratings = append(ratings, r)
		}

		c.JSON(http.StatusOK, ratings)
	}
}

func GetAverageRating(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var resp model.AverageRatingResponse
		err := pool.QueryRow(c.Request.Context(),
			`SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM user_ratings WHERE user_id = $1::uuid`, userID,
		).Scan(&resp.Average, &resp.Count)
		if err != nil {
			log.Error("get average rating failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get average"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
