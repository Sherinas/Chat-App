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
	handler := handler.NewGroupHandler(groupUsecase, authService)
	authMiddleware := middleware.AuthMiddleware(authService, redisService)
	adminMiddleware := middleware.AdminMiddleware()

	groupGroup := router.Group("/groups")
	{
		// Protected routes (all require auth, some admin-only)
		groupGroup.Use(authMiddleware)
		groupGroup.POST("/", adminMiddleware, handler.CreateGroup)
		groupGroup.DELETE("/delete-group", adminMiddleware, handler.DeleteGroup)
		groupGroup.POST("/join", authMiddleware, handler.RequestToJoinGroup)
		groupGroup.POST("/approve", adminMiddleware, handler.ApproveGroupRequest)
		groupGroup.POST("/reject", adminMiddleware, handler.RejectGroupRequest)
		groupGroup.POST("/add-user", adminMiddleware, handler.AddUserToGroup)
		groupGroup.POST("/remove-user", adminMiddleware, handler.RemoveUserFromGroup)
		groupGroup.GET("/profile/groups", authMiddleware, handler.GetAllGroupswithMember)

	}
}

// func RegisterUserRoutes(router *gin.Engine, userUsecase usecase.UserUsecase, authService usecase.AuthService, redisService usecase.RedisService, groupUsecase usecase.GroupUsecase) {
// 	handler := handler.NewUserHandler(userUsecase, redisService, authService, groupUsecase)
// 	authMiddleware := middleware.AuthMiddleware(authService, redisService)
// 	adminMiddleware := middleware.AdminMiddleware()

// 	userGroup := router.Group("/users")
// 	{
// 		// Public routes
// 		userGroup.POST("/signup", handler.SignUpWithEmployeeID)
// 		userGroup.POST("/verify-otp", handler.VerifyOTP)

// 		//profile
// 		userGroup.GET("/me", authMiddleware, handler.GetProfile)
// 		userGroup.PATCH("/me-resetpassword", authMiddleware, handler.ResetPasswordfromProfile)
// 		userGroup.PATCH("/me/dp", authMiddleware, handler.UploadProfilePhoto)

// 		//UserMgmt

// 		userGroup.PUT("/update-user", adminMiddleware, handler.UpdateUser)
// 		userGroup.DELETE("/delete-user", adminMiddleware, handler.DeleteUser)

// 		//userGroup.GET()
// 		userGroup.POST("/login", handler.LoginWithEmployeeID)
// 		userGroup.GET("/all-users", authMiddleware, handler.GetAllUsers)
// 		userGroup.GET("/user-groups", authMiddleware, handler.GetUserGroups)

// 		//want change the handlerand routes

//			// Protected routes
//			userGroup.Use(authMiddleware)
//			userGroup.POST("/logout", handler.Logout)
//			userGroup.POST("/employee", adminMiddleware, handler.CreateEmployeeID) // Admin-only
//		}
//	}
// func RegisterUserRoutes(router *gin.Engine, userUsecase usecase.UserUsecase, authService usecase.AuthService, redisService usecase.RedisService, groupUsecase usecase.GroupUsecase) {
// 	handler := handler.NewUserHandler(userUsecase, redisService, authService, groupUsecase)
// 	authMiddleware := middleware.AuthMiddleware(authService, redisService)
// 	adminMiddleware := middleware.AdminMiddleware()

// 	userGroup := router.Group("/users", authMiddleware)
// 	{
// 		// Public routes (no admin required, but token validation optional)
// 		userGroup.POST("/signup", handler.SignUpWithEmployeeID)
// 		userGroup.POST("/verify-otp", handler.VerifyOTP)

// 		// Profile routes (require auth)
// 		userGroup.GET("/me", handler.GetProfile)
// 		userGroup.PATCH("/me-resetpassword", handler.ResetPasswordfromProfile)
// 		userGroup.PATCH("/me/dp", handler.UploadProfilePhoto)

// 		// User management routes (admin-only)
// 		userGroup.PUT("/update-user", adminMiddleware, handler.UpdateUser)
// 		userGroup.DELETE("/delete-user", adminMiddleware, handler.DeleteUser)
// 		userGroup.GET("/all", adminMiddleware, handler.GetAllUsers)

// 		// Login and user info routes
// 		userGroup.POST("/login", handler.LoginWithEmployeeID)
// 		userGroup.GET("/all-users", handler.GetAllUsers)
// 		userGroup.GET("/user-groups", handler.GetUserGroups)

//			// Protected routes
//			userGroup.POST("/logout", handler.Logout)
//			userGroup.POST("/employee", adminMiddleware, handler.CreateEmployeeID)
//		}
//	}
func RegisterUserRoutes(router *gin.Engine, userUsecase usecase.UserUsecase, authService usecase.AuthService, redisService usecase.RedisService, groupUsecase usecase.GroupUsecase) {
	handler := handler.NewUserHandler(userUsecase, redisService, authService, groupUsecase)
	authMiddleware := middleware.AuthMiddleware(authService, redisService)
	adminMiddleware := middleware.AdminMiddleware()

	// Public routes (no authentication required)
	userGroupPublic := router.Group("/users")
	{
		userGroupPublic.POST("/signup", handler.SignUpWithEmployeeID)
		userGroupPublic.POST("/verify-otp", handler.VerifyOTP)
		userGroupPublic.POST("/login", handler.LoginWithEmployeeID)
	}

	// Protected routes (require authentication)
	userGroupProtected := router.Group("/users", authMiddleware)
	{
		// Profile routes
		userGroupProtected.GET("/me", handler.GetProfile)
		userGroupProtected.PATCH("/me-resetpassword", handler.ResetPasswordfromProfile)
		userGroupProtected.PATCH("/me/dp", handler.UploadProfilePhoto)

		// User management routes (admin-only)
		userGroupProtected.PUT("/update-user", adminMiddleware, handler.UpdateUser)
		userGroupProtected.DELETE("/delete-user", adminMiddleware, handler.DeleteUser)
		userGroupProtected.GET("/all", adminMiddleware, handler.GetAllUsers)

		// Login and user info routes
		userGroupProtected.GET("/all-users", handler.GetAllUsers)
		userGroupProtected.GET("/user-groups", handler.GetUserGroups)

		// Protected routes
		userGroupProtected.POST("/logout", handler.Logout)
		userGroupProtected.POST("/employee", adminMiddleware, handler.CreateEmployeeID)
	}
}
