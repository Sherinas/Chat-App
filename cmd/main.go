package main

import (
	"log"
	"net/http"
	"os"
	"time"

	database "github.com/Sherinas/Chat-App-Clean/Internal/infrastructure"
	"github.com/Sherinas/Chat-App-Clean/Internal/infrastructure/auth"
	"github.com/Sherinas/Chat-App-Clean/Internal/infrastructure/redis"
	"github.com/Sherinas/Chat-App-Clean/Internal/infrastructure/websocket"
	"github.com/Sherinas/Chat-App-Clean/Internal/repository"
	"github.com/Sherinas/Chat-App-Clean/Internal/routes"
	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize dependencies
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	} else {
		log.Println("Loaded .env successfully")
	}
	db, _ := database.NewDB()
	redisService := redis.NewRedisService("localhost:6379", "")
	authService := auth.NewJWTService("your-secret-key")

	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	groupRequestRepo := repository.NewGroupRequestRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	userUsecase := usecase.NewUserUsecase(userRepo, otpRepo, authService, redisService)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, groupRequestRepo, authService, redisService)
	chatUsecase := usecase.NewChatUsecase(messageRepo, userRepo, groupRepo, authService, redisService)

	seedAdmin(userUsecase)

	// Set up Gin router for HTTP
	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"}) // Fix trusted proxies warning

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://127.0.0.1:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterUserRoutes(r, *userUsecase, authService, redisService, *groupUsecase)
	routes.RegisterGroupRoutes(r, *groupUsecase, authService, redisService)
	routes.RegisterChatRoutes(r, *chatUsecase, authService, redisService)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Static("/static", "./public")
	r.StaticFile("auth/login", "./public/login.html")
	r.StaticFile("users/signup", "./public/signup.html")
	r.StaticFile("users/otp", "./public/otp.html")
	r.StaticFile("users/resetPassword", "./public/resetPassword.html")
	r.StaticFile("dashboard", "./public/user_dashbord.html")
	r.StaticFile("users/profile", "./public/profile.html")
	//r.StaticFile("/admin", "./public/admin.html")

	// Set up WebSocket
	mux := http.NewServeMux()
	websocket.RegisterWebSocketRoute(mux, *chatUsecase, redisService)

	// Start HTTP server
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start WebSocket server
	go func() {
		log.Println("Starting WebSocket server on :8081")
		server := &http.Server{
			Addr:    ":8081",
			Handler: mux,
		}
		// Log before listening to confirm setup
		log.Println("WebSocket server about to listen on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server failed: %v", err)
		}
		log.Println("WebSocket server started on :8081")
	}()

	log.Println("start APP")
	select {}
}

func seedAdmin(userUsecase *usecase.UserUsecase) {
	employeeID := os.Getenv("ADMIN_EMPLOYEE_ID")
	name := os.Getenv("ADMIN_NAME")
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	if employeeID == "" || name == "" || email == "" || password == "" {
		log.Println("Admin credentials not provided in env, skipping admin creation")
		return
	}

	token, err := userUsecase.CreateAdminUser(employeeID, name, email, password)
	if err != nil {
		if err.Error() == "user already exists" {
			log.Printf("Admin user %s already exists", employeeID)
			return
		}
		log.Fatalf("Failed to create admin user: %v", err)
	}
	log.Printf("Admin user %'Connells created with token: %s", employeeID, token)
}
