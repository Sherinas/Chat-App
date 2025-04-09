package handler

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) GetUserGroups(c *gin.Context) {

	log.Println(" from handler GetUserGroups stsrt")
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	log.Println(" from handler GetUserGroups stsrtaaa", userID)
	userIDInt, ok := userID.(int)
	if !ok {
		log.Printf("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	log.Printf("Fetching groups for userID: %d", userIDInt)
	groups, err := h.gropUse.GetUserGroups(userIDInt) // Fixed typo
	if err != nil {
		log.Printf("Failed to fetch user groups for userID %d: %v", userIDInt, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user groups: " + err.Error()})
		return
	}

	log.Println("oooooooooooooooooooooooooooooooooooooooooooooo", groups)

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// func (h *UserHandler) GetAllUsers(c *gin.Context) {
// 	users, _ := h.userUsecase.GetAllUsers()

// 	userList := make([]map[string]interface{}, len(users))
// 	for i, user := range users {
// 		status, _ := h.redisser.GetUserStatus(user.ID)
// 		userList[i] = map[string]interface{}{
// 			"id":     user.ID,
// 			"name":   user.Name,
// 			"status": status,
// 		}
// 	}
// 	c.JSON(http.StatusOK, gin.H{"users": userList})
// }

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Check if this is the admin-only endpoint (e.g., /all) by inspecting the path
	isAdminEndpoint := c.Request.URL.Path == "/users/all"

	// Get the role from the context (set by AuthMiddleware)
	role, exists := c.Get("role")
	if !exists || (isAdminEndpoint && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	users, err := h.userUsecase.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users: " + err.Error()})
		return
	}

	// Prepare user list with relevant fields
	var userList []map[string]interface{}
	for _, user := range users {
		userData := map[string]interface{}{
			"id":            user.ID,
			"name":          user.Name,
			"email":         user.Email,
			"status":        user.State,
			"created_at":    user.CreatedAt,
			"profile_photo": user.ProfilePhoto,
		}
		if isAdminEndpoint {

			userData["deleted_at"] = user.DeletedAt
			userData["role"] = user.Role
			userData["profile_photo"] = user.ProfilePhoto
		}
		userList = append(userList, userData)
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}

// Profile based code

func (h *UserHandler) GetProfile(c *gin.Context) {

	userID, _, err := h.authService.ValidateToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	user, err := h.userUsecase.FindUserDetails(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch profile: " + err.Error()})
		return
	}
	status, _ := h.redisser.GetUserStatus(userID)
	if status == "" {
		status = "offline"
	}
	c.JSON(http.StatusOK, gin.H{
		"employee_id": user.EmployeeID,
		"name":        user.Name,
		"email":       user.Email,
		"status":      status,
	})
}

func (h *UserHandler) GetAllGroup(c *gin.Context) {

	groups, err := h.gropUse.GetAllGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch groups: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": groups,
	})
}

func (h *UserHandler) ResetPasswordfromProfile(c *gin.Context) {

	userID, _, err := h.authService.ValidateToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	user, err := h.userUsecase.FindUserDetails(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user: " + err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password: " + err.Error()})
		return
	}
	// Update the user password
	user.Password = string(hashedPassword)
	if err := h.userUsecase.UpdateUser(c.GetHeader("Authorization"), userID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *UserHandler) UploadProfilePhoto(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Get the file from the request
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded: " + err.Error()})
		return
	}

	// Validate file type (e.g., only images)
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	if !contains(allowedTypes, file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPEG, PNG, and GIF are allowed"})
		return
	}

	// Validate file size (e.g., max 5MB)
	if file.Size > 5*1024*1024 { // 5MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	// Generate a unique filename
	timestamp := time.Now().Unix()
	filename := filepath.Join("static/uploads", fmt.Sprintf("user%d_%d%s", uid, timestamp, filepath.Ext(file.Filename)))
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	// Update user with the new photo path
	user, err := h.userUsecase.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user: " + err.Error()})
		return
	}
	user.ProfilePhoto = filename
	if err := h.userUsecase.UpdateUser(c.GetHeader("Authorization"), uid, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile photo uploaded successfully", "photoPath": filename})
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
