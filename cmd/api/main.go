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

	"github.com/findr-app/findr-backend/internal/api/router"
	"github.com/findr-app/findr-backend/internal/chat"
	"github.com/findr-app/findr-backend/internal/config"
	"github.com/findr-app/findr-backend/internal/db"
	"github.com/findr-app/findr-backend/internal/firebase"
	"github.com/findr-app/findr-backend/internal/kafka"
	"github.com/findr-app/findr-backend/internal/logger"
	"github.com/findr-app/findr-backend/internal/ws"
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

	// Firebase (optional)
	fcm := firebase.InitMessaging(ctx, cfg.FirebaseCredentialsPath, zapLog)

	// Kafka (optional)
	kafkaCfg := &kafka.KafkaConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    "chat-messages",
		GroupID:  "findr-chat-consumer",
		UseTLS:   cfg.KafkaUseTLS,
		Username: cfg.KafkaUsername,
		Password: cfg.KafkaPassword,
	}
	producer := kafka.NewProducer(kafkaCfg, zapLog)
	reader := kafka.NewReader(kafkaCfg, zapLog)

	// WebSocket hub
	hub := ws.NewHub(zapLog)

	// Start Kafka chat consumer
	go chat.StartConsumer(ctx, reader, pool, hub, fcm, zapLog)

	// Router
	r := gin.New()
	r.Use(gin.Recovery())

	router.Setup(r, &router.Deps{
		Pool:      pool,
		Log:       zapLog,
		Hub:       hub,
		Producer:  producer,
		FCM:       fcm,
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

	if producer != nil {
		producer.Close()
	}
	if reader != nil {
		reader.Close()
	}

	zapLog.Info("server exited")
}
