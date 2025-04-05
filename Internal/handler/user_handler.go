package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
	authService usecase.AuthService
	gropUse     usecase.GroupUsecase
	redisser    usecase.RedisService
	Chats       usecase.ChatUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase, redisser usecase.RedisService, authService usecase.AuthService, gropUse usecase.GroupUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, redisser: redisser, authService: authService, gropUse: gropUse}
}

func (h *UserHandler) CreateEmployeeID(c *gin.Context) {
	type request struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	employeeID, err := h.userUsecase.CreateEmployeeID(req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"employee_id": employeeID})
}

func (h *UserHandler) SignUpWithEmployeeID(c *gin.Context) {
	type request struct {
		EmployeeID  string `json:"employee_id" binding:"required"`
		Password    string `json:"password" binding:"required"`
		Mobile      string `json:"mobile" binding:"required"`
		Designation string `json:"designation" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	otp, err := h.userUsecase.SignUpWithEmployeeID(req.EmployeeID, req.Password, req.Mobile, req.Designation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"otp": otp, "message": "OTP sent, verify to complete signup"})
}

func (h *UserHandler) VerifyOTP(c *gin.Context) {
	type request struct {
		OTP string `json:"otp" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token, err := h.userUsecase.VerifyOTP(req.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "signup completed"})
}

func (h *UserHandler) LoginWithEmployeeID(c *gin.Context) {
	type request struct {
		EmployeeID string `json:"employee_id" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	token, err := h.userUsecase.LoginWithEmployeeID(req.EmployeeID, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.SetCookie("token", token, 3600, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "login successful"})
}

func (h *UserHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization") // Middleware ensures this exists
	err := h.userUsecase.Logout(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

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

//	func (h *UserHandler) GetUserGroups(c *gin.Context) {
//		token := c.GetHeader("Authorization")
//		userID, _, err := h.authService.ValidateToken(token)
//		if err != nil {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
//			return
//		}
//		groups, err := h.gropUse.GetUserGroups(userID)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		groupList := make([]map[string]interface{}, len(groups))
//		for i, group := range groups {
//			groupList[i] = map[string]interface{}{
//				"id":   group.ID,
//				"name": group.Name,
//				// "memberCount": group.MemberCount,
//			}
//		}
//		c.JSON(http.StatusOK, gin.H{"groups": groupList})
//	}
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, _ := h.userUsecase.GetAllUsers()

	fmt.Println("userssssssss GETAlll", users)
	userList := make([]map[string]interface{}, len(users))
	for i, user := range users {
		status, _ := h.redisser.GetUserStatus(user.ID)
		userList[i] = map[string]interface{}{
			"id":     user.ID,
			"name":   user.Name,
			"status": status,
		}
	}
	c.JSON(http.StatusOK, gin.H{"users": userList})
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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

// func (h *UserHandler) GetMessages(c *gin.Context) {
// 	userID, _, err := h.authService.ValidateToken(c.GetHeader("Authorization"))
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 		return
// 	}
// 	messages, err := h.Chats.GetUserMessages(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages: " + err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"messages": messages})
// }

// // SendMessage sends a chat message
// func (h *UserHandler) SendMessage(c *gin.Context) {
// 	userID, _, err := h.authService.ValidateToken(c.GetHeader("Authorization"))
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 		return
// 	}
// 	var req struct {
// 		Message string `json:"message" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
// 		return
// 	}
// 	if err := h.chatUsecase.SendMessage(userID, req.Message); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message: " + err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"success": true})
// }
