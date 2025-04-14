// package websocket

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"strconv"

// 	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin:     func(r *http.Request) bool { return true }, // Tighten in production
// }

// type ChatWebSocketHandler struct {
// 	chatUsecase  usecase.ChatUsecase
// 	redisService usecase.RedisService
// }

// func NewChatWebSocketHandler(chatUsecase usecase.ChatUsecase, redisService usecase.RedisService) *ChatWebSocketHandler {
// 	return &ChatWebSocketHandler{
// 		chatUsecase:  chatUsecase,
// 		redisService: redisService,
// 	}
// }

// // func (h *ChatWebSocketHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
// // 	// Upgrade to WebSocket
// // 	conn, err := upgrader.Upgrade(w, r, nil)
// // 	if err != nil {
// // 		log.Printf("WebSocket upgrade failed: %v", err)
// // 		return
// // 	}
// // 	defer conn.Close()

// // 	log.Printf("http upgraded to ws")

// // 	// Validate token
// // 	token := r.URL.Query().Get("token")
// // 	if token == "" {
// // 		conn.WriteJSON(map[string]string{"error": "missing token"})
// // 		return
// // 	}
// // 	userID, _, err := h.chatUsecase.ValidateTokenWithRedis(token)
// // 	if err != nil {
// // 		conn.WriteJSON(map[string]string{"error": "invalid token: " + err.Error()})
// // 		return
// // 	}
// // 	log.Printf("User %d connected with token", userID)

// // 	// Determine subscription channels
// // 	// groupIDStr := r.URL.Query().Get("group_id")
// // 	groupIDStr := "5"
// // 	log.Printf("%T", groupIDStr)
// // 	log.Println("g----------------------------------------------------------p---", groupIDStr)
// // 	var channels []string
// // 	if groupIDStr != "" {
// // 		groupID, err := strconv.Atoi(groupIDStr)
// // 		if err != nil {
// // 			conn.WriteJSON(map[string]string{"error": "invalid group_id"})
// // 			return
// // 		}
// // 		channels = []string{"group:" + strconv.Itoa(groupID), "user:" + strconv.Itoa(userID)}
// // 		log.Printf("Subscribing user %d to channels: %v", userID, channels)
// // 	} else {
// // 		channels = []string{"user:" + strconv.Itoa(userID)}
// // 		log.Println("HELLOOOOOOOO,ELSE")
// // 	}

// // 	// Subscribe to Redis channels
// // 	msgChan, err := h.redisService.SubscribeToMultipleChannels(channels)

// // 	log.Println("............................", msgChan)
// // 	if err != nil {
// // 		conn.WriteJSON(map[string]string{"error": "subscription failed: " + err.Error()})
// // 		return
// // 	}

// // 	// Handle client messages and Redis events concurrently
// // 	done := make(chan struct{})
// // 	go func() {
// // 		defer close(done)
// // 		for {
// // 			// Read client messages
// // 			_, clientMsg, err := conn.ReadMessage()
// // 			log.Println("read---------", string(clientMsg))
// // 			if err != nil {
// // 				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// // 					log.Printf("WebSocket read error for user %d: %v", userID, err)
// // 				}
// // 				return
// // 			}

// // 			type clientRequest struct {
// // 				Type       string `json:"type"` // Fixed from "targetType"
// // 				ReceiverID int    `json:"targetId,omitempty"`
// // 				GroupID    int    `json:"group_id,omitempty"`
// // 				Content    string `json:"content"`
// // 				SenderID   int    `json:"sender_id,omitempty"` // Add for validation
// // 			}
// // 			var req clientRequest

// // 			if err := json.Unmarshal(clientMsg, &req); err != nil {
// // 				conn.WriteJSON(map[string]string{"error": "invalid message format"})
// // 				continue
// // 			}

// // 			log.Println("reww==", token, req.GroupID, req.Content)
// // 			log.Println("req--", req)
// // 			if req.SenderID == 0 {
// // 				req.SenderID = userID // Default to authenticated userID
// // 			} else if req.SenderID != userID {
// // 				conn.WriteJSON(map[string]string{"error": "sender_id mismatch"})
// // 				continue
// // 			}

// // 			switch req.Type {
// // 			case "personal_message":
// // 				if err := h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content); err != nil {
// // 					conn.WriteJSON(map[string]string{"error": "failed to send personal message: " + err.Error()})
// // 				}
// // 			case "group_message":
// // 				if err := h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content); err != nil {
// // 					conn.WriteJSON(map[string]string{"error": "failed to send group message: " + err.Error()})
// // 				}
// // 			case "voice_message":
// // 				var receiverID, groupID *int
// // 				if req.ReceiverID != 0 {
// // 					receiverID = &req.ReceiverID
// // 				}
// // 				if req.GroupID != 0 {
// // 					groupID = &req.GroupID
// // 				}
// // 				if err := h.chatUsecase.SendVoiceMessage(token, receiverID, groupID, req.Content); err != nil {
// // 					conn.WriteJSON(map[string]string{"error": "failed to send voice message: " + err.Error()})
// // 				}
// // 			default:
// // 				conn.WriteJSON(map[string]string{"error": "unknown message type"})
// // 			}
// // 		}
// // 	}()

// // 	for {
// // 		select {
// // 		case msg, ok := <-msgChan:
// // 			if !ok {
// // 				log.Printf("Redis channel closed for user %d", userID)
// // 				return
// // 			}
// // 			log.Printf("Received from Redis for user %d: %s", userID, msg)
// // 			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
// // 				log.Printf("WebSocket write error for user %d: %v", userID, err)
// // 				return
// // 			}
// // 		case <-done:
// // 			return
// // 		}
// // 	}
// // }
// //------------------down working
// func (h *ChatWebSocketHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("WebSocket upgrade failed: %v", err)
// 		return
// 	}
// 	defer conn.Close()

// 	log.Printf("http upgraded to ws")

// 	token := r.URL.Query().Get("token")
// 	if token == "" {
// 		conn.WriteJSON(map[string]string{"error": "missing token"})
// 		return
// 	}
// 	userID, _, err := h.chatUsecase.ValidateTokenWithRedis(token)
// 	if err != nil {
// 		conn.WriteJSON(map[string]string{"error": "invalid token: " + err.Error()})
// 		return
// 	}
// 	log.Printf("User %d connected with token", userID)

// 	// Subscribe to channels (existing logic)
// 	groupIDStr := "5" // Hardcoded for now
// 	channels := []string{"group:" + groupIDStr, "user:" + strconv.Itoa(userID)}
// 	log.Printf("Subscribing user %d to channels: %v", userID, channels)
// 	msgChan, err := h.redisService.SubscribeToMultipleChannels(channels)
// 	if err != nil {
// 		conn.WriteJSON(map[string]string{"error": "subscription failed: " + err.Error()})
// 		return
// 	}

// 	// Handle messages
// 	done := make(chan struct{})
// 	go func() {
// 		defer close(done)
// 		for {
// 			_, clientMsg, err := conn.ReadMessage()
// 			log.Println("read---------", string(clientMsg))
// 			if err != nil {
// 				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 					log.Printf("WebSocket read error for user %d: %v", userID, err)
// 				}
// 				return
// 			}

// 			var req struct {
// 				Type       string `json:"type"`
// 				ReceiverID int    `json:"receiver_id,omitempty"`
// 				GroupID    int    `json:"group_id,omitempty"`
// 				Content    string `json:"content"`
// 				SenderID   int    `json:"sender_id,omitempty"`
// 				Filename   string `json:"filename,omitempty"`
// 				Filetype   string `json:"filetype,omitempty"`
// 			}
// 			if err := json.Unmarshal(clientMsg, &req); err != nil {
// 				conn.WriteJSON(map[string]string{"error": "invalid message format"})
// 				continue
// 			}

// 			if req.SenderID == 0 {
// 				req.SenderID = userID
// 			} else if req.SenderID != userID {
// 				conn.WriteJSON(map[string]string{"error": "sender_id mismatch"})
// 				continue
// 			}

// 			switch req.Type {
// 			case "personal_message", "group_message":
// 				if err := h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content); err != nil {
// 					conn.WriteJSON(map[string]string{"error": "failed to send message: " + err.Error()})
// 				}
// 			case "audio_message", "file_message":
// 				if req.GroupID == 0 && req.ReceiverID == 0 {
// 					conn.WriteJSON(map[string]string{"error": "group_id or receiver_id required"})
// 					continue
// 				}
// 				if err := h.chatUsecase.SendMultimediaMessage(token, req.ReceiverID, req.GroupID, req.Content, req.Filename, req.Filetype, req.Type); err != nil {
// 					conn.WriteJSON(map[string]string{"error": "failed to send multimedia: " + err.Error()})
// 				}
// 			default:
// 				conn.WriteJSON(map[string]string{"error": "unknown message type"})
// 			}
// 		}
// 	}()

// 	// Subscribe loop (existing)
// 	for {
// 		select {
// 		case msg, ok := <-msgChan:
// 			if !ok {
// 				log.Printf("Redis channel closed for user %d", userID)
// 				return
// 			}
// 			log.Printf("Received from Redis for user %d: %s", userID, msg)
// 			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
// 				log.Printf("WebSocket write error for user %d: %v", userID, err)
// 				return
// 			}
// 		case <-done:
// 			return
// 		}
// 	}
// }

//	func RegisterWebSocketRoute(mux *http.ServeMux, chatUsecase usecase.ChatUsecase, redisService usecase.RedisService) {
//		handler := NewChatWebSocketHandler(chatUsecase, redisService)
//		mux.HandleFunc("/ws/chat", handler.HandleChat)
//	}
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("HTTP upgraded to WebSocket")

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

	// Determine subscription channels dynamically
	//groupIDStr := r.URL.Query().Get("group_id")
	groupIDStr := "5"

	log.Println("????????????????????????????????????????????????????????????????????????????", groupIDStr)
	// Remove hardcoded value
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
		log.Printf("Subscribing user %d to personal channel only", userID)
	}

	// Subscribe to Redis channels
	msgChan, err := h.redisService.SubscribeToMultipleChannels(channels)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "subscription failed: " + err.Error()})
		return
	}

	// Notify online status and fetch unread messages
	if err := h.chatUsecase.UpdateUserStatus(userID, "online"); err != nil {
		log.Printf("Failed to update status for user %d: %v", userID, err)
	}
	unreadMessages, err := h.chatUsecase.GetUnreadMessages(userID)
	if err != nil {
		log.Printf("Failed to fetch unread messages for user %d: %v", userID, err)
	} else {
		for _, msg := range unreadMessages {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send unread message to user %d: %v", userID, err)
			}
		}
	}

	// Handle client messages and Redis events concurrently
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, clientMsg, err := conn.ReadMessage()
			log.Printf("Received client message from user %d: %s", userID, string(clientMsg))
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error for user %d: %v", userID, err)
				}
				return
			}

			var req struct {
				Type       string  `json:"type"`
				ReceiverID int     `json:"receiver_id,omitempty"`
				GroupID    int     `json:"group_id,omitempty"`
				Content    string  `json:"content"`
				SenderID   int     `json:"sender_id,omitempty"`
				Filename   *string `json:"filename,omitempty"`
				Filetype   *string `json:"filetype,omitempty"`
				ReplyTo    *int    `json:"reply_to,omitempty"`   // For reply functionality
				ForwardTo  *int    `json:"forward_to,omitempty"` // For forward functionality
			}
			if err := json.Unmarshal(clientMsg, &req); err != nil {
				conn.WriteJSON(map[string]string{"error": "invalid message format"})
				continue
			}

			if req.SenderID == 0 {
				req.SenderID = userID
			} else if req.SenderID != userID {
				conn.WriteJSON(map[string]string{"error": "sender_id mismatch"})
				continue
			}

			switch req.Type {
			case "personal_message", "group_message":
				var err error
				if req.Type == "personal_message" {
					err = h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content)
				} else {
					err = h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content)
				}
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send message: " + err.Error()})
				}
				// Update status to delivered after sending
				if err == nil && req.ReceiverID != 0 {
					h.chatUsecase.UpdateUserStatus(req.ReceiverID, "delivered")
				}

			case "audio_message", "file_message":
				if req.GroupID == 0 && req.ReceiverID == 0 {
					conn.WriteJSON(map[string]string{"error": "group_id or receiver_id required"})
					continue
				}
				err := h.chatUsecase.SendMultimediaMessage(token, req.ReceiverID, req.GroupID, req.Content, *req.Filename, *req.Filetype, req.Type)
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send multimedia: " + err.Error()})
				}
				// Update status to delivered
				if err == nil && req.ReceiverID != 0 {
					h.chatUsecase.UpdateUserStatus(req.ReceiverID, "delivered")
				}

			case "reply":
				if req.ReplyTo == nil || req.Content == "" {
					conn.WriteJSON(map[string]string{"error": "reply_to and content required"})
					continue
				}
				err := h.chatUsecase.SendReplyMessage(token, req.ReceiverID, req.GroupID, req.Content, *req.ReplyTo)
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send reply: " + err.Error()})
				}

			case "forward":
				if req.ForwardTo == nil || req.Content == "" {
					conn.WriteJSON(map[string]string{"error": "forward_to and content required"})
					continue
				}
				err := h.chatUsecase.ForwardMessage(token, *req.ForwardTo, req.Content)
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to forward message: " + err.Error()})
				}

			case "mark_seen":
				chatType := "user"
				chatID := req.ReceiverID
				if req.GroupID != 0 {
					chatType = "group"
					chatID = req.GroupID
				}
				if err := h.chatUsecase.MarkMessagesAsSeen(chatType, chatID, userID); err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to mark as seen: " + err.Error()})
				}

			default:
				conn.WriteJSON(map[string]string{"error": "unknown message type"})
			}
		}
	}()

	// Handle Redis messages
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				log.Printf("Redis channel closed for user %d", userID)
				return
			}
			log.Printf("Received from Redis for user %d: %s", userID, msg)
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &data); err != nil {
				log.Printf("Failed to unmarshal Redis message: %v", err)
				continue
			}

			// Handle status updates (e.g., delivered, seen)
			if status, ok := data["status"].(string); ok {
				if status == "delivered" || status == "seen" {
					if err := conn.WriteJSON(data); err != nil {
						log.Printf("Failed to send status update to user %d: %v", userID, err)
					}
				}
			} else {
				// Broadcast new messages as notifications
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Printf("WebSocket write error for user %d: %v", userID, err)
					return
				}
			}
		case <-done:
			if err := h.chatUsecase.UpdateUserStatus(userID, "offline"); err != nil {
				log.Printf("Failed to update offline status for user %d: %v", userID, err)
			}
			return
		}
	}
}

func RegisterWebSocketRoute(mux *http.ServeMux, chatUsecase usecase.ChatUsecase, redisService usecase.RedisService) {
	handler := NewChatWebSocketHandler(chatUsecase, redisService)
	mux.HandleFunc("/ws/chat", handler.HandleChat)
}
