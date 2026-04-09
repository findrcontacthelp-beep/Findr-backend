package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"firebase.google.com/go/v4/messaging"
	"github.com/findr-app/findr-backend/internal/api/middleware"
	fb "github.com/findr-app/findr-backend/internal/firebase"
	"github.com/findr-app/findr-backend/internal/model"
	"github.com/findr-app/findr-backend/internal/ws"
)

func GetMessages(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("id")
		userUID := middleware.GetUserUID(c)

		// Verify participant
		var exists bool
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1::uuid AND user_uid = $2)`,
			chatID, userUID,
		).Scan(&exists)
		if !exists {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "not a participant"})
			return
		}

		var params model.CursorParams
		_ = c.ShouldBindQuery(&params)
		params.SetDefaults()

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, message_id, chat_id, sender_uid, sender_id, receiver_uid,
			        message, status, reply_to, media, created_at
			 FROM chat_messages
			 WHERE chat_id = $1::uuid AND created_at < $2
			 ORDER BY created_at DESC LIMIT $3`,
			chatID, params.Before, params.Limit,
		)
		if err != nil {
			log.Error("get messages failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get messages"})
			return
		}
		defer rows.Close()

		var messages []model.ChatMessage
		for rows.Next() {
			var m model.ChatMessage
			if err := rows.Scan(&m.ID, &m.MessageID, &m.ChatID, &m.SenderUID, &m.SenderID,
				&m.ReceiverUID, &m.Message, &m.Status, &m.ReplyTo, &m.Media, &m.CreatedAt); err != nil {
				log.Error("scan message failed", zap.Error(err))
				continue
			}
			messages = append(messages, m)
		}

		c.JSON(http.StatusOK, messages)
	}
}

func SendMessage(pool *pgxpool.Pool, log *zap.Logger, producer *kafkago.Writer, hub *ws.Hub, fcm *messaging.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("id")
		userUID := middleware.GetUserUID(c)

		// Verify participant
		var exists bool
		_ = pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1::uuid AND user_uid = $2)`,
			chatID, userUID,
		).Scan(&exists)
		if !exists {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "not a participant"})
			return
		}

		var req model.SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		// If Kafka is available, publish to topic
		if producer != nil {
			data, _ := json.Marshal(map[string]interface{}{
				"chat_id":      chatID,
				"sender_uid":   userUID,
				"message":      req.Message,
				"receiver_uid": req.ReceiverUID,
				"reply_to":     req.ReplyTo,
				"media":        req.Media,
			})
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			err := producer.WriteMessages(ctx, kafkago.Message{
				Key:   []byte(chatID),
				Value: data,
			})
			if err != nil {
				log.Error("kafka publish failed, falling back to direct write", zap.Error(err))
			} else {
				c.JSON(http.StatusAccepted, model.SuccessResponse{Message: "message queued"})
				return
			}
		}

		// Direct write fallback
		var replyToJSON, mediaJSON []byte
		if req.ReplyTo != nil {
			replyToJSON, _ = json.Marshal(req.ReplyTo)
		}
		if req.Media != nil {
			mediaJSON, _ = json.Marshal(req.Media)
		}

		var msgID string
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO chat_messages (chat_id, sender_uid, receiver_uid, message, reply_to, media, status)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, 'sent')
			 RETURNING id`,
			chatID, userUID, req.ReceiverUID, req.Message, replyToJSON, mediaJSON,
		).Scan(&msgID)
		if err != nil {
			log.Error("send message failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to send message"})
			return
		}

		_, _ = pool.Exec(c.Request.Context(),
			`UPDATE chats SET last_message = $1, last_message_at = now() WHERE id = $2::uuid`,
			req.Message, chatID,
		)

		// Broadcast via WebSocket when a hub is configured.
		if hub != nil {
			broadcast, _ := json.Marshal(map[string]interface{}{
				"type":         "message",
				"id":           msgID,
				"chat_id":      chatID,
				"sender_uid":   userUID,
				"message":      req.Message,
				"receiver_uid": req.ReceiverUID,
				"reply_to":     req.ReplyTo,
				"media":        req.Media,
				"status":       "sent",
			})
			hub.BroadcastToChat(chatID, broadcast)
		}

		// Push notification for offline receiver when FCM and WebSocket presence are configured.
		if fcm != nil && hub != nil && req.ReceiverUID != nil && !hub.IsUserOnline(chatID, *req.ReceiverUID) {
			var fcmToken string
			_ = pool.QueryRow(c.Request.Context(),
				`SELECT fcm_token FROM users WHERE firebase_uid = $1 AND fcm_token IS NOT NULL`,
				*req.ReceiverUID,
			).Scan(&fcmToken)
			if fcmToken != "" {
				fb.SendPush(c.Request.Context(), fcm, fcmToken, "New Message", req.Message,
					map[string]string{"chat_id": chatID, "sender_uid": userUID}, log)
			}
		}

		c.JSON(http.StatusCreated, gin.H{"id": msgID, "message": "sent"})
	}
}
