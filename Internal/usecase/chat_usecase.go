package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func (u *ChatUsecase) SendGroupMessage(token string, groupID int, content string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

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

	message := &domain.Message{
		SenderID:  senderID,
		GroupID:   &groupID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	event := map[string]interface{}{
		"type":       "text",
		"targetType": "group",
		"message_id": message.ID,
		"group_id":   groupID,
		"sender_id":  senderID,
		"content":    content,
		"created_at": message.CreatedAt,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if err := u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON)); err != nil {
		log.Printf("Failed to publish message to %s: %v", "group:"+strconv.Itoa(groupID), err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// func (u *ChatUsecase) SendGroupMessage(token string, groupID int, content string) error {
// 	senderID, _, err := u.ValidateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}

// 	if _, err := u.userRepo.FindByID(senderID); err != nil {
// 		return fmt.Errorf("sender not found: %w", err)
// 	}

// 	group, err := u.groupRepo.FindByID(groupID)
// 	if err != nil {
// 		return fmt.Errorf("group not found: %w", err)
// 	}
// 	inGroup := false
// 	for _, member := range group.Members {
// 		if member.ID == senderID {
// 			inGroup = true
// 			break
// 		}
// 	}
// 	if !inGroup {
// 		return errors.New("sender not in group")
// 	}
// 	if !group.Permission["can_send"] {
// 		return errors.New("group does not allow sending messages")
// 	}

// 	message := &domain.Message{
// 		SenderID:  senderID,
// 		GroupID:   &groupID,
// 		Content:   content,
// 		CreatedAt: time.Now(),
// 	}
// 	if err := u.messageRepo.Create(message); err != nil {
// 		return fmt.Errorf("failed to save message: %w", err)
// 	}

// 	event := map[string]interface{}{
// 		"type":       "text",
// 		"targetType": "group", // Explicitly indicate group message
// 		"message_id": message.ID,
// 		"group_id":   groupID, // Group ID
// 		"sender_id":  senderID,
// 		"content":    content,
// 		"created_at": message.CreatedAt,
// 	}
// 	eventJSON, err := json.Marshal(event)
// 	if err != nil {
// 		log.Printf("Failed to marshal event: %v", err)
// 		return nil
// 	}
// 	u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON))

// 	return nil
// }

// func (u *ChatUsecase) SendPersonalMessage(token string, receiverID int, content string) error {
// 	senderID, _, err := u.ValidateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}

// 	if _, err := u.userRepo.FindByID(senderID); err != nil {
// 		return fmt.Errorf("sender not found: %w", err)
// 	}
// 	if _, err := u.userRepo.FindByID(receiverID); err != nil {
// 		return fmt.Errorf("receiver not found: %w", err)
// 	}

// 	message := &domain.Message{
// 		SenderID:   senderID,
// 		ReceiverID: &receiverID,
// 		Content:    content,
// 		CreatedAt:  time.Now(),
// 	}
// 	if err := u.messageRepo.Create(message); err != nil {
// 		return fmt.Errorf("failed to save message: %w", err)
// 	}

// 	receiverStatus, err := u.redisService.GetUserStatus(receiverID)
// 	if err != nil {
// 		receiverStatus = "offline"
// 	}

// 	event := map[string]interface{}{
// 		"type":        "text",
// 		"targetType":  "user", // Explicitly indicate personal message
// 		"sender_id":   senderID,
// 		"receiver_id": receiverID, // User ID
// 		"content":     content,
// 		"created_at":  message.CreatedAt,
// 	}
// 	eventJSON, err := json.Marshal(event)
// 	if err != nil {
// 		log.Printf("Failed to marshal event: %v", err)
// 		return nil
// 	}

// 	if receiverStatus == "online" {
// 		u.redisService.PublishMessage("user:"+strconv.Itoa(receiverID), string(eventJSON))
// 	} else {
// 		u.redisService.PublishMessage("user:queue:"+strconv.Itoa(receiverID), string(eventJSON))
// 	}

// 	return nil
// }

// func (u *ChatUsecase) SendPersonalMessage(token string, receiverID int, content string, msgType string) error {
// 	senderID, _, err := u.ValidateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}

// 	// Verify sender and receiver exist (discard unused variables)
// 	if _, err := u.userRepo.FindByID(senderID); err != nil {
// 		return fmt.Errorf("sender not found: %w", err)
// 	}
// 	if _, err := u.userRepo.FindByID(receiverID); err != nil {
// 		return fmt.Errorf("receiver not found: %w", err)
// 	}

// 	// Create and store the message
// 	message := &domain.Message{

// 		SenderID:   senderID,
// 		ReceiverID: &receiverID,
// 		Content:    content,
// 		//IsVoice:    false,
// 		CreatedAt: time.Now(),
// 	}
// 	if err := u.messageRepo.Create(message); err != nil {
// 		return fmt.Errorf("failed to save message: %w", err)
// 	}

// 	// Broadcast or queue based on receiver status
// 	receiverStatus, err := u.redisService.GetUserStatus(receiverID)
// 	if err != nil {
// 		receiverStatus = "offline" // Default to offline if Redis fails
// 	}

// 	event := map[string]interface{}{
// 		"type": msgType,
// 		//"message_id":  message.ID,
// 		"sender_id":   senderID,
// 		"receiver_id": receiverID,
// 		"content":     content,
// 		"created_at":  message.CreatedAt,
// 	}
// 	eventJSON, err := json.Marshal(event)
// 	if err != nil {
// 		// Log: log.Printf("Failed to marshal event: %v", err)
// 		return nil // Don’t fail the operation
// 	}

// 	if receiverStatus == "online" {
// 		u.redisService.PublishMessage("user:"+strconv.Itoa(receiverID), string(eventJSON))
// 	} else {
// 		u.redisService.PublishMessage("user:queue:"+strconv.Itoa(receiverID), string(eventJSON))
// 	}

// 	return nil
// }

// func (u *ChatUsecase) SendGroupMessage(token string, groupID int, content string) error {
// 	senderID, _, err := u.ValidateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}

// 	// Verify sender exists (discard unused sender)
// 	if _, err := u.userRepo.FindByID(senderID); err != nil {
// 		return fmt.Errorf("sender not found: %w", err)
// 	}

// 	// Verify group exists and sender has permission
// 	group, err := u.groupRepo.FindByID(groupID)
// 	if err != nil {
// 		return fmt.Errorf("group not found: %w", err)
// 	}
// 	inGroup := false
// 	for _, member := range group.Members {
// 		if member.ID == senderID {
// 			inGroup = true
// 			break
// 		}
// 	}
// 	if !inGroup {
// 		return errors.New("sender not in group")
// 	}
// 	if !group.Permission["can_send"] {
// 		return errors.New("group does not allow sending messages")
// 	}

// 	// Create and store the message
// 	message := &domain.Message{
// 		SenderID:  senderID,
// 		GroupID:   &groupID,
// 		Content:   content,
// 		IsVoice:   false,
// 		CreatedAt: time.Now(),
// 	}
// 	if err := u.messageRepo.Create(message); err != nil {
// 		return fmt.Errorf("failed to save message: %w", err)
// 	}

// 	// Broadcast to group channel
// 	event := map[string]interface{}{
// 		"type":       "group_message",
// 		"message_id": message.ID,
// 		"group_id":   groupID,
// 		"sender_id":  senderID,
// 		"content":    content,
// 		"created_at": message.CreatedAt,
// 	}
// 	eventJSON, err := json.Marshal(event)
// 	if err != nil {
// 		// Log: log.Printf("Failed to marshal event: %v", err)
// 		return nil // Don’t fail the operation
// 	}
// 	u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON))

//		return nil
//	}
func (u *ChatUsecase) SendPersonalMessage(token string, receiverID int, content string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}
	if _, err := u.userRepo.FindByID(receiverID); err != nil {
		return fmt.Errorf("receiver not found: %w", err)
	}

	message := &domain.Message{
		SenderID:   senderID,
		ReceiverID: &receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	receiverStatus, err := u.redisService.GetUserStatus(receiverID)
	if err != nil {
		receiverStatus = "offline"
	}

	event := map[string]interface{}{
		"type":        "text",
		"targetType":  "user", // Explicitly indicate personal message
		"sender_id":   senderID,
		"receiver_id": receiverID, // User ID
		"content":     content,
		"created_at":  message.CreatedAt,
	}

	log.Println(event)
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return nil
	}

	if receiverStatus == "online" {
		u.redisService.PublishMessage("user:"+strconv.Itoa(receiverID), string(eventJSON))
	} else {
		u.redisService.PublishMessage("user:queue:"+strconv.Itoa(receiverID), string(eventJSON))
	}

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

func (u *ChatUsecase) SendMultimediaMessage(token string, receiverID, groupID int, content, filename, filetype, msgType string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if senderID == 0 {
		return fmt.Errorf("invalid sender ID: %d", senderID)
	}

	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	message := &domain.Message{
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	isFile := msgType == "file_message"
	if isFile {
		message.IsFile = true
		message.Filename = &filename
		message.Filetype = &filetype
	} else {
		message.IsVoice = true
	}

	if groupID != 0 {
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
		message.GroupID = &groupID
	} else if receiverID != 0 {
		if _, err := u.userRepo.FindByID(receiverID); err != nil {
			return fmt.Errorf("receiver not found: %w", err)
		}
		message.ReceiverID = &receiverID
	} else {
		return errors.New("either receiverID or groupID must be provided")
	}

	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	event := map[string]interface{}{
		"type":       msgType,
		"message_id": message.ID,
		"sender_id":  senderID,
		"content":    content,
		"filename":   filename,
		"filetype":   filetype,
		"created_at": message.CreatedAt,
		"status":     "sent",
	}
	var targetType string
	if groupID != 0 {
		targetType = "group"
		event["group_id"] = groupID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		if err := u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON)); err != nil {
			log.Printf("Failed to publish to group %d: %v", groupID, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	} else {
		targetType = "user"
		event["receiver_id"] = receiverID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		receiverStatus, err := u.redisService.GetUserStatus(receiverID)
		if err != nil {
			receiverStatus = "offline"
		}
		var channel string
		if receiverStatus == "online" {
			channel = "user:" + strconv.Itoa(receiverID)
		} else {
			channel = "user:queue:" + strconv.Itoa(receiverID)
		}
		if err := u.redisService.PublishMessage(channel, string(eventJSON)); err != nil {
			log.Printf("Failed to publish to %s: %v", channel, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	}
	event["targetType"] = targetType
	return nil
}

func (u *ChatUsecase) GetMessageHistory(chatType string, chatID int) ([]domain.Message, error) {
	return u.messageRepo.FindHistory(chatType, chatID)
}

func (u *ChatUsecase) GetUnreadMessages(userID int) ([]domain.Message, error) {
	return u.messageRepo.FindUnread(userID)
}

func (u *ChatUsecase) MarkMessagesAsSeen(chatType string, chatID, userID int) error {
	messages, err := u.messageRepo.FindHistory(chatType, chatID)
	if err != nil {
		return err
	}
	for _, msg := range messages {
		if msg.ReceiverID != nil && *msg.ReceiverID == userID && msg.Status != "seen" {
			if err := u.messageRepo.UpdateStatus(msg.ID, "seen"); err != nil {
				return err
			}
			event := map[string]interface{}{
				"type":       "status_update",
				"message_id": msg.ID,
				"status":     "seen",
				"sender_id":  userID,
			}
			if chatType == "user" {
				event["receiver_id"] = *msg.ReceiverID
			} else {
				event["group_id"] = chatID
			}
			eventJSON, _ := json.Marshal(event)
			u.redisService.PublishMessage("status:"+strconv.Itoa(userID), string(eventJSON))
		}
	}
	return nil
}

func (u *ChatUsecase) UpdateUserStatus(userID int, status string) error {
	if err := u.userRepo.UpdateStatus(userID, status); err != nil {
		return err
	}
	u.redisService.SetUserStatus(userID, status)
	return nil
}

func (u *ChatUsecase) SendReplyMessage(token string, receiverID, groupID int, content string, replyTo int) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if senderID == 0 {
		return fmt.Errorf("invalid sender ID: %d", senderID)
	}

	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Validate the replyTo message exists
	//originalMsg := &domain.Message{}
	if _, err := u.messageRepo.FindByID(replyTo); err != nil {
		return fmt.Errorf("original message not found: %w", err)
	}

	message := &domain.Message{
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now(),
		ReplyTo:   &replyTo,
	}

	if groupID != 0 {
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
		message.GroupID = &groupID
	} else if receiverID != 0 {
		if _, err := u.userRepo.FindByID(receiverID); err != nil {
			return fmt.Errorf("receiver not found: %w", err)
		}
		message.ReceiverID = &receiverID
	} else {
		return errors.New("either receiverID or groupID must be provided")
	}

	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save reply message: %w", err)
	}

	event := map[string]interface{}{
		"type":       "reply",
		"message_id": message.ID,
		"sender_id":  senderID,
		"content":    content,
		"reply_to":   replyTo,
		"created_at": message.CreatedAt,
		"status":     "sent",
	}
	var targetType string
	if groupID != 0 {
		targetType = "group"
		event["group_id"] = groupID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal reply event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		if err := u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON)); err != nil {
			log.Printf("Failed to publish reply to group %d: %v", groupID, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	} else {
		targetType = "user"
		event["receiver_id"] = receiverID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal reply event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		receiverStatus, err := u.redisService.GetUserStatus(receiverID)
		if err != nil {
			receiverStatus = "offline"
		}
		var channel string
		if receiverStatus == "online" {
			channel = "user:" + strconv.Itoa(receiverID)
		} else {
			channel = "user:queue:" + strconv.Itoa(receiverID)
		}
		if err := u.redisService.PublishMessage(channel, string(eventJSON)); err != nil {
			log.Printf("Failed to publish reply to %s: %v", channel, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	}
	event["targetType"] = targetType
	return nil
}

func (u *ChatUsecase) ForwardMessage(token string, targetID int, content string) error {
	senderID, _, err := u.ValidateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if senderID == 0 {
		return fmt.Errorf("invalid sender ID: %d", senderID)
	}

	if _, err := u.userRepo.FindByID(senderID); err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Determine if target is a user or group
	var receiverID, groupID int
	var isGroup bool
	if group, err := u.groupRepo.FindByID(targetID); err == nil {
		isGroup = true
		groupID = targetID
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
	} else {
		isGroup = false
		receiverID = targetID
		if _, err := u.userRepo.FindByID(receiverID); err != nil {
			return fmt.Errorf("receiver not found: %w", err)
		}
	}

	message := &domain.Message{
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if isGroup {
		message.GroupID = &groupID
	} else {
		message.ReceiverID = &receiverID
	}

	if err := u.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to save forwarded message: %w", err)
	}

	event := map[string]interface{}{
		"type":       "forward",
		"message_id": message.ID,
		"sender_id":  senderID,
		"content":    content,
		"created_at": message.CreatedAt,
		"status":     "sent",
	}
	var targetType string
	if isGroup {
		targetType = "group"
		event["group_id"] = groupID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal forward event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		if err := u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON)); err != nil {
			log.Printf("Failed to publish forward to group %d: %v", groupID, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	} else {
		targetType = "user"
		event["receiver_id"] = receiverID
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal forward event: %v", err)
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		receiverStatus, err := u.redisService.GetUserStatus(receiverID)
		if err != nil {
			receiverStatus = "offline"
		}
		var channel string
		if receiverStatus == "online" {
			channel = "user:" + strconv.Itoa(receiverID)
		} else {
			channel = "user:queue:" + strconv.Itoa(receiverID)
		}
		if err := u.redisService.PublishMessage(channel, string(eventJSON)); err != nil {
			log.Printf("Failed to publish forward to %s: %v", channel, err)
			return fmt.Errorf("failed to publish: %w", err)
		}
	}
	event["targetType"] = targetType
	return nil
}
