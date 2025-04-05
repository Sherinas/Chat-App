package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"github.com/Sherinas/Chat-App-Clean/Internal/repository"
)

type ChatUsecase struct {
	messageRepo  repository.MessageRepository
	userRepo     repository.UserRepository
	groupRepo    repository.GroupRepository
	authService  AuthService
	redisService RedisService
}

func NewChatUsecase(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	groupRepo repository.GroupRepository,
	authService AuthService,
	redisService RedisService,
) *ChatUsecase {
	return &ChatUsecase{
		messageRepo:  messageRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		authService:  authService,
		redisService: redisService,
	}
}

func (u *ChatUsecase) ValidateTokenWithRedis(token string) (int, string, error) {
	userID, role, err := u.authService.ValidateToken(token)
	if err != nil {
		return 0, "", fmt.Errorf("jwt validation failed: %w", err)
	}
	storedToken, err := u.redisService.GetToken(userID)
	if err != nil {
		return 0, "", fmt.Errorf("redis error: %w", err)
	}
	if storedToken == "" || storedToken != token {
		return 0, "", errors.New("token not active or mismatched")
	}
	return userID, role, nil
}

func (u *ChatUsecase) SendPersonalMessage(token string, receiverID int, content string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Verify sender and receiver exist (discard unused variables)
	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}
	if _, err := u.userRepo.FindByID(receiverID); err != nil {
		return fmt.Errorf("receiver not found: %w", err)
	}

	// Create and store the message
	message := &domain.Message{
		SenderID:   senderID,
		ReceiverID: &receiverID,
		Content:    content,
		//IsVoice:    false,
		CreatedAt: time.Now(),
	}
	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Broadcast or queue based on receiver status
	receiverStatus, err := u.redisService.GetUserStatus(receiverID)
	if err != nil {
		receiverStatus = "offline" // Default to offline if Redis fails
	}

	event := map[string]interface{}{
		//"type":        "personal_message",
		//"message_id":  message.ID,
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"content":     content,
		"created_at":  message.CreatedAt,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		// Log: log.Printf("Failed to marshal event: %v", err)
		return nil // Don’t fail the operation
	}

	if receiverStatus == "online" {
		u.redisService.PublishMessage("user:"+strconv.Itoa(receiverID), string(eventJSON))
	} else {
		u.redisService.PublishMessage("user:queue:"+strconv.Itoa(receiverID), string(eventJSON))
	}

	return nil
}

func (u *ChatUsecase) SendGroupMessage(token string, groupID int, content string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Verify sender exists (discard unused sender)
	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Verify group exists and sender has permission
	group, err := u.groupRepo.FindByID(groupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}
	inGroup := false
	for _, member := range group.Members {
		if member.ID == senderID {
			inGroup = true
			break
		}
	}
	if !inGroup {
		return errors.New("sender not in group")
	}
	if !group.Permission["can_send"] {
		return errors.New("group does not allow sending messages")
	}

	// Create and store the message
	message := &domain.Message{
		SenderID:  senderID,
		GroupID:   &groupID,
		Content:   content,
		IsVoice:   false,
		CreatedAt: time.Now(),
	}
	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Broadcast to group channel
	event := map[string]interface{}{
		"type":       "group_message",
		"message_id": message.ID,
		"group_id":   groupID,
		"sender_id":  senderID,
		"content":    content,
		"created_at": message.CreatedAt,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		// Log: log.Printf("Failed to marshal event: %v", err)
		return nil // Don’t fail the operation
	}
	u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON))

	return nil
}

func (u *ChatUsecase) SendVoiceMessage(token string, receiverID, groupID *int, filePath string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Verify sender exists (discard unused sender)
	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Create the message
	message := &domain.Message{
		SenderID:  senderID,
		Content:   filePath,
		IsVoice:   true,
		CreatedAt: time.Now(),
	}

	// Handle 1:1 or group context
	if groupID != nil {
		// Group message
		group, err := u.groupRepo.FindByID(*groupID)
		if err != nil {
			return fmt.Errorf("group not found: %w", err)
		}
		inGroup := false
		for _, member := range group.Members {
			if member.ID == senderID {
				inGroup = true
				break
			}
		}
		if !inGroup {
			return errors.New("sender not in group")
		}
		if !group.Permission["can_send"] {
			return errors.New("group does not allow sending messages")
		}
		message.GroupID = groupID
	} else if receiverID != nil {
		// Personal message
		if _, err := u.userRepo.FindByID(*receiverID); err != nil {
			return fmt.Errorf("receiver not found: %w", err)
		}
		message.ReceiverID = receiverID
	} else {
		return errors.New("either receiverID or groupID must be provided")
	}

	// Save the message
	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save voice message: %w", err)
	}

	// Broadcast or queue
	event := map[string]interface{}{
		"type":       "voice_message",
		"message_id": message.ID,
		"sender_id":  senderID,
		"content":    filePath,
		"created_at": message.CreatedAt,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		// Log: log.Printf("Failed to marshal event: %v", err)
		return nil // Don’t fail the operation
	}

	if groupID != nil {
		event["group_id"] = *groupID
		u.redisService.PublishMessage("group:"+strconv.Itoa(*groupID), string(eventJSON))
	} else {
		receiverStatus, err := u.redisService.GetUserStatus(*receiverID)
		if err != nil {
			receiverStatus = "offline" // Default if Redis fails
		}
		event["receiver_id"] = *receiverID
		eventJSON, _ = json.Marshal(event) // Re-marshal with receiver_id
		if receiverStatus == "online" {
			u.redisService.PublishMessage("user:"+strconv.Itoa(*receiverID), string(eventJSON))
		} else {
			u.redisService.PublishMessage("user:queue:"+strconv.Itoa(*receiverID), string(eventJSON))
		}
	}

	return nil
}
