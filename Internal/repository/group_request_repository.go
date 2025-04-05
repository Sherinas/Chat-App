package repository

import (
	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type GroupRequestRepository interface {
	Create(request *domain.GroupRequest) error
	FindByID(id int) (*domain.GroupRequest, error)
	FindPendingByUserID(userID int) ([]domain.GroupRequest, error)
	Update(request *domain.GroupRequest) error

	Delete(id int) error
}

type groupRequestRepository struct {
	db *gorm.DB
}

func NewGroupRequestRepository(db *gorm.DB) GroupRequestRepository {
	return &groupRequestRepository{db: db}
}

func (r *groupRequestRepository) Create(request *domain.GroupRequest) error {
	return r.db.Create(request).Error
}

func (r *groupRequestRepository) FindByID(id int) (*domain.GroupRequest, error) {
	var request domain.GroupRequest
	err := r.db.First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *groupRequestRepository) FindPendingByUserID(userID int) ([]domain.GroupRequest, error) {
	var requests []domain.GroupRequest
	err := r.db.Where("user_id = ? AND status = ?", userID, "pending").Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *groupRequestRepository) Update(request *domain.GroupRequest) error {
	return r.db.Save(request).Error
}

func (r *groupRequestRepository) Delete(id int) error {
	return r.db.Delete(&domain.GroupRequest{}, id).Error
}
