package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatUsecase usecase.ChatUsecase
}

func NewChatHandler(chatUsecase usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{chatUsecase: chatUsecase}
}

func (h *ChatHandler) SendPersonalMessage(c *gin.Context) {
	type request struct {
		ReceiverID int    `json:"receiver_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	log.Println(token)
	_, err := h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "personal message sent"})
}

func (h *ChatHandler) SendGroupMessage(c *gin.Context) {
	type request struct {
		GroupID int    `json:"group_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	_, err := h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group message sent"})
}

func (h *ChatHandler) SendVoiceMessage(c *gin.Context) {
	type request struct {
		ReceiverID *int   `json:"receiver_id"`
		GroupID    *int   `json:"group_id"`
		FilePath   string `json:"file_path" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.ReceiverID == nil && req.GroupID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either receiver_id or group_id must be provided"})
		return
	}

	token := c.GetHeader("Authorization")
	_, err := h.chatUsecase.SendVoiceMessage(token, req.ReceiverID, req.GroupID, req.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "voice message sent"})
}

func (h *ChatHandler) GetUnreadMessages(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	userID, _, err := h.chatUsecase.ValidateTokenWithRedis(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	messages, err := h.chatUsecase.GetUnreadMessages(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatHandler) GetUnreadMessage(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")[7:]
	userID, _, err := h.chatUsecase.ValidateTokenWithRedis(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	messages, err := h.chatUsecase.GetUnreadMessages(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatHandler) GetMessageHistory(c *gin.Context) {
	chatType := c.DefaultQuery("type", "user")
	var chatIDStr string

	log.Fatalln("*********** ", chatIDStr)
	switch chatType {
	case "user":
		chatIDStr = c.Query("user_id")
		if chatIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required for user chat"})
			return
		}
		receiverIDStr := c.Query("receiver_id")
		if receiverIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id is required for user chat"})
			return
		}
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		receiverID, err := strconv.Atoi(receiverIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
			return
		}
		// Pass both IDs to use case if needed, or handle in repository
		messages, err := h.chatUsecase.GetMessageHistory(chatType, chatID, receiverID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, messages)
	case "group":
		chatIDStr = c.Query("group_id")
		if chatIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required for group chat"})
			return
		}
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		messages, err := h.chatUsecase.GetMessageHistory(chatType, chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, messages)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat type, use 'user' or 'group'"})
		return
	}
}
