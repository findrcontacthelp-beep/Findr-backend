package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	UserUID string
	ChatID  string
	Send    chan []byte
}

func (c *Client) ReadPump(pool *pgxpool.Pool, producer *kafkago.Writer, log *zap.Logger) {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn("websocket read error", zap.Error(err))
			}
			break
		}

		var wsMsg model.WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Warn("invalid websocket message", zap.Error(err))
			continue
		}

		switch wsMsg.Type {
		case "message":
			c.handleChatMessage(pool, producer, wsMsg.Payload, log)
		case "typing":
			c.Hub.BroadcastToChat(c.ChatID, msg)
		case "read_receipt":
			c.handleReadReceipt(pool, wsMsg.Payload, log)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleChatMessage(pool *pgxpool.Pool, producer *kafkago.Writer, payload json.RawMessage, log *zap.Logger) {
	var chatMsg model.WSChatMessage
	if err := json.Unmarshal(payload, &chatMsg); err != nil {
		log.Warn("invalid chat message payload", zap.Error(err))
		return
	}

	if producer != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"chat_id":      c.ChatID,
			"sender_uid":   c.UserUID,
			"message":      chatMsg.Message,
			"receiver_uid": chatMsg.ReceiverUID,
			"reply_to":     chatMsg.ReplyTo,
			"media":        chatMsg.Media,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := producer.WriteMessages(ctx, kafkago.Message{
			Key:   []byte(c.ChatID),
			Value: data,
		})
		if err != nil {
			log.Error("kafka write failed, falling back to direct write", zap.Error(err))
			c.writeMessageDirect(pool, chatMsg, log)
		}
		return
	}

	c.writeMessageDirect(pool, chatMsg, log)
}

func (c *Client) writeMessageDirect(pool *pgxpool.Pool, chatMsg model.WSChatMessage, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var replyToJSON, mediaJSON []byte
	if chatMsg.ReplyTo != nil {
		replyToJSON, _ = json.Marshal(chatMsg.ReplyTo)
	}
	if chatMsg.Media != nil {
		mediaJSON, _ = json.Marshal(chatMsg.Media)
	}

	var msgID string
	err := pool.QueryRow(ctx,
		`INSERT INTO chat_messages (chat_id, sender_uid, receiver_uid, message, reply_to, media, status)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, 'sent')
		 RETURNING id`,
		c.ChatID, c.UserUID, chatMsg.ReceiverUID, chatMsg.Message, replyToJSON, mediaJSON,
	).Scan(&msgID)
	if err != nil {
		log.Error("failed to insert message", zap.Error(err))
		return
	}

	_, err = pool.Exec(ctx,
		`UPDATE chats SET last_message = $1, last_message_at = now() WHERE id = $2::uuid`,
		chatMsg.Message, c.ChatID,
	)
	if err != nil {
		log.Error("failed to update chat last_message", zap.Error(err))
	}

	broadcast, _ := json.Marshal(map[string]interface{}{
		"type":         "message",
		"id":           msgID,
		"chat_id":      c.ChatID,
		"sender_uid":   c.UserUID,
		"message":      chatMsg.Message,
		"receiver_uid": chatMsg.ReceiverUID,
		"reply_to":     chatMsg.ReplyTo,
		"media":        chatMsg.Media,
		"status":       "sent",
	})
	c.Hub.BroadcastToChat(c.ChatID, broadcast)
}

func (c *Client) handleReadReceipt(pool *pgxpool.Pool, payload json.RawMessage, log *zap.Logger) {
	var receipt struct {
		MessageID string `json:"message_id"`
	}
	if err := json.Unmarshal(payload, &receipt); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx,
		`UPDATE chat_messages SET status = 'read' WHERE id = $1::uuid AND chat_id = $2::uuid`,
		receipt.MessageID, c.ChatID,
	)
	if err != nil {
		log.Error("failed to update read receipt", zap.Error(err))
		return
	}

	broadcast, _ := json.Marshal(map[string]interface{}{
		"type":       "read_receipt",
		"message_id": receipt.MessageID,
		"reader_uid": c.UserUID,
	})
	c.Hub.BroadcastToChat(c.ChatID, broadcast)
}
