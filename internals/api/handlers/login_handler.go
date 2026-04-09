package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
	authmodel "github.com/findr-app/findr-backend/internals/model"
	"github.com/findr-app/findr-backend/internals/repository"
)

const accessTokenTTL = 24 * time.Hour

func Login(pool *pgxpool.Pool, log *zap.Logger, jwtSecret string) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(pool)

	return func(c *gin.Context) {
		var req authmodel.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		u, err := userRepo.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			switch err {
			case repository.ErrInvalidCredentials:
				c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid email or password"})
				return
			case repository.ErrEmailAlreadyExists:
				c.JSON(http.StatusConflict, model.ErrorResponse{Error: "multiple users found for this email"})
				return
			default:
				log.Error("login failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to login"})
				return
			}
		}

		token, err := issueAccessToken(u.UserUUID, jwtSecret)
		if err != nil {
			log.Error("issue access token failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to login"})
			return
		}

		c.JSON(http.StatusOK, authmodel.LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(accessTokenTTL.Seconds()),
			User:        u,
		})
	}
}

func issueAccessToken(subject, jwtSecret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
