package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func VerifyUser(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		verifierUID := middleware.GetUserUID(c)

		_, err := pool.Exec(c.Request.Context(),
			`INSERT INTO user_verifications (user_id, verifier_uid) VALUES ($1::uuid, $2)
			 ON CONFLICT (user_id, verifier_uid) DO NOTHING`,
			userID, verifierUID,
		)
		if err != nil {
			log.Error("verify user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to verify user"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "verification recorded"})
	}
}

func GetUserVerifications(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, user_id, verifier_uid, verified_at FROM user_verifications
			 WHERE user_id = $1::uuid ORDER BY verified_at DESC`, userID,
		)
		if err != nil {
			log.Error("get verifications failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get verifications"})
			return
		}
		defer rows.Close()

		var verifications []model.UserVerification
		for rows.Next() {
			var v model.UserVerification
			if err := rows.Scan(&v.ID, &v.UserID, &v.VerifierUID, &v.VerifiedAt); err != nil {
				continue
			}
			verifications = append(verifications, v)
		}

		c.JSON(http.StatusOK, verifications)
	}
}

func RemoveVerification(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		verifierUID := middleware.GetUserUID(c)

		_, err := pool.Exec(c.Request.Context(),
			`DELETE FROM user_verifications WHERE user_id = $1::uuid AND verifier_uid = $2`,
			userID, verifierUID,
		)
		if err != nil {
			log.Error("remove verification failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to remove verification"})
			return
		}

		c.JSON(http.StatusOK, model.SuccessResponse{Message: "verification removed"})
	}
}
