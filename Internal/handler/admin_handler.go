package handler

import (
	"log"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) DeleteUser(c *gin.Context) {

	log.Println("hiiii")
	type request struct {
		UserID int `json:"user_id" binding:"required"`
	}

	var req request

	log.Println(req.UserID)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	err := h.userUsecase.DeleteUser(token, req.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "user deactivated successfully"})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	type request struct {
		UserID int    `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	log.Println("work")
	token := c.GetHeader("Authorization")
	log.Println("dd", token)
	user := &domain.User{
		Email: req.Email,
		Name:  req.Name,
		Role:  req.Role,
	}

	err := h.userUsecase.UpdateUser(token, req.UserID, user)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "user updated successfully"})
}
