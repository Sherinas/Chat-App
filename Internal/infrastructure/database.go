package database

import (
	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	dsn := "host=localhost user=sherinascdlm password=admin123 dbname=chatapp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate tables based on domain entities
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Group{},
		&domain.Message{},
		&domain.GroupRequest{},
		&domain.OTPVerification{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
