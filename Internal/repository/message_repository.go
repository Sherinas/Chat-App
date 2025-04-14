package repository

import (
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *domain.Message) error
	FindByID(id int) (*domain.Message, error)
	FindByUserID(userID int) ([]domain.Message, error)
	FindByGroupID(groupID int) ([]domain.Message, error)
	UpdateReadAt(messageID int) error
	UpdateStatus(messageID int, status string) error
	FindUnread(userID int) ([]domain.Message, error)
	FindHistory(chatType string, chatID int) ([]domain.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *domain.Message) error {
	return r.db.Create(message).Error

}

func (r *messageRepository) FindByID(id int) (*domain.Message, error) {
	var message domain.Message

	err := r.db.First(&message, id).Error

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *messageRepository) FindByUserID(userID int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.Where("sender_id = ? OR receiver_id = ?", userID, userID).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) FindByGroupID(groupID int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.Where("group_id = ?", groupID).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) UpdateReadAt(messageID int) error {
	return r.db.Model(&domain.Message{}).Where("id = ?", messageID).Update("read_at", time.Now()).Error
}

func (r *messageRepository) FindHistory(chatType string, chatID int) ([]domain.Message, error) {
	var messages []domain.Message
	query := r.db.Order("created_at DESC")
	if chatType == "user" {
		query = query.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			chatID, r.db.Table("users").Select("id").Where("id = ?", chatID),
			r.db.Table("users").Select("id").Where("id = ?", chatID), chatID)
	} else {
		query = query.Where("group_id = ?", chatID)
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) FindUnread(userID int) ([]domain.Message, error) {
	var messages []domain.Message
	if err := r.db.Where("receiver_id = ? AND status = ?", userID, "sent").Or("receiver_id = ? AND status = ?", userID, "delivered").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) UpdateStatus(messageID int, status string) error {
	return r.db.Model(&domain.Message{}).Where("id = ?", messageID).Update("status", status).Error
}

// FindByID retrieves a message by its ID and populates the provided message struct
