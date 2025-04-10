package handler

import (
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

	token, id, err := h.userUsecase.LoginWithEmployeeID(req.EmployeeID, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.SetCookie("token", token, 3600, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"user_id": id, "token": token, "message": "login successful"})
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

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

// func (h *UserHandler)Logout(c *gin.Context) {
// 	// Extract the token from the Authorization header
// 	authHeader := c.GetHeader("Authorization")
// 	if authHeader == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token provided"})
// 		return
// 	}

// 	// Validate the token to get user ID and expiration
// 	userID, claims, err := h.authService.ValidateToken(authHeader)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 		return
// 	}

// 	// Extract token expiration from claims (assuming claims include exp)
// 	exp, ok := claims["exp"].(float64)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token expiration"})
// 		return
// 	}
// 	expirationTime := time.Unix(int64(exp), 0)
// 	ttl := time.Until(expirationTime)

// 	// Invalidate the token by storing it in Redis blacklist
// 	token := authHeader[len("Bearer "):] // Remove "Bearer " prefix
// 	err = h.redisser.BlacklistToken(token, ttl)
// 	if err != nil {
// 		log.Printf("Failed to blacklist token for user %d: %v", userID, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout: " + err.Error()})
// 		return
// 	}

// 	// Respond with success and instruct client to clear token
// 	c.JSON(http.StatusOK, gin.H{
// 		"message":  "Logged out successfully",
// 		"redirect": "/users/login", // Suggest client to redirect
// 	})
// }
