package repository

import (
	"fmt"
	"log"
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
	FindHistoryBetweenUsers(userID, receiverID int) ([]domain.Message, error)
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

// func (r *messageRepository) FindHistory(chatType string, chatID int) ([]domain.Message, error) {
// 	var messages []domain.Message

// 	log.Println("chat details", chatID, chatType)
// 	query := r.db.Order("created_at DESC")
// 	if chatType == "user" {
// 		query = query.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
// 			chatID, r.db.Table("users").Select("id").Where("id = ?", chatID),
// 			r.db.Table("users").Select("id").Where("id = ?", chatID), chatID)
// 	} else {
// 		query = query.Where("group_id = ?", chatID)
// 	}
// 	if err := query.Find(&messages).Error; err != nil {
// 		return nil, err
// 	}
// 	return messages, nil
// }

func (r *messageRepository) FindUnread(userID int) ([]domain.Message, error) {
	var messages []domain.Message
	if err := r.db.Where("receiver_id = ? AND status = ?", userID, "sent").Or("receiver_id = ? AND status = ?", userID, "delivered").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) UpdateStatus(messageID int, status string) error {
	log.Println("test1\\\\\\\\\\\\\\\\\\\\", messageID, status)
	return r.db.Model(&domain.Message{}).Where("id = ?", messageID).Update("status", status).Error
}

// FindByID retrieves a message by its ID and populates the provided message struct
func (r *messageRepository) FindHistoryBetweenUsers(userID, receiverID int) ([]domain.Message, error) {
	var messages []domain.Message
	query := `SELECT * FROM "messages" WHERE 
              (sender_id = $1 AND receiver_id = $2) 
              OR (sender_id = $2 AND receiver_id = $1) 
              AND group_id IS NULL 
              ORDER BY created_at DESC`
	if err := r.db.Raw(query, userID, receiverID).Scan(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) FindHistory(chatType string, chatID int) ([]domain.Message, error) {
	var messages []domain.Message
	log.Println("chat details", chatID, chatType)
	if chatType == "user" {
		query := `SELECT * FROM "messages" WHERE 
                  (sender_id = $1 OR receiver_id = $1) 
                  AND group_id IS NULL 
                  ORDER BY created_at DESC`
		if err := r.db.Raw(query, chatID).Scan(&messages).Error; err != nil {
			return nil, err
		}
	} else if chatType == "group" {
		query := `SELECT * FROM "messages" WHERE group_id = $1 ORDER BY created_at DESC`
		if err := r.db.Raw(query, chatID).Scan(&messages).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid chat type: %s", chatType)
	}
	return messages, nil
}
