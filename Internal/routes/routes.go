package routes

import (
	"github.com/Sherinas/Chat-App-Clean/Internal/handler"
	"github.com/Sherinas/Chat-App-Clean/Internal/middleware"
	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(router *gin.Engine, chatUsecase usecase.ChatUsecase, authService usecase.AuthService, redisService usecase.RedisService) {
	handler := handler.NewChatHandler(chatUsecase)
	authMiddleware := middleware.AuthMiddleware(authService, redisService)

	chatGroup := router.Group("/chat")
	{
		// All chat routes require auth
		chatGroup.Use(authMiddleware)
		chatGroup.POST("/personal", handler.SendPersonalMessage)
		chatGroup.POST("/group", handler.SendGroupMessage)
		chatGroup.POST("/voice", handler.SendVoiceMessage)
	}

}

func RegisterGroupRoutes(router *gin.Engine, groupUsecase usecase.GroupUsecase, authService usecase.AuthService, redisService usecase.RedisService) {
	handler := handler.NewGroupHandler(groupUsecase)
	authMiddleware := middleware.AuthMiddleware(authService, redisService)
	adminMiddleware := middleware.AdminMiddleware()

	groupGroup := router.Group("/groups")
	{
		// Protected routes (all require auth, some admin-only)
		groupGroup.Use(authMiddleware)
		groupGroup.POST("/", adminMiddleware, handler.CreateGroup)
		groupGroup.POST("/join", authMiddleware, handler.RequestToJoinGroup)
		groupGroup.POST("/approve", adminMiddleware, handler.ApproveGroupRequest)
		groupGroup.POST("/reject", adminMiddleware, handler.RejectGroupRequest)
		groupGroup.POST("/add-user", adminMiddleware, handler.AddUserToGroup)
		groupGroup.POST("/remove-user", adminMiddleware, handler.RemoveUserFromGroup)

	}
}

func RegisterUserRoutes(router *gin.Engine, userUsecase usecase.UserUsecase, authService usecase.AuthService, redisService usecase.RedisService, groupUsecase usecase.GroupUsecase) {
	handler := handler.NewUserHandler(userUsecase, redisService, authService, groupUsecase)
	authMiddleware := middleware.AuthMiddleware(authService, redisService)
	adminMiddleware := middleware.AdminMiddleware()

	userGroup := router.Group("/users")
	{
		// Public routes
		userGroup.POST("/signup", handler.SignUpWithEmployeeID)
		userGroup.POST("/verify-otp", handler.VerifyOTP)
		userGroup.GET("/me", authMiddleware, handler.GetProfile)
		//userGroup.GET()
		userGroup.POST("/login", handler.LoginWithEmployeeID)
		userGroup.GET("/all-users", authMiddleware, handler.GetAllUsers)
		userGroup.GET("/user-groups", authMiddleware, handler.GetUserGroups)
		//want change the handlerand routes

		// Protected routes
		userGroup.Use(authMiddleware)
		userGroup.POST("/logout", handler.Logout)
		userGroup.POST("/employee", adminMiddleware, handler.CreateEmployeeID) // Admin-only
	}
}
