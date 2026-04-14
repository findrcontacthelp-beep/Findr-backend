package handlers

import (
	"errors"
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

// accessTokenTTL defines how long a JWT access token remains valid.
const accessTokenTTL = 24 * time.Hour

// Login authenticates a user with email and password, and returns a JWT access token.
//
// Flow:
//  1. Parse and validate the login request body (email + password).
//  2. Authenticate the user against the database via the user repository.
//  3. Issue a signed JWT access token with the user's UUID as the subject.
//  4. Return the token along with user details.
//
// Error cases:
//   - Invalid credentials (wrong email/password) -> 401
//   - Duplicate email conflict (data integrity issue) -> 409
//   - Internal failure (DB error, token signing) -> 500
func Login(pool *pgxpool.Pool, log *zap.Logger, jwtSecret string) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(pool)

	return func(c *gin.Context) {
		// Step 1: Parse and validate the login request
		var req authmodel.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid request body",
				Code:  "INVALID_REQUEST",
			})
			return
		}

		// Step 2: Authenticate user against the database.
		u, err := userRepo.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error: "invalid email or password",
					Code:  "INVALID_CREDENTIALS",
				})
				return
			case errors.Is(err, repository.ErrEmailAlreadyExists):
				// This indicates a data integrity issue — multiple rows share the same email.
				c.JSON(http.StatusConflict, model.ErrorResponse{
					Error: "multiple users found for this email",
					Code:  "EMAIL_CONFLICT",
				})
				return
			default:
				log.Error("login failed", zap.String("email", req.Email), zap.Error(err))
				c.JSON(http.StatusInternalServerError, model.ErrorResponse{
					Error: "failed to login",
					Code:  "LOGIN_FAILED",
				})
				return
			}
		}

		// Step 3: Issue a signed JWT access token with the user's UUID as the subject claim.
		token, err := issueAccessToken(u.UserUUID, jwtSecret)
		if err != nil {
			log.Error("issue access token failed", zap.String("user_uuid", u.UserUUID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "failed to login",
				Code:  "TOKEN_ISSUE_FAILED",
			})
			return
		}

		// Step 4: Return the access token and user profile
		c.JSON(http.StatusOK, authmodel.LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(accessTokenTTL.Seconds()),
			User:        u,
		})
	}
}

// issueAccessToken creates and signs a JWT token using HS256.
//
// Claims:
//   - sub: the user's UUID (used to identify the user in authenticated requests)
//   - iat: issued-at timestamp
//   - exp: expiration timestamp (iat + accessTokenTTL)
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
