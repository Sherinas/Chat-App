package auth

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTService struct {
	secretKey string
}

// NewJWTService creates a new JWTService instance
func NewJWTService(secretKey string) usecase.AuthService {
	return &JWTService{secretKey: secretKey}
}

// GenerateToken creates a JWT for a user
func (s *JWTService) GenerateToken(userID int, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken verifies a JWT and returns userID and role
func (s *JWTService) ValidateToken(tokenString string) (int, string, error) {

	log.Println("running")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(s.secretKey), nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64) // JSON numbers are float64
		if !ok {
			return 0, "", errors.New("invalid user_id in token")
		}
		role, ok := claims["role"].(string)
		if !ok {
			return 0, "", errors.New("invalid role in token")
		}
		return int(userID), role, nil
	}

	return 0, "", errors.New("invalid token")
}

func (s *JWTService) HashPassword(password string) (string, error) {
	// Use bcrypt.DefaultCost (10) or adjust based on performance/security needs
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *JWTService) ComparePasswords(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
