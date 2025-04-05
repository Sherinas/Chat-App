package domain

import (
	"time"
)

type User struct {
	ID          int       `gorm:"primaryKey;autoIncrement"`
	EmployeeID  string    `gorm:"type:varchar(50);ubique;not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Email       string    `gorm:"type:varchar(255);unique;not null"`
	Password    string    `gorm:"type:varchar(255)"`
	Designation string    `json:"designation"`
	Role        string    `gorm:"varchar(20);not null"`
	State       string    `gorm:"type:varchar(20);default:offline"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdateAt    time.Time `gorm:"autoUpdateTime"`
	Groups      []Group   `gorm:"many2many:user_groups"`
}
