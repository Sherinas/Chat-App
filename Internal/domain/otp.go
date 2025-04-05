package domain

import "time"

type OTPVerification struct {
	ID         int       `gorm:"primaryKey;autoIncrement"`
	EmployeeID string    `gorm:"type:varchar(50);not null"`
	OTP        string    `gorm:"type:varchar(6);not null"`
	ExpiresAt  time.Time `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null"`
}
