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
