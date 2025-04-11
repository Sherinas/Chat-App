package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Tighten in production
}

type ChatWebSocketHandler struct {
	chatUsecase  usecase.ChatUsecase
	redisService usecase.RedisService
}

func NewChatWebSocketHandler(chatUsecase usecase.ChatUsecase, redisService usecase.RedisService) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		chatUsecase:  chatUsecase,
		redisService: redisService,
	}
}
func (h *ChatWebSocketHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("http upgraded to ws")

	// Validate token
	token := r.URL.Query().Get("token")
	if token == "" {
		conn.WriteJSON(map[string]string{"error": "missing token"})
		return
	}
	userID, _, err := h.chatUsecase.ValidateTokenWithRedis(token)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "invalid token: " + err.Error()})
		return
	}
	log.Printf("User %d connected with token", userID)

	// Determine subscription channels
	// groupIDStr := r.URL.Query().Get("group_id")
	groupIDStr := "5"
	log.Printf("%T", groupIDStr)
	log.Println("g----------------------------------------------------------p---", groupIDStr)
	var channels []string
	if groupIDStr != "" {
		groupID, err := strconv.Atoi(groupIDStr)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "invalid group_id"})
			return
		}
		channels = []string{"group:" + strconv.Itoa(groupID), "user:" + strconv.Itoa(userID)}
		log.Printf("Subscribing user %d to channels: %v", userID, channels)
	} else {
		channels = []string{"user:" + strconv.Itoa(userID)}
		log.Println("HELLOOOOOOOO,ELSE")
	}

	// Subscribe to Redis channels
	msgChan, err := h.redisService.SubscribeToMultipleChannels(channels)

	log.Println("............................", msgChan)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "subscription failed: " + err.Error()})
		return
	}

	// Handle client messages and Redis events concurrently
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			// Read client messages
			_, clientMsg, err := conn.ReadMessage()
			log.Println("read---------", string(clientMsg))
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error for user %d: %v", userID, err)
				}
				return
			}

			type clientRequest struct {
				Type       string `json:"type"` // Fixed from "targetType"
				ReceiverID int    `json:"targetId,omitempty"`
				GroupID    int    `json:"group_id,omitempty"`
				Content    string `json:"content"`
				SenderID   int    `json:"sender_id,omitempty"` // Add for validation
			}
			var req clientRequest

			if err := json.Unmarshal(clientMsg, &req); err != nil {
				conn.WriteJSON(map[string]string{"error": "invalid message format"})
				continue
			}

			log.Println("reww==", token, req.GroupID, req.Content)
			log.Println("req--", req)
			if req.SenderID == 0 {
				req.SenderID = userID // Default to authenticated userID
			} else if req.SenderID != userID {
				conn.WriteJSON(map[string]string{"error": "sender_id mismatch"})
				continue
			}

			switch req.Type {
			case "personal_message":
				if err := h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content); err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send personal message: " + err.Error()})
				}
			case "group_message":
				if err := h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content); err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send group message: " + err.Error()})
				}
			case "voice_message":
				var receiverID, groupID *int
				if req.ReceiverID != 0 {
					receiverID = &req.ReceiverID
				}
				if req.GroupID != 0 {
					groupID = &req.GroupID
				}
				if err := h.chatUsecase.SendVoiceMessage(token, receiverID, groupID, req.Content); err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send voice message: " + err.Error()})
				}
			default:
				conn.WriteJSON(map[string]string{"error": "unknown message type"})
			}
		}
	}()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				log.Printf("Redis channel closed for user %d", userID)
				return
			}
			log.Printf("Received from Redis for user %d: %s", userID, msg)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Printf("WebSocket write error for user %d: %v", userID, err)
				return
			}
		case <-done:
			return
		}
	}
}

func RegisterWebSocketRoute(mux *http.ServeMux, chatUsecase usecase.ChatUsecase, redisService usecase.RedisService) {
	handler := NewChatWebSocketHandler(chatUsecase, redisService)
	mux.HandleFunc("/ws/chat", handler.HandleChat)
}
