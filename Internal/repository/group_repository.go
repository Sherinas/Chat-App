package repository

import (
	"fmt"
	"log"
	"strings"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(group *domain.Group) error
	FindByID(id int) (*domain.Group, error)
	Update(group *domain.Group) error
	Delete(id int) error
	GetAll() ([]*domain.Group, error)
	GetAllGroupsByUserID(userID int) ([]*domain.Group, error)
	AddMember(groupID, userID int) error
	RemoveMember(groupID, userID int) error
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {

	return &groupRepository{db: db}

}

func (r *groupRepository) Create(group *domain.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) GetAll() ([]*domain.Group, error) {
	var groups []*domain.Group
	if err := r.db.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}
func (r *groupRepository) FindByID(id int) (*domain.Group, error) {
	var group domain.Group
	log.Println("................gpid", id)
	err := r.db.Preload("Members").First(&group, id).Error

	log.Println("errorrrrrrrrrrrrr", err)
	if err != nil {

		return nil, err
	}

	return &group, nil
}

func (r *groupRepository) Update(group *domain.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) Delete(id int) error {
	return r.db.Delete(&domain.Group{}, id).Error
}

func (r *groupRepository) AddMember(groupID, userID int) error {
	return r.db.Exec("INSERT INTO user_groups (group_id, user_id) VALUES (?, ?)", groupID, userID).Error
}

func (r *groupRepository) RemoveMember(groupID, userID int) error {
	err := r.db.Exec("DELETE FROM user_groups WHERE group_id = ? AND user_id = ?", groupID, userID).Error
	if err != nil {
		fmt.Println("Error removing user from group:", err)
	}
	return err
	// return r.db.Exec("DELETE FROM user_groups WHERE group_id = ? AND user_id = ?", groupID, userID).Error
}

// func (r *groupRepository) GetAllGroupsByUserID(userID int) ([]*domain.Group, error) {
// 	log.Println("running")
// 	var groups []*domain.Group
// 	if err := r.db.Joins("JOIN group_members ON group_members.group_id = groups.id").
// 		Where("group_members.user_id = ?", userID).
// 		Find(&groups).Error; err != nil {
// 		return nil, err
// 	}
// 	return groups, nil
// }

func (r *groupRepository) GetAllGroupsByUserID(userID int) ([]*domain.Group, error) {
	log.Println("running")
	var groups []*domain.Group
	err := r.db.Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		// Handle missing table gracefully
		if strings.Contains(err.Error(), "user_groups does not exist") ||
			strings.Contains(err.Error(), "SQLSTATE 42P01") ||
			strings.Contains(err.Error(), "no such table") {
			log.Printf("Table 'user_groups' not found, returning empty groups: %v", err)
			return []*domain.Group{}, nil
		}
		log.Printf("Failed to fetch groups for userID %d: %v", userID, err)
		return nil, err
	}
	return groups, nil
}
