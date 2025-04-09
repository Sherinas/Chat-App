package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"github.com/Sherinas/Chat-App-Clean/Internal/repository"
)

type GroupUsecase struct {
	groupRepo        repository.GroupRepository
	userRepo         repository.UserRepository
	groupRequestRepo repository.GroupRequestRepository
	authService      AuthService
	redisService     RedisService // Use interface from usecase
}

func NewGroupUsecase(
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	groupRequestRepo repository.GroupRequestRepository,
	authService AuthService,
	redisService RedisService,
) *GroupUsecase {
	return &GroupUsecase{
		groupRepo:        groupRepo,
		userRepo:         userRepo,
		groupRequestRepo: groupRequestRepo,
		authService:      authService,
		redisService:     redisService,
	}
}

func (u *GroupUsecase) GetAllGroups() ([]*domain.Group, error) {
	groups, err := u.groupRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (u *GroupUsecase) CreateGroup(userID int, name string, permissions map[string]bool) (int, error) {
	log.Println("creategroup running")

	_, err := u.userRepo.FindByID(userID)
	if err != nil {
		return 0, fmt.Errorf("admin not found: %w", err)
	}

	// Create the group
	group := &domain.Group{
		Name:       name,
		CreatedBy:  userID,
		Permission: permissions,
		CreatedAt:  time.Now(),
	}
	if err := u.groupRepo.Create(group); err != nil {
		return 0, fmt.Errorf("group creation failed: %w", err)
	}

	// Publish event to Redis
	event := map[string]interface{}{
		"type":     "group_created",
		"group_id": group.ID,
		"name":     name,
	}
	eventJSON, _ := json.Marshal(event)
	u.redisService.PublishMessage("group_events", string(eventJSON)) // Fire-and-forget

	return group.ID, nil
}

func (u *GroupUsecase) RequestToJoinGroup(token string, groupID int) (int, error) {
	userID, _, err := u.validateTokenWithRedis(token)
	if err != nil {
		return 0, fmt.Errorf("token validation failed: %w", err)
	}
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}
	_, err = u.groupRepo.FindByID(groupID)
	if err != nil {
		return 0, fmt.Errorf("group not found: %w", err)
	}
	for _, g := range user.Groups {
		if g.ID == groupID {
			return 0, errors.New("user already in group")
		}
	}
	pendingRequests, err := u.groupRequestRepo.FindPendingByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to check pending requests: %w", err)
	}
	for _, req := range pendingRequests {
		if req.GroupID == groupID {
			return 0, errors.New("request already pending")
		}
	}
	request := &domain.GroupRequest{
		UserID:    userID,
		GroupID:   groupID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	if err := u.groupRequestRepo.Create(request); err != nil {
		return 0, fmt.Errorf("request creation failed: %w", err)
	}
	return request.ID, nil
}

func (u *GroupUsecase) ApproveGroupRequest(token string, requestID int) error {
	adminID, role, err := u.validateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	if role != "admin" {
		return errors.New("only admins can approve requests")
	}
	_, err = u.userRepo.FindByID(adminID)
	if err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}
	request, err := u.groupRequestRepo.FindByID(requestID)
	if err != nil {
		return fmt.Errorf("request not found: %w", err)
	}
	if request.Status != "pending" {
		return errors.New("request already processed")
	}
	request.Status = "approved"
	if err := u.groupRequestRepo.Update(request); err != nil {
		return fmt.Errorf("request update failed: %w", err)
	}
	if err := u.userRepo.AddToGroup(request.UserID, request.GroupID); err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	event := map[string]interface{}{
		"type":     "user_joined",
		"group_id": request.GroupID,
		"user_id":  request.UserID,
	}
	eventJSON, _ := json.Marshal(event)
	u.redisService.PublishMessage("group:"+strconv.Itoa(request.GroupID), string(eventJSON))

	return nil
}

func (u *GroupUsecase) RejectGroupRequest(token string, requestID int) error {
	adminID, role, err := u.validateTokenWithRedis(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	if role != "admin" {
		return errors.New("only admins can reject requests")
	}
	_, err = u.userRepo.FindByID(adminID)
	if err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}
	request, err := u.groupRequestRepo.FindByID(requestID)
	if err != nil {
		return fmt.Errorf("request not found: %w", err)
	}
	if request.Status != "pending" {
		return errors.New("request already processed")
	}
	request.Status = "rejected"
	return u.groupRequestRepo.Update(request)
}

// func (u *GroupUsecase) AddUserToGroup(token string, userID, groupID int) error {
// 	adminID, role, err := u.validateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}
// 	if role != "admin" {
// 		return errors.New("only admins can add users to groups")
// 	}
// 	_, err = u.userRepo.FindByID(adminID)
// 	if err != nil {
// 		return fmt.Errorf("admin not found: %w", err)
// 	}
// 	user, err := u.userRepo.FindByID(userID)
// 	if err != nil {
// 		return fmt.Errorf("user not found: %w", err)
// 	}
// 	_, err = u.groupRepo.FindByID(groupID)
// 	if err != nil {
// 		return fmt.Errorf("group not found: %w", err)
// 	}
// 	for _, g := range user.Groups {
// 		if g.ID == groupID {
// 			return errors.New("user already in group")
// 		}
// 	}
// 	if err := u.userRepo.AddToGroup(userID, groupID); err != nil {
// 		return fmt.Errorf("failed to add user to group: %w", err)
// 	}

// 	event := map[string]interface{}{
// 		"type":     "user_joined",
// 		"group_id": groupID,
// 		"user_id":  userID,
// 	}
// 	eventJSON, _ := json.Marshal(event)
// 	u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON))

// 	return nil
// }

// func (u *GroupUsecase) RemoveUserFromGroup(token string, userID, groupID int) error {
// 	adminID, role, err := u.validateTokenWithRedis(token)
// 	if err != nil {
// 		return fmt.Errorf("token validation failed: %w", err)
// 	}
// 	if role != "admin" {
// 		return errors.New("only admins can remove users from groups")
// 	}
// 	_, err = u.userRepo.FindByID(adminID)
// 	if err != nil {
// 		return fmt.Errorf("admin not found: %w", err)
// 	}
// 	user, err := u.userRepo.FindByID(userID)
// 	if err != nil {
// 		return fmt.Errorf("user not found: %w", err)
// 	}

// 	log.Println("removing", groupID, userID)

// 	if _, err := u.groupRepo.FindByID(groupID); err != nil { // Simplified check
// 		return fmt.Errorf("group not found: %w", err)
// 	}

//		log.Println(user.Groups)
//		inGroup := false
//		for _, g := range user.Groups {
//			if g.ID == groupID {
//				inGroup = true
//				break
//			}
//		}
//		if !inGroup {
//			return errors.New("user not in group")
//		}
//		if err := u.groupRepo.RemoveMember(groupID, userID); err != nil {
//			return fmt.Errorf("failed to remove user from group: %w", err)
//		}
//		event := map[string]interface{}{
//			"type":     "user_left",
//			"group_id": groupID,
//			"user_id":  userID,
//		}
//		eventJSON, err := json.Marshal(event)
//		if err != nil {
//			// Log: log.Printf("Failed to marshal event: %v", err)
//		} else {
//			u.redisService.PublishMessage("group:"+strconv.Itoa(groupID), string(eventJSON))
//		}
//		return nil
//	}
func (u *GroupUsecase) validateTokenWithRedis(token string) (int, string, error) {

	log.Println("validation with rid")

	userID, role, err := u.authService.ValidateToken(token)

	log.Println(role, userID)
	if err != nil {
		return 0, "", fmt.Errorf("jwt validation failed: %w", err)
	}
	storedToken, err := u.redisService.GetToken(userID)

	log.Println("ttocken,", storedToken)
	log.Println("tocen............,", token)

	tokenstr := strings.TrimPrefix(token, "Bearer")
	tcn := strings.TrimSpace(tokenstr)

	if err != nil {
		return 0, "", fmt.Errorf("redis error: %w", err)
	}
	if storedToken == "" || storedToken != tcn {
		return 0, "", errors.New("token not active or mismatched")
	}
	return userID, role, nil
}

//getting group which is user added groups

func (u *GroupUsecase) GetUserGroups(userID int) ([]*domain.Group, error) {
	groups, err := u.groupRepo.GetAllGroupsByUserID(userID)

	log.Println(groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (u *GroupUsecase) GetAllGroupswithMember(userID int) ([]*domain.Group, error) {
	// Fetch all groups with membership status
	groups, err := u.groupRepo.GetAllWithMembership(userID)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (u *GroupUsecase) AddUserToGroup(token string, userID, groupID int) error {

	adminID, _, err := u.validateTokenWithRedis(token)
	if err != nil {
		return err
	}
	// Fetch admin to check role
	admin, err := u.userRepo.FindByID(adminID)
	if err != nil || admin.Role != "admin" {
		return errors.New("unauthorized: only admins can add users to groups")
	}

	if err := u.groupRepo.AddMember(userID, groupID); err != nil {
		return err
	}

	if groupID == domain.AdminGroupID {
		user, err := u.userRepo.FindByID(userID)
		if err != nil {
			return err
		}
		if user.Role != "admin" {
			user.Role = "admin"
			if err := u.userRepo.Update(userID, user); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *GroupUsecase) RemoveUserFromGroup(token string, userID, groupID int) error {
	adminID, _, err := u.validateTokenWithRedis(token)
	if err != nil {
		return err
	}
	// Fetch admin to check role
	admin, err := u.userRepo.FindByID(adminID)
	if err != nil || admin.Role != "admin" {
		return errors.New("unauthorized: only admins can add users to groups")
	}

	// Check if this is the last admin
	if groupID == domain.AdminGroupID {
		user, err := u.userRepo.FindByID(userID)
		if err != nil {
			return err
		}
		if user.Role == "admin" {
			adminCount, err := u.userRepo.CountAdmins()
			if err != nil {
				return err
			}
			// If removing this admin leaves no admins, block the action
			if adminCount <= 1 {
				return errors.New("cannot remove the last admin from the admin group")
			}
		}
	}

	// Remove user from group
	if err := u.groupRepo.RemoveMember(userID, groupID); err != nil {
		return err
	}

	// Revert role if removing from admin group
	if groupID == domain.AdminGroupID {
		user, err := u.userRepo.FindByID(userID)
		if err != nil {
			return err
		}
		groups, err := u.groupRepo.GetAllGroupsByUserID(userID)
		if err != nil {
			return err
		}
		isStillAdmin := false
		for _, group := range groups {
			if group.ID == domain.AdminGroupID {
				isStillAdmin = true
				break
			}
		}
		if !isStillAdmin && user.Role == "admin" {
			user.Role = "employee"
			if err := u.userRepo.Update(userID, user); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *GroupUsecase) DeleteGroup(token string, groupID int) error {
	adminID, _, err := u.validateTokenWithRedis(token)
	if err != nil {
		return err
	}
	admin, err := u.userRepo.FindByID(adminID)
	if err != nil || admin.Role != "admin" {
		return errors.New("unauthorized: only admins can delete groups")
	}

	// Fetch the group using repository
	group, err := u.groupRepo.FindByID(groupID)
	if err != nil {
		return errors.New("group not found")
	}
	fmt.Println(group) // Debugging line

	// If deleting the admin group, check for remaining admins
	if groupID == domain.AdminGroupID {
		adminCount, err := u.userRepo.CountAdmins()
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return errors.New("cannot delete the admin group if it would leave no admins")
		}
	}

	// Delegate deletion to repository
	if err := u.groupRepo.Delete(groupID); err != nil {
		return err
	}

	// If it was the admin group, update users' roles
	if groupID == domain.AdminGroupID {
		users, err := u.userRepo.GetAllUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			isStillAdmin := false
			for _, g := range user.Groups {
				if g.ID == domain.AdminGroupID {
					isStillAdmin = true
					break
				}
			}
			if !isStillAdmin && user.Role == "admin" {
				if err := u.userRepo.UpdateRole(user.ID, "user"); err != nil { // Changed to "user"
					return err
				}
			}
		}
	}

	return nil
}
