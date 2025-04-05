package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService usecase.AuthService, redisService usecase.RedisService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Println("////////////////", authHeader)
		log.Printf("Middleware - Raw Authorization header: %q", authHeader)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Middleware - Invalid or missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr) // Remove extra whitespace
		log.Printf("Middleware - Parsed token: %q", tokenStr)

		userID, role, err := authService.ValidateToken(tokenStr)
		if err != nil {
			log.Printf("Middleware - Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
			c.Abort()
			return
		}

		redisToken, err := redisService.GetToken(userID)
		if err != nil || redisToken == "" || redisToken != tokenStr {
			log.Printf("Middleware - Token mismatch or not active: Redis=%q, Request=%q", redisToken, tokenStr)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token not active or mismatched"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// func AuthMiddleware(authService usecase.AuthService, redisService usecase.RedisService) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")

// 		log.Println(authHeader)
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
// 			c.Abort()
// 			return
// 		}

// 		// Expect "Bearer <token>"
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
// 			c.Abort()
// 			return
// 		}
// 		token := parts[1]

// 		// Validate token with AuthService
// 		userID, role, err := authService.ValidateToken(token)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
// 			c.Abort()
// 			return
// 		}

// 		// Check token in Redis
// 		storedToken, err := redisService.GetToken(userID)
// 		if err != nil || storedToken == "" || storedToken != token {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "token not active or mismatched"})
// 			c.Abort()
// 			return
// 		}

// 		// Set userID and role in context for handlers
// 		c.Set("userID", userID)
// 		c.Set("role", role)
// 		c.Next()
// 	}
// }

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
