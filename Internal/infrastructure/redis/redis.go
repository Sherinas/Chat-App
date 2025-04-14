package redis

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/usecase"
	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisService(addr, password string) usecase.RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	log.Println("radis connected")
	return &RedisService{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisService) SetUserStatus(userID int, status string) error {
	key := "user:status:" + strconv.Itoa(userID) // Fixed
	return r.client.Set(r.ctx, key, status, 24*time.Hour).Err()
}

func (r *RedisService) GetUserStatus(userID int) (string, error) {
	key := "user:status:" + strconv.Itoa(userID) // Fixed
	status, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "offline", nil
	}

	log.Println(status)
	return status, err
}

func (r *RedisService) PublishMessage(channel string, message string) error {

	log.Println("out", message, channel)

	log.Println("publishinggg....")
	return r.client.Publish(r.ctx, channel, message).Err()
}

func (r *RedisService) SubscribeChannel(channel string) (<-chan string, error) {
	pubsub := r.client.Subscribe(r.ctx, channel)
	_, err := pubsub.Receive(r.ctx)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan string)
	go func() {
		defer pubsub.Close() // Clean up subscription
		defer close(msgChan) // Close channel when done
		for msg := range pubsub.Channel() {
			msgChan <- msg.Payload
		}
	}()
	return msgChan, nil
}

func (r *RedisService) SetToken(userID int, token string) error {
	key := "token:" + strconv.Itoa(userID) // Fixed
	return r.client.Set(r.ctx, key, token, 24*time.Hour).Err()
}

func (r *RedisService) GetToken(userID int) (string, error) {
	key := "token:" + strconv.Itoa(userID) // Fixed
	token, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // No token found
	}

	return token, err
}

func (r *RedisService) RemoveToken(userID int) error {
	key := "token:" + strconv.Itoa(userID) // Fixed
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisService) SetWithTTL(key, value string, ttl time.Duration) error {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *RedisService) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

func (r *RedisService) Get(key string) (string, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key not found
	}
	return val, err
}

func (r *RedisService) BlacklistToken(token string, ttl time.Duration) error {
	return r.client.Set(r.ctx, "blacklist:"+token, "1", ttl).Err()
}

func (r *RedisService) SubscribeToMultipleChannels(channels []string) (<-chan string, error) {
	pubsub := r.client.Subscribe(context.Background(), channels...)
	if err := pubsub.Ping(context.Background()); err != nil {
		return nil, err
	}
	msgChan := make(chan string)
	go func() {
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.Printf("Subscription error: %v", err)
				close(msgChan)
				return
			}
			msgChan <- msg.Payload
		}
	}()
	return msgChan, nil
}

func saveFile(content, filename, filetype string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	filePath := "./uploads/" + filename
	if err := os.WriteFile(filePath, decoded, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}
