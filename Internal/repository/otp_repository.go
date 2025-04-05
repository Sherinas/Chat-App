package repository

import (
	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(otp *domain.OTPVerification) error
	FindByEmployeeID(employeeID string) (*domain.OTPVerification, error)
	Delete(id int) error
}
type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Create(otp *domain.OTPVerification) error {
	return r.db.Create(otp).Error
}

func (r *otpRepository) FindByEmployeeID(employeeID string) (*domain.OTPVerification, error) {
	var otp domain.OTPVerification
	err := r.db.Where("employee_id = ?", employeeID).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) Delete(id int) error {
	return r.db.Delete(&domain.OTPVerification{}, id).Error
}
