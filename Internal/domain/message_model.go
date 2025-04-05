package domain

import (
	"time"
)

type Message struct {
	ID         int        `gorm:"primaryKey;autoIncrement"`
	SenderID   int        `gorm:"not null"` // References User.ID
	ReceiverID *int       // Nullable for group messages, references User.ID
	GroupID    *int       // Nullable for 1:1 messages, references Group.ID
	Content    string     `gorm:"type:text;not null"`
	IsVoice    bool       `gorm:"default:false"`
	CreatedAt  time.Time  `gorm:"not null"`
	ReadAt     *time.Time // Nullable for read receipts
}
