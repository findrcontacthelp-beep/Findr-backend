package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWebSocket(hub *Hub, pool *pgxpool.Pool, producer *kafkago.Writer, jwtSecret string, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		chatID := c.Query("chat_id")

		if tokenStr == "" || chatID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token and chat_id required"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userUUID, _ := claims["sub"].(string)
		if userUUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing subject"})
			return
		}

		var exists bool
		err = pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM chat_participants cp
				WHERE cp.chat_id = $1::uuid AND cp.user_uuid = $2
			)`, chatID, userUUID,
		).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a participant of this chat"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error("websocket upgrade failed", zap.Error(err))
			return
		}

		client := &Client{
			Hub:      hub,
			Conn:     conn,
			UserUUID: userUUID,
			ChatID:   chatID,
			Send:     make(chan []byte, 256),
		}

		hub.Register(client)
		go client.WritePump()
		go client.ReadPump(pool, producer, log)
	}
}
