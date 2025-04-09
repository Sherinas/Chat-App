package handler

import (
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

	log.Println("hello")

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
