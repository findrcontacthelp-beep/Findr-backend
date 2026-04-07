package firebase

import (
	"context"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func InitMessaging(ctx context.Context, credentialsPath string, log *zap.Logger) *messaging.Client {
	if credentialsPath == "" {
		log.Info("firebase credentials not configured, FCM disabled")
		return nil
	}

	app, err := fb.NewApp(ctx, nil, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Error("failed to initialize firebase app", zap.Error(err))
		return nil
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Error("failed to initialize firebase messaging", zap.Error(err))
		return nil
	}

	log.Info("firebase messaging initialized")
	return client
}
