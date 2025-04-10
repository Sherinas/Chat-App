package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// PermissionMap wraps map[string]bool for JSON handling
type PermissionMap map[string]bool

// Value implements driver.Valuer to serialize to JSON
func (p PermissionMap) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan implements sql.Scanner to deserialize from JSON
func (p *PermissionMap) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("unsupported scan type for PermissionMap")
	}
}

const AdminGroupID = 6

type Group struct {
	ID          int           `gorm:"primaryKey;autoIncrement"`
	Name        string        `gorm:"type:varchar(255);not null"`
	CreatedBy   int           `gorm:"not null"` // ref user.id
	Permission  PermissionMap `gorm:"type:json;default:'{\"can_send\":true}'"`
	CreatedAt   time.Time
	Members     []User `gorm:"many2many:user_groups;"`
	MemberCount int    `json:"memberCount" gorm:"column:member_count"`
}

type UserGroup struct {
	GroupID int `gorm:"column:group_id;not null;index"` // Matches user_groups.group_id
	UserID  int `gorm:"column:user_id;not null;index"`  // Matches user_groups.user_id
}
