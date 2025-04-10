package repository

import (
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
	GetAllWithMembership(userID int) ([]*domain.Group, error)
	GetAllGroupsByUserID(userID int) ([]*domain.Group, error)
	AddMember(userID, groupID int) error
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

	err := r.db.Preload("Members").First(&group, id).Error

	if err != nil {

		return nil, err
	}

	return &group, nil
}

func (r *groupRepository) Update(group *domain.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) AddMember(userID, groupID int) error {
	log.Printf("Repository: Adding userID=%d to groupID=%d", userID, groupID)
	return r.db.Model(&domain.User{ID: userID}).Association("Groups").Append(&domain.Group{ID: groupID})
}

func (r *groupRepository) RemoveMember(userID, groupID int) error {
	return r.db.Model(&domain.User{ID: userID}).Association("Groups").Delete(&domain.Group{ID: groupID})
}

//	func (r *groupRepository) GetAllGroupsByUserID(userID int) ([]*domain.Group, error) {
//		log.Println("running")
//		var groups []*domain.Group
//		err := r.db.Joins("JOIN user_groups ON user_groups.group_id = groups.id").
//			Where("user_groups.user_id = ?", userID).
//			Find(&groups).Error
//		if err != nil {
//			// Handle missing table gracefully
//			if strings.Contains(err.Error(), "user_groups does not exist") ||
//				strings.Contains(err.Error(), "SQLSTATE 42P01") ||
//				strings.Contains(err.Error(), "no such table") {
//				log.Printf("Table 'user_groups' not found, returning empty groups: %v", err)
//				return []*domain.Group{}, nil
//			}
//			log.Printf("Failed to fetch groups for userID %d: %v", userID, err)
//			return nil, err
//		}
//		return groups, nil
//	}
// func (r *groupRepository) GetAllGroupsByUserID(userID int) ([]*domain.Group, error) {
// 	log.Println("running GetAllGroupsByUserID for userID:", userID)

// 	var groups []*domain.Group

// 	// Raw SQL query to handle JSON field and member count
// 	sql := `
//         SELECT g.*, COUNT(ug.user_id) as member_count
//         FROM groups g
//         LEFT JOIN user_groups ug ON ug.group_id = g.id
//         INNER JOIN (
//             SELECT group_id
//             FROM user_groups
//             WHERE user_id = ?
//         ) uj ON uj.group_id = g.id
//         GROUP BY g.id
//     `
// 	err := r.db.Raw(sql, userID).Scan(&groups).Error

// 	if err != nil {
// 		// Handle missing table gracefully
// 		if strings.Contains(err.Error(), "user_groups does not exist") ||
// 			strings.Contains(err.Error(), "SQLSTATE 42P01") ||
// 			strings.Contains(err.Error(), "no such table") {
// 			log.Printf("Table 'user_groups' not found, returning empty groups: %v", err)
// 			return []*domain.Group{}, nil
// 		}
// 		log.Printf("Failed to fetch groups for userID %d: %v", userID, err)
// 		return nil, err
// 	}

// 	// Log the fetched groups for debugging
// 	log.Println("Fetched groups with member counts:", groups)

// 	return groups, nil
// }

func (r *groupRepository) GetAllGroupsByUserID(userID int) ([]*domain.Group, error) {
	log.Println("running GetAllGroupsByUserID for userID:", userID)

	var groups []*domain.Group
	err := r.db.Model(&domain.Group{}).
		Joins("INNER JOIN user_groups ug ON ug.group_id = groups.id").
		Joins("INNER JOIN (SELECT group_id FROM user_groups WHERE user_id = ?) uj ON uj.group_id = groups.id", userID).
		Group("groups.id").
		Select("groups.*, COUNT(ug.user_id) as member_count").
		Scan(&groups).Error

	if err != nil {
		if strings.Contains(err.Error(), "user_groups does not exist") ||
			strings.Contains(err.Error(), "SQLSTATE 42P01") ||
			strings.Contains(err.Error(), "no such table") {
			log.Printf("Table 'user_groups' not found, returning empty groups: %v", err)
			return []*domain.Group{}, nil
		}
		log.Printf("Failed to fetch groups for userID %d: %v", userID, err)
		return nil, err
	}

	for _, g := range groups {
		log.Printf("Fetched group: ID=%d, Name=%s, MemberCount=%d", g.ID, g.Name, g.MemberCount)
	}

	return groups, nil
}
func (r *groupRepository) GetAllWithMembership(userID int) ([]*domain.Group, error) {
	var groups []*domain.Group
	err := r.db.Model(&domain.Group{}).
		Joins("LEFT JOIN user_groups ON user_groups.group_id = groups.id AND user_groups.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}

	// Add isJoined based on the join result (simplified; GORM doesn't directly set this)
	for _, group := range groups {
		group.Members = nil // Clear to avoid circular references
		var count int64
		r.db.Model(&domain.User{}).
			Where("id = ? AND EXISTS (SELECT 1 FROM user_groups WHERE user_id = ? AND group_id = ?)", userID, userID, group.ID).
			Count(&count)
		if group.Permission == nil {
			group.Permission = make(domain.PermissionMap)
		}
		group.Permission["isJoined"] = count > 0
	}

	return groups, nil
}

func (r *groupRepository) Delete(groupID int) error {
	log.Printf("Repository: Deleting groupID=%d", groupID)
	var group domain.Group
	if err := r.db.First(&group, groupID).Error; err != nil {
		return err
	}

	if err := r.db.Model(&domain.UserGroup{}).Where("group_id = ?", groupID).Delete(&domain.UserGroup{}).Error; err != nil {
		return err
	}

	return r.db.Delete(&group).Error
}
