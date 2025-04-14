package usecase

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Sherinas/Chat-App-Clean/Internal/domain"
	"github.com/Sherinas/Chat-App-Clean/Internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken(userID int, role string) (string, error)
	ValidateToken(token string) (int, string, error)
	HashPassword(password string) (string, error)
}

type UserUsecase struct {
	userRepo     repository.UserRepository
	otpRepo      repository.OTPRepository
	authService  AuthService
	redisService RedisService
}

func NewUserUsecase(userRepo repository.UserRepository, otpRepo repository.OTPRepository, authService AuthService, redisService RedisService) *UserUsecase {
	return &UserUsecase{
		userRepo:     userRepo,
		otpRepo:      otpRepo,
		authService:  authService,
		redisService: redisService,
	}
}

// CreateEmployeeID (Admin-only) - No Redis change needed
func (u *UserUsecase) CreateEmployeeID(name, email string) (string, error) {
	employeeID := "EMP" + time.Now().Format("205")
	user := &domain.User{
		EmployeeID: employeeID,
		Name:       name,
		Email:      email,
		Role:       "employee",
		State:      "offline",
	}
	if err := u.userRepo.Create(user); err != nil {
		return "", err
	}
	return employeeID, nil
}

func (u *UserUsecase) CreateAdminUser(employeeID, name, email, password string) (string, error) {
	// Check if user already exists
	if _, err := u.userRepo.FindByEmployeeID(employeeID); err == nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := u.authService.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		EmployeeID: employeeID,
		Name:       name,
		Email:      email,
		Password:   hashedPassword,
		Role:       "admin", // Set as admin
		CreatedAt:  time.Now(),
	}
	if err := u.userRepo.Create(user); err != nil {
		return "", err
	}

	// Generate token
	token, err := u.authService.GenerateToken(user.ID, "admin")
	if err != nil {
		return "", err
	}
	if err := u.redisService.SetToken(user.ID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserUsecase) SignUpWithEmployeeID(employeeID, password, mobile, designation string) (string, error) {
	user, err := u.userRepo.FindByEmployeeID(employeeID)
	if err != nil {
		return "", errors.New("employee ID not found")
	}
	if user.Password != "" {
		return "", errors.New("user already signed up")
	}

	otp := generateOTP()
	hashedPassword, err := u.authService.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Prepare data to store in Redis
	type signupData struct {
		EmployeeID     string `json:"employee_id"`
		HashedPassword string `json:"hashed_password"`
		Email          string `json:"email"`
		Mobile         string `json:"mobile"`
		Designation    string `json:"designation"`
	}
	data := signupData{
		EmployeeID:     employeeID,
		HashedPassword: hashedPassword,
		Email:          user.Email,
		Mobile:         mobile,
		Designation:    designation,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Store in Redis with OTP as key, expire in 5 minutes
	if err := u.redisService.SetWithTTL("signup:otp:"+otp, string(dataBytes), 5*time.Minute); err != nil {
		return "", err
	}

	// Optionally store OTP in otpRepo if still needed
	if err := u.otpRepo.Create(&domain.OTPVerification{
		EmployeeID: employeeID,
		OTP:        otp,
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}); err != nil {
		return "", err
	}

	return otp, nil

}

func (u *UserUsecase) VerifyOTP(otp string) (string, error) {
	// Retrieve signup data from Redis
	dataStr, err := u.redisService.Get("signup:otp:" + otp)
	if err != nil || dataStr == "" {
		return "", errors.New("invalid or expired OTP")
	}

	// Unmarshal the stored data
	type signupData struct {
		EmployeeID     string `json:"employee_id"`
		HashedPassword string `json:"hashed_password"`
		Mobile         string `json:"mobile"`
		Designation    string `json:"designation"`
	}
	var data signupData
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return "", err
	}

	// Find user by employeeID from stored data
	user, err := u.userRepo.FindByEmployeeID(data.EmployeeID)
	if err != nil {
		return "", errors.New("user not found")
	}
	if user.Password != "" {
		return "", errors.New("user already signed up")
	}

	user.Designation = data.Designation
	user.Password = data.HashedPassword
	user.State = "online" // Assuming State is a field; adjust if it’s Role
	user.UpdateAt = time.Now()
	if err := u.userRepo.Update(user.ID, user); err != nil {
		return "", err
	}

	// Generate token
	token, err := u.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	// Store token and status in Redis
	if err := u.redisService.SetToken(user.ID, token); err != nil {
		return "", err
	}
	if err := u.redisService.SetUserStatus(user.ID, "online"); err != nil {
		return "", err
	}

	// Publish login event
	event := map[string]interface{}{
		"type":    "user_login",
		"user_id": user.ID,
	}
	eventJSON, _ := json.Marshal(event)
	u.redisService.PublishMessage("user_events", string(eventJSON))
	key := "signup:otp:" + otp
	// Clean up Redis
	u.redisService.Delete(key)

	////////////////////////////////////////////////////////////////////////////////////////////////////delete  otp from db

	return token, nil
}

// LoginWithEmployeeID - Add Redis for token and status
// func (u *UserUsecase) LoginWithEmployeeID(employeeID, password string) (string, error) {
// 	user, err := u.userRepo.FindByEmployeeID(employeeID)

// 	if err != nil {
// 		return "", errors.New("invalid employee ID")
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
// 		return "", errors.New("invalid password")
// 	}

// 	user.State = "online"
// 	if err := u.userRepo.Update(user.ID, user); err != nil {
// 		return "", err
// 	}

// 	token, err := u.authService.GenerateToken(user.ID, user.Role)
// 	if err != nil {
// 		return "", err
// 	}

// 	u.redisService.SetToken(user.ID, token)
// 	u.redisService.SetUserStatus(user.ID, "online")

//		event := map[string]interface{}{
//			"type":    "user_login",
//			"user_id": user.ID,
//		}
//		eventJSON, _ := json.Marshal(event)
//		u.redisService.PublishMessage("user_events", string(eventJSON))
//		return token, nil
//	}
func (u *UserUsecase) LoginWithEmployeeID(employeeID, password string) (string, int, error) {
	if employeeID == "" || password == "" {
		return "", 0, errors.New("employee ID and password are required")
	}

	// Find user by EmployeeID
	user, err := u.userRepo.FindByEmployeeID(employeeID)
	if err != nil {
		return "", 0, errors.New("failed to retrieve user: " + err.Error())
	}

	// Check if user is soft-deleted
	if user.DeletedAt != nil {
		return "", 0, errors.New("account has been deactivated")
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", 0, errors.New("invalid password")
		}
		return "", 0, errors.New("password verification failed: " + err.Error())
	}

	// Update user state to "online"
	if user.State != "online" {
		user.State = "online"
		if err := u.userRepo.Update(user.ID, user); err != nil {
			return "", 0, errors.New("failed to update user state: " + err.Error())
		}
	}

	// Generate and store token
	token, err := u.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", 0, errors.New("failed to generate token: " + err.Error())
	}

	// Store token and status in Redis
	if err := u.redisService.SetToken(user.ID, token); err != nil {
		return "", 0, errors.New("failed to store token: " + err.Error())
	}
	if err := u.redisService.SetUserStatus(user.ID, "online"); err != nil {
		// Optionally roll back token if status fails, but for now log and continue
		log.Printf("Warning: Failed to set user status in Redis: %v", err)
	}

	// Publish login event
	event := map[string]interface{}{
		"type":    "user_login",
		"user_id": user.ID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Warning: Failed to marshal login event: %v", err)
	} else {
		if err := u.redisService.PublishMessage("user_events", string(eventJSON)); err != nil {
			log.Printf("Warning: Failed to publish login event: %v", err)
		}
	}

	return token, user.ID, nil
}

// Logout - Remove token and update status in Redis
func (u *UserUsecase) Logout(token string) error {
	userID, _, err := u.authService.ValidateToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	user.State = "offline"
	if err := u.userRepo.Update(userID, user); err != nil {
		return err
	}

	u.redisService.RemoveToken(userID)
	u.redisService.SetUserStatus(userID, "offline")

	event := map[string]interface{}{
		"type":    "user_logout",
		"user_id": userID,
	}
	eventJSON, _ := json.Marshal(event)
	u.redisService.PublishMessage("user_events", string(eventJSON))
	return nil
}

func generateOTP() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return fmt.Sprintf("%06x", bytes)[0:6] // 6-digit hex
}

func (u *UserUsecase) GetAllUsers() ([]domain.User, error) {

	users, err := u.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// Enrich with Redis status
	for _, user := range users {
		status, err := u.redisService.GetUserStatus(user.ID)
		if err != nil || status == "" {
			user.State = "offline"
		} else {
			user.State = status
		}
	}
	return users, nil
}

func (u *UserUsecase) FindUserDetails(userID int) (*domain.User, error) {

	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// func (u *UserUsecase) UpdateUser(user *domain.User) error {
// 	return u.userRepo.Update(user)
// }

func (u *UserUsecase) GetUserByID(id int) (*domain.User, error) {
	return u.userRepo.FindByID(id)
}
func (u *UserUsecase) DeleteUser(token string, userID int) error {
	adminID, role, err := u.authService.ValidateToken(token)
	if err != nil {
		return err
	}
	if role != "admin" && adminID != userID {
		return errors.New("unauthorized: only admins or the user themselves can delete the user")
	}

	// Check if user exists and is not already deleted
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	return u.userRepo.Delete(user.ID) // check this
}

func (u *UserUsecase) UpdateUser(token string, userID int, user *domain.User) error {
	adminID, role, err := u.authService.ValidateToken(token)
	if err != nil {
		return err
	}
	if role != "admin" && adminID != userID {
		return errors.New("unauthorized: only admins or the user themselves can update the user")
	}

	// Check if user exists and is not deleted
	_, err = u.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	return u.userRepo.Update(userID, user)
}
