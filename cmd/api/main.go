package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internals/config"
	"github.com/findr-app/findr-backend/internals/db"
	"github.com/findr-app/findr-backend/internals/logger"
	router "github.com/findr-app/findr-backend/internals/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLog := logger.New(cfg.Environment)
	defer zapLog.Sync()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, zapLog)
	if err != nil {
		zapLog.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Router
	r := gin.New()
	r.Use(gin.Recovery())

	router.Setup(r, &router.Deps{
		Pool:      pool,
		Log:       zapLog,
		JWTSecret: cfg.SupabaseJWTSecret,
		AdminUIDs: cfg.AdminUIDs,
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		zapLog.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLog.Info("shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLog.Error("server forced to shutdown", zap.Error(err))
	}

	zapLog.Info("server exited")
}
