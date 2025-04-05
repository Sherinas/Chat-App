package handler

import (
	"net/http"

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
	err := h.chatUsecase.SendPersonalMessage(token, req.ReceiverID, req.Content)
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
	err := h.chatUsecase.SendGroupMessage(token, req.GroupID, req.Content)
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
	err := h.chatUsecase.SendVoiceMessage(token, req.ReceiverID, req.GroupID, req.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "voice message sent"})
}
