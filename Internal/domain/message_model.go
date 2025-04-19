// package domain

// import (
// 	"time"
// )

//	type Message struct {
//		ID         int        `gorm:"primaryKey;autoIncrement"`
//		SenderID   int        `gorm:"not null"` // References User.ID
//		ReceiverID *int       // Nullable for group messages, references User.ID
//		GroupID    *int       // Nullable for 1:1 messages, references Group.ID
//		Content    string     `gorm:"type:text;not null"`
//		IsVoice    bool       `gorm:"default:false"`
//		CreatedAt  time.Time  `gorm:"not null"`
//		ReadAt     *time.Time // Nullable for read receipts
//	}
package domain

import (
	"time"
)

type Message struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderID   int       `json:"sender_id"`
	ReceiverID *int      `json:"receiver_id,omitempty"`
	GroupID    *int      `json:"group_id,omitempty"`
	Content    string    `json:"content"`
	Filename   *string   `json:"filename,omitempty"`
	Filetype   *string   `json:"filetype,omitempty"`
	IsFile     bool      `json:"is_file"`
	IsVoice    bool      `json:"is_voice"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status;default:'sent'"` // "sent", "delivered", "seen"
	ReplyTo    *int      `json:"reply_to,omitempty"`
}
