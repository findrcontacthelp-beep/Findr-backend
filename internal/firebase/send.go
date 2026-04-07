package firebase

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
)

func SendPush(ctx context.Context, client *messaging.Client, token, title, body string, data map[string]string, log *zap.Logger) {
	if client == nil || token == "" {
		return
	}

	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := client.Send(ctx, msg)
	if err != nil {
		log.Warn("failed to send push notification", zap.Error(err), zap.String("token", token[:min(len(token), 10)]+"..."))
	}
}

func SendToTopic(ctx context.Context, client *messaging.Client, topic, title, body string, data map[string]string, log *zap.Logger) {
	if client == nil {
		return
	}

	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := client.Send(ctx, msg)
	if err != nil {
		log.Warn("failed to send topic notification", zap.Error(err), zap.String("topic", topic))
	}
}
