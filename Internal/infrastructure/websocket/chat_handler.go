package websocket

import (
	"encoding/json"
	"fmt"
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
	useruse      usecase.UserUsecase
}

func NewChatWebSocketHandler(chatUsecase usecase.ChatUsecase, redisService usecase.RedisService, useruse usecase.UserUsecase) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		chatUsecase:  chatUsecase,
		redisService: redisService,
		useruse:      useruse,
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
	groupIDStr := r.URL.Query().Get("group_id")

	log.Println("getting group ID=", groupIDStr)

	//var channels []string
	var initialChannels []string
	if groupIDStr != "" {
		groupID, err := strconv.Atoi(groupIDStr)

		log.Println("getting group covert to int :=", groupIDStr)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "invalid group_id"})
			return
		}
		initialChannels = []string{fmt.Sprintf("group:%d", groupID), fmt.Sprintf("user:%d", userID)}
		//channels = []string{"group:" + strconv.Itoa(groupID), "user:" + strconv.Itoa(userID)}

		//log.Printf("Subscribing user %d to channels: %v", userID, channels)
	} else {
		initialChannels = []string{fmt.Sprintf("user:%d", userID)}
		//channels = []string{"user:" + strconv.Itoa(userID)}
		//log.Printf("USERaaaaaaaaaa%T", channels)
		log.Printf("Subscribing user %d to personal channel only", userID)
	}

	var msgChans []<-chan string
	for _, ch := range initialChannels {
		msgChan, err := h.redisService.SubscribeChannel(ch)
		if err != nil {
			log.Printf("Failed to subscribe to initial channel %s: %v", ch, err)
			conn.WriteJSON(map[string]string{"error": "initial subscription failed: " + err.Error()})
			return
		}
		msgChans = append(msgChans, msgChan)
	}
	if len(msgChans) == 0 {
		conn.WriteJSON(map[string]string{"error": "no valid subscriptions established"})
		return
	}
	// Subscribe to Redis channels
	fmt.Println("working 111")

	//msgChan, err := h.redisService.SubscribeToMultipleChannels(channels)
	// if err != nil {
	// 	conn.WriteJSON(map[string]string{"error": "subscription failed: " + err.Error()})
	// 	return
	// }
	//	fmt.Println("working 122", msgChan)
	// Notify online status and fetch unread messages

	msgChan := make(chan string)
	for _, ch := range msgChans {
		go func(c <-chan string) {
			for msg := range c {
				msgChan <- msg
			}
		}(ch)
	}

	if err := h.useruse.SetUserState(userID, "online"); err != nil {
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
	go func() {
		log.Printf("Starting Redis listener goroutine for user %d", userID)
		for msg := range msgChan {
			log.Printf("Received from Redis for user %d on channels %v: %s", userID, initialChannels, msg)

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &data); err != nil {
				log.Printf("Failed to unmarshal Redis message for user %d: %v, raw: %s", userID, err, msg)
				continue
			}
			log.Printf("Unmarshalled Redis message for user %d: %#v", userID, data)

			if err := conn.WriteJSON(data); err != nil {
				log.Printf("Failed to send Redis message to user %d: %v", userID, err)
				continue
			}
			log.Printf("Successfully sent Redis message to user %d", userID)
		}
		log.Printf("Redis listener stopped for user %d", userID)
	}()

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
				log.Println("Received message:...............1", string(clientMsg))
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
				var messageID int
				if req.Type == "personal_message" {
					// err = h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content)
					messageID, err = h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content)

				} else {
					messageID, err = h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content)

					log.Println("done")
				}
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send message: " + err.Error()})
				}

				// Update status to delivered after sending

				log.Println("testttt", messageID)
				// if err == nil && req.ReceiverID != 0 {
				if err == nil && messageID != 0 {

					log.Println("test1", req.ReceiverID, err)
					h.chatUsecase.UpdateMessageStatus(messageID, "delivered")

					newChannel := fmt.Sprintf("group:%d", req.GroupID)
					if !contains(initialChannels, newChannel) {
						log.Printf("Adding new subscription for channel %s", newChannel)
						newMsgChan, err := h.redisService.SubscribeChannel(newChannel)
						if err != nil {
							log.Printf("Failed to subscribe to new channel %s: %v", newChannel, err)
							continue
						}
						go func(c <-chan string) {
							for msg := range c {
								msgChan <- msg
							}
						}(newMsgChan)
						initialChannels = append(initialChannels, newChannel)
					}
				}

			case "audio_message", "image_message", "file_message":
				var receiverID, groupID *int
				if req.ReceiverID != 0 {
					id := req.ReceiverID
					receiverID = &id
				}
				if req.GroupID != 0 {
					id := req.GroupID
					groupID = &id
				}
				if receiverID == nil && groupID == nil {
					conn.WriteJSON(map[string]string{"error": "group_id or receiver_id required"})
					continue
				}
				messageID, err := h.chatUsecase.SendMultimediaMessage(token, req.ReceiverID, req.GroupID, req.Content, *req.Filename, *req.Filetype, req.Type)
				if err != nil {
					conn.WriteJSON(map[string]string{"error": "failed to send multimedia: " + err.Error()})
				}
				if err == nil && messageID != 0 {
					if err := h.chatUsecase.UpdateMessageStatus(messageID, "delivered"); err != nil {
						log.Printf("Failed to update message status for message %d: %v", messageID, err)
					}
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

	log.Println("Starting WebSocket listener loop for user", userID, msgChan)

	if msgChan == nil {
		log.Println("DEBUG: msgChan is nil!")
	}

	if done == nil {
		log.Println("DEBUG: done channel is nil!")
	}

	for {
		select {
		case msg, ok := <-msgChan:
			log.Println("DEBUG: Entered msgChan case")

			if !ok {
				log.Printf("DEBUG: msgChan closed for user %d", userID)
				return
			}

			log.Printf("DEBUG: Raw message from Redis for user %d: %s", userID, msg)

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &data); err != nil {
				log.Printf("DEBUG: Failed to unmarshal message: %v", err)
				continue
			}

			log.Printf("DEBUG: Unmarshalled JSON: %#v", data)

			if status, ok := data["status"].(string); ok {
				log.Printf("DEBUG: Message is a status update: %s", status)

				if status == "sent" || status == "seen" || status == "delivered" {
					if err := conn.WriteJSON(data); err != nil {
						log.Printf("ERROR: Failed to send status update to user %d: %v", userID, err)
					} else {
						log.Printf("DEBUG: Sent status update to user %d", userID)
					}
				} else {
					log.Println("DEBUG: Status is not a handled type")
				}
			} else {
				log.Println("DEBUG: No status field, treating as a message")

				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Printf("ERROR: Failed to send message to user %d: %v", userID, err)
					return
				} else {
					log.Printf("DEBUG: Sent message to user %d", userID)
				}
			}

		case <-done:
			log.Println("DEBUG: Received signal from done channel")

			if err := h.useruse.SetUserState(userID, "offline"); err != nil {
				log.Printf("ERROR: Failed to update offline status for user %d: %v", userID, err)
			} else {
				log.Printf("DEBUG: Updated user %d to offline", userID)
			}
			return
		}
	}

	// for {
	// 	select {
	// 	case msg, ok := <-msgChan:

	// 		log.Println("============", msg)

	// 		if !ok {
	// 			log.Printf("Redis channel closed for user %d", userID)
	// 			return
	// 		}
	// 		log.Printf("Received from Redis for---- user %d: %s", userID, msg)
	// 		var data map[string]interface{}
	// 		if err := json.Unmarshal([]byte(msg), &data); err != nil {
	// 			log.Printf("Failed to unmarshal Redis message: %v", err)
	// 			continue
	// 		}

	// 		// Handle status updates (e.g., delivered, seen)
	// 		log.Println("sttissssssssssssssssssssssss")
	// 		if status, ok := data["status"].(string); ok {
	// 			log.Println("dddata..........................................", status)
	// 			if status == "sent" || status == "seen" || status == "delivered" {

	// 				if err := conn.WriteJSON(data); err != nil {
	// 					log.Printf("Failed to send status update to user %d: %v", userID, err)
	// 				}
	// 			}
	// 		} else {
	// 			log.Println("sttisssssssssssssssssssssss............s")
	// 			// Broadcast new messages as notifications
	// 			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
	// 				log.Printf("WebSocket write error for user %d: %v", userID, err)
	// 				return
	// 			}
	// 		}
	// 	case <-done:
	// 		if err := h.useruse.SetUserState(userID, "offline"); err != nil {
	// 			log.Printf("Failed to update offline status for user %d: %v", userID, err)
	// 		}
	// 		return
	// 	}
	// }
}

func RegisterWebSocketRoute(mux *http.ServeMux, chatUsecase usecase.ChatUsecase, redisService usecase.RedisService, useruse usecase.UserUsecase) {
	handler := NewChatWebSocketHandler(chatUsecase, redisService, useruse)
	mux.HandleFunc("/ws/chat", handler.HandleChat)
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
