package repository

import (
	"log"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id int) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindByEmployeeID(employeeID string) (*domain.User, error)
	GetAllUsers() ([]domain.User, error)
	Update(userID int, user *domain.User) error
	Delete(id int) error
	AddToGroup(userID, groupID int) error
	CountAdmins() (int64, error)
	UpdateRole(id int, role string) error
	UpdateStatus(id int, status string) error
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

func (r *userRepository) AddToGroup(userID, groupID int) error {
	return r.db.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", userID, groupID).Error
}

func (r *userRepository) CountAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("role = ?", "admin").Count(&count).Error
	return count, err
}
func (r *userRepository) UpdateRole(id int, role string) error {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return err
	}
	user.Role = role
	return r.db.Save(&user).Error
}

func (r *userRepository) Delete(userID int) error {
	log.Printf("Repository: Deactivating userID=%d", userID)
	var user domain.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	//Soft delete the user by setting DeletedAt
	now := time.Now()
	user.DeletedAt = &now
	return r.db.Delete(&user).Error
}

func (r *userRepository) Update(userID int, user *domain.User) error {
	log.Printf("Repository: Updating userID=%d", userID)
	var existingUser domain.User
	if err := r.db.Where("deleted_at IS NULL").First(&existingUser, userID).Error; err != nil {
		return err
	}

	// Update only the fields provided in the user object
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Password != "" {
		existingUser.Password = user.Password
	}
	if user.Role != "" {
		existingUser.Role = user.Role
	}

	return r.db.Save(&existingUser).Error
}
func (r *userRepository) UpdateStatus(id int, status string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("state", status).Error
}
