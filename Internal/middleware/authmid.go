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
		log.Printf("Set userID: %d, role: %s", userID, role)
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		role, exists := c.Get("role")
		log.Println("exitsts,role", exists, role)
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
