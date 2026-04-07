package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/middleware"
	"github.com/findr-app/findr-backend/internal/model"
)

func CreateOrGetChat(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)
		userID := middleware.GetUserID(c)

		var req model.CreateChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		// Include current user in participants
		allUIDs := append(req.ParticipantUIDs, userUID)
		sort.Strings(allUIDs)
		allUIDs = uniqueStrings(allUIDs)

		// Check if chat already exists with these exact participants
		for _, uid := range allUIDs {
			if uid == userUID {
				continue
			}
			var chatID uuid.UUID
			err := pool.QueryRow(c.Request.Context(),
				`SELECT cp1.chat_id FROM chat_participants cp1
				 JOIN chat_participants cp2 ON cp1.chat_id = cp2.chat_id
				 WHERE cp1.user_uid = $1 AND cp2.user_uid = $2
				 LIMIT 1`, userUID, uid,
			).Scan(&chatID)
			if err == nil {
				var chat model.Chat
				_ = pool.QueryRow(c.Request.Context(),
					`SELECT id, last_message, last_message_at, unread_messages, created_at
					 FROM chats WHERE id = $1`, chatID,
				).Scan(&chat.ID, &chat.LastMessage, &chat.LastMessageAt, &chat.UnreadMessages, &chat.CreatedAt)
				c.JSON(http.StatusOK, chat)
				return
			}
		}

		// Create new chat
		var chat model.Chat
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO chats (unread_messages) VALUES ('{}') RETURNING id, created_at`,
		).Scan(&chat.ID, &chat.CreatedAt)
		if err != nil {
			log.Error("create chat failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create chat"})
			return
		}

		// Add participants
		for _, uid := range allUIDs {
			var pUserID *uuid.UUID
			var id uuid.UUID
			if err := pool.QueryRow(c.Request.Context(),
				`SELECT id FROM users WHERE firebase_uid = $1`, uid).Scan(&id); err == nil {
				pUserID = &id
			}

			_, _ = pool.Exec(c.Request.Context(),
				`INSERT INTO chat_participants (chat_id, user_uid, user_id) VALUES ($1, $2, $3)`,
				chat.ID, uid, pUserID,
			)
		}
		_ = userID // used for lookup above

		c.JSON(http.StatusCreated, chat)
	}
}

func GetMyChats(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUID := middleware.GetUserUID(c)

		rows, err := pool.Query(c.Request.Context(),
			`SELECT c.id, c.last_message, c.last_message_at, c.unread_messages, c.created_at
			 FROM chats c
			 JOIN chat_participants cp ON cp.chat_id = c.id
			 WHERE cp.user_uid = $1
			 ORDER BY COALESCE(c.last_message_at, c.created_at) DESC`, userUID,
		)
		if err != nil {
			log.Error("get my chats failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get chats"})
			return
		}
		defer rows.Close()

		var chats []model.Chat
		for rows.Next() {
			var chat model.Chat
			if err := rows.Scan(&chat.ID, &chat.LastMessage, &chat.LastMessageAt, &chat.UnreadMessages, &chat.CreatedAt); err != nil {
				continue
			}
			chats = append(chats, chat)
		}

		c.JSON(http.StatusOK, chats)
	}
}

func GetChatDetail(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("id")
		userUID := middleware.GetUserUID(c)

		// Verify participant
		var exists bool
		err := pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1::uuid AND user_uid = $2)`,
			chatID, userUID,
		).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "not a participant"})
			return
		}

		var chat model.ChatWithParticipants
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, last_message, last_message_at, unread_messages, created_at
			 FROM chats WHERE id = $1::uuid`, chatID,
		).Scan(&chat.ID, &chat.LastMessage, &chat.LastMessageAt, &chat.UnreadMessages, &chat.CreatedAt)
		if err != nil {
			log.Error("get chat detail failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get chat"})
			return
		}

		pRows, err := pool.Query(c.Request.Context(),
			`SELECT chat_id, user_uid, user_id FROM chat_participants WHERE chat_id = $1::uuid`, chatID,
		)
		if err == nil {
			defer pRows.Close()
			for pRows.Next() {
				var p model.ChatParticipant
				if err := pRows.Scan(&p.ChatID, &p.UserUID, &p.UserID); err == nil {
					chat.Participants = append(chat.Participants, p)
				}
			}
		}

		c.JSON(http.StatusOK, chat)
	}
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
