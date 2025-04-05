package repository

import (
	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id int) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindByEmployeeID(employeeID string) (*domain.User, error)
	GetAllUsers() ([]domain.User, error)
	Update(user *domain.User) error
	Delete(id int) error
	AddToGroup(userID, groupID int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllUsers() ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
func (r *userRepository) Create(user *domain.User) error {

	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id int) (*domain.User, error) {

	var user domain.User

	err := r.db.Preload("Groups").First(&user, id).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmployeeID(employeeID string) (*domain.User, error) {
	var user domain.User

	err := r.db.Where("employee_id = ?", employeeID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id int) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *userRepository) AddToGroup(userID, groupID int) error {
	return r.db.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", userID, groupID).Error
}
