package chat

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/ws"
)

type kafkaMessage struct {
	ChatID       string      `json:"chat_id"`
	SenderUUID   string      `json:"sender_uuid"`
	Message      string      `json:"message"`
	ReceiverUUID *string     `json:"receiver_uuid"`
	ReplyTo      interface{} `json:"reply_to"`
	Media        interface{} `json:"media"`
}

func StartConsumer(ctx context.Context, reader *kafkago.Reader, pool *pgxpool.Pool, hub *ws.Hub, log *zap.Logger) {
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
			`INSERT INTO chat_messages (chat_id, sender_uuid, receiver_uuid, message, reply_to, media, status)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, 'sent')
			 RETURNING id`,
			msg.ChatID, msg.SenderUUID, msg.ReceiverUUID, msg.Message, replyToJSON, mediaJSON,
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
			"type":          "message",
			"id":            msgID,
			"chat_id":       msg.ChatID,
			"sender_uuid":   msg.SenderUUID,
			"message":       msg.Message,
			"receiver_uuid": msg.ReceiverUUID,
			"reply_to":      msg.ReplyTo,
			"media":         msg.Media,
			"status":        "sent",
		})
		hub.BroadcastToChat(msg.ChatID, broadcast)
	}
}
