package domain

import "time"

type GroupRequest struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `gorm:"not null"` // References User.ID
	GroupID   int       `gorm:"not null"` // References Group.ID
	Status    string    `gorm:"type:varchar(20);default:pending"`
	CreatedAt time.Time `gorm:"not null"`
}
