package usecase

import "time"

type RedisService interface {
	SetUserStatus(userID int, status string) error
	GetUserStatus(userID int) (string, error)
	PublishMessage(channel string, message string) error
	SubscribeChannel(channel string) (<-chan string, error)
	SetToken(userID int, token string) error // New
	GetToken(userID int) (string, error)     // New
	RemoveToken(userID int) error
	SetWithTTL(key, value string, ttl time.Duration) error // New
	Delete(key string) error
	Get(key string) (string, error)
	BlacklistToken(token string, ttl time.Duration) error
	SubscribeToMultipleChannels(channels []string) (<-chan string, error)
}
