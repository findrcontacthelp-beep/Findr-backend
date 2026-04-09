package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	authhandlers "github.com/findr-app/findr-backend/internals/api/handlers"
	"github.com/findr-app/findr-backend/internals/api/middleware"
)

type Deps struct {
	Pool      *pgxpool.Pool
	Log       *zap.Logger
	JWTSecret string
}

func Setup(r *gin.Engine, d *Deps) {
	auth := middleware.AuthRequired(d.Pool, d.JWTSecret, d.Log)

	v1 := r.Group("/api/v1")

	// Register
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", authhandlers.Login(d.Pool, d.Log, d.JWTSecret))
		authGroup.POST("/register", auth, authhandlers.Register(d.Pool, d.Log))
	}

	v1.GET("/feed", authhandlers.GetFeed(d.Pool, d.Log))
}
