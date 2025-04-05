package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupUsecase usecase.GroupUsecase
}

func NewGroupHandler(groupUsecase usecase.GroupUsecase) *GroupHandler {
	return &GroupHandler{groupUsecase: groupUsecase}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	type request struct {
		Name        string          `json:"name" binding:"required"`
		Permissions map[string]bool `json:"permissions" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Get userID and role from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("CreateGroup - user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Ensure userID is an int (matches your use case)
	userIDInt, ok := userID.(int)
	if !ok {
		log.Println("CreateGroup - user_id is not an integer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Log for debugging
	log.Printf("CreateGroup - Admin userID: %d", userIDInt)

	// Call use case with userID instead of token
	groupID, err := h.groupUsecase.CreateGroup(userIDInt, req.Name, req.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"group_id": groupID, "message": "group created"})
}

// func (h *GroupHandler) CreateGroup(c *gin.Context) {
// 	type request struct {
// 		Name        string          `json:"name" binding:"required"`
// 		Permissions map[string]bool `json:"permissions" binding:"required"`
// 	}
// 	var req request
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
// 		return
// 	}

// 	token := c.GetHeader("Authorization") // Middleware ensures this exists

// 	log.Println("admin tocken", token)
// 	groupID, err := h.groupUsecase.CreateGroup(token, req.Name, req.Permissions)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{"group_id": groupID, "message": "group created"})
// }

func (h *GroupHandler) RequestToJoinGroup(c *gin.Context) {
	type request struct {
		GroupID int `json:"group_id" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	requestID, err := h.groupUsecase.RequestToJoinGroup(token, req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"request_id": requestID, "message": "join request submitted"})
}

func (h *GroupHandler) ApproveGroupRequest(c *gin.Context) {
	requestIDStr := c.Query("request_id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing request_id"})
		return
	}
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	token := c.GetHeader("Authorization")
	err = h.groupUsecase.ApproveGroupRequest(token, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "request approved"})
}

func (h *GroupHandler) RejectGroupRequest(c *gin.Context) {
	requestIDStr := c.Query("request_id")
	if requestIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing request_id"})
		return
	}
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	token := c.GetHeader("Authorization")
	err = h.groupUsecase.RejectGroupRequest(token, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "request rejected"})
}

func (h *GroupHandler) AddUserToGroup(c *gin.Context) {
	type request struct {
		UserID  int `json:"user_id" binding:"required"`
		GroupID int `json:"group_id" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	err := h.groupUsecase.AddUserToGroup(token, req.UserID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user added to group"})
}

func (h *GroupHandler) RemoveUserFromGroup(c *gin.Context) {
	type request struct {
		UserID  int `json:"user_id" binding:"required"`
		GroupID int `json:"group_id" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	err := h.groupUsecase.RemoveUserFromGroup(token, req.UserID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user removed from group"})
}

// func (h *UserHandler) GetAllGroupsBYID(c *gin.Context) {
// 	groups, _ := h.gropUse.GetUserGroups()

// 	c.JSON(http.StatusOK, gin.H{"Groups": groups})
// }
