package chat

import (
	"context"
	"encoding/json"

	"firebase.google.com/go/v4/messaging"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	fb "github.com/findr-app/findr-backend/internal/firebase"
	"github.com/findr-app/findr-backend/internal/ws"
)

type kafkaMessage struct {
	ChatID      string      `json:"chat_id"`
	SenderUID   string      `json:"sender_uid"`
	Message     string      `json:"message"`
	ReceiverUID *string     `json:"receiver_uid"`
	ReplyTo     interface{} `json:"reply_to"`
	Media       interface{} `json:"media"`
}

func StartConsumer(ctx context.Context, reader *kafkago.Reader, pool *pgxpool.Pool, hub *ws.Hub, fcm *messaging.Client, log *zap.Logger) {
	if reader == nil {
		log.Info("kafka reader is nil, chat consumer not started")
		return
	}

	log.Info("starting chat kafka consumer")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info("chat consumer context cancelled, shutting down")
				return
			}
			log.Error("kafka read error", zap.Error(err))
			continue
		}

		var msg kafkaMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Error("failed to unmarshal kafka message", zap.Error(err))
			continue
		}

		replyToJSON, _ := json.Marshal(msg.ReplyTo)
		mediaJSON, _ := json.Marshal(msg.Media)

		var msgID string
		err = pool.QueryRow(ctx,
			`INSERT INTO chat_messages (chat_id, sender_uid, receiver_uid, message, reply_to, media, status)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, 'sent')
			 RETURNING id`,
			msg.ChatID, msg.SenderUID, msg.ReceiverUID, msg.Message, replyToJSON, mediaJSON,
		).Scan(&msgID)
		if err != nil {
			log.Error("failed to insert chat message from kafka", zap.Error(err))
			continue
		}

		_, _ = pool.Exec(ctx,
			`UPDATE chats SET last_message = $1, last_message_at = now() WHERE id = $2::uuid`,
			msg.Message, msg.ChatID,
		)

		broadcast, _ := json.Marshal(map[string]interface{}{
			"type":         "message",
			"id":           msgID,
			"chat_id":      msg.ChatID,
			"sender_uid":   msg.SenderUID,
			"message":      msg.Message,
			"receiver_uid": msg.ReceiverUID,
			"reply_to":     msg.ReplyTo,
			"media":        msg.Media,
			"status":       "sent",
		})
		hub.BroadcastToChat(msg.ChatID, broadcast)

		// Send push notification to offline participants
		if fcm != nil && msg.ReceiverUID != nil && !hub.IsUserOnline(msg.ChatID, *msg.ReceiverUID) {
			var fcmToken string
			_ = pool.QueryRow(ctx,
				`SELECT fcm_token FROM users WHERE firebase_uid = $1 AND fcm_token IS NOT NULL`,
				*msg.ReceiverUID,
			).Scan(&fcmToken)

			if fcmToken != "" {
				fb.SendPush(ctx, fcm, fcmToken, "New Message", msg.Message, map[string]string{
					"chat_id":    msg.ChatID,
					"sender_uid": msg.SenderUID,
				}, log)
			}
		}
	}
}
