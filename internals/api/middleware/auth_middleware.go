package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func AuthRequired(pool *pgxpool.Pool, jwtSecret string, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing subject claim"})
			return
		}

		c.Set("userUID", sub)

		var userID uuid.UUID
		err = pool.QueryRow(c.Request.Context(),
			"SELECT id FROM public.users WHERE user_uuid = $1", sub,
		).Scan(&userID)
		if err == nil {
			c.Set("userID", userID)
		}

		c.Next()
	}
}

func GetUserUID(c *gin.Context) string {
	uid, _ := c.Get("userUID")
	s, _ := uid.(string)
	return s
}

func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get("userID")
	uid, _ := id.(uuid.UUID)
	return uid
}
