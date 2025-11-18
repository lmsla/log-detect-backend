package services

import (
	"errors"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// getJWTSecret returns the JWT secret key from environment variable or default
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// TODO: Remove this fallback in production
		log.Logrecord_no_rotate("WARN", "JWT_SECRET environment variable not set, using default key (NOT SECURE)")
		secret = "your-secret-key-change-in-production"
	}
	return []byte(secret)
}

// JWT configuration
const (
	JWTExpireHours = 24
)

// AuthService handles authentication operations
type AuthService struct{}

// NewAuthService creates a new auth service instance
func NewAuthService() *AuthService {
	return &AuthService{}
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against a hash
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for a user
func (s *AuthService) GenerateJWT(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"exp":      time.Now().Add(time.Hour * JWTExpireHours).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to generate JWT token: %s", err.Error()))
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns claims
func (s *AuthService) ValidateJWT(tokenString string) (*entities.AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok1 := claims["user_id"].(float64)
		username, ok2 := claims["username"].(string)
		roleID, ok3 := claims["role_id"].(float64)

		if !ok1 || !ok2 || !ok3 {
			return nil, errors.New("invalid token claims")
		}

		return &entities.AuthClaims{
			UserID:   uint(userID),
			Username: username,
			RoleID:   uint(roleID),
		}, nil
	}

	return nil, errors.New("invalid token")
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (*entities.LoginResponse, error) {
	// Find user by username
	var user entities.User
	err := global.Mysql.Preload("Role.Permissions").Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Login failed for user %s: %s", username, err.Error()))
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !s.CheckPassword(password, user.Password) {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Invalid password for user %s", username))
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.GenerateJWT(&user)
	if err != nil {
		return nil, err
	}

	return &entities.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// Register creates a new user
func (s *AuthService) Register(username, email, password string, roleID uint) (*entities.User, error) {
	// Check if user already exists
	var existingUser entities.User
	if err := global.Mysql.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := entities.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		RoleID:   roleID,
		IsActive: true,
	}

	if err := global.Mysql.Create(&user).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to create user %s: %s", username, err.Error()))
		return nil, err
	}

	// Load role information
	global.Mysql.Preload("Role.Permissions").First(&user, user.ID)

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID uint) (*entities.User, error) {
	var user entities.User
	err := global.Mysql.Preload("Role.Permissions").Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CheckPermission checks if a user has a specific permission
func (s *AuthService) CheckPermission(userID uint, resource, action string) (bool, error) {
	var user entities.User
	err := global.Mysql.Preload("Role.Permissions").Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		return false, err
	}

	// Check if user has the required permission
	for _, permission := range user.Role.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// CreateDefaultRolesAndPermissions creates default roles and permissions for initial setup
func (s *AuthService) CreateDefaultRolesAndPermissions() error {
	// Create permissions
	permissions := []entities.Permission{
		{Name: "device:create", Resource: "device", Action: "create", Description: "Create devices"},
		{Name: "device:read", Resource: "device", Action: "read", Description: "Read devices"},
		{Name: "device:update", Resource: "device", Action: "update", Description: "Update devices"},
		{Name: "device:delete", Resource: "device", Action: "delete", Description: "Delete devices"},

		{Name: "target:create", Resource: "target", Action: "create", Description: "Create targets"},
		{Name: "target:read", Resource: "target", Action: "read", Description: "Read targets"},
		{Name: "target:update", Resource: "target", Action: "update", Description: "Update targets"},
		{Name: "target:delete", Resource: "target", Action: "delete", Description: "Delete targets"},

		{Name: "indices:create", Resource: "indices", Action: "create", Description: "Create indices"},
		{Name: "indices:read", Resource: "indices", Action: "read", Description: "Read indices"},
		{Name: "indices:update", Resource: "indices", Action: "update", Description: "Update indices"},
		{Name: "indices:delete", Resource: "indices", Action: "delete", Description: "Delete indices"},

		{Name: "user:create", Resource: "user", Action: "create", Description: "Create users"},
		{Name: "user:read", Resource: "user", Action: "read", Description: "Read users"},
		{Name: "user:update", Resource: "user", Action: "update", Description: "Update users"},
		{Name: "user:delete", Resource: "user", Action: "delete", Description: "Delete users"},

		{Name: "elasticsearch:create", Resource: "elasticsearch", Action: "create", Description: "Create Elasticsearch monitors"},
		{Name: "elasticsearch:read", Resource: "elasticsearch", Action: "read", Description: "Read Elasticsearch monitors"},
		{Name: "elasticsearch:update", Resource: "elasticsearch", Action: "update", Description: "Update Elasticsearch monitors"},
		{Name: "elasticsearch:delete", Resource: "elasticsearch", Action: "delete", Description: "Delete Elasticsearch monitors"},
	}

	for _, perm := range permissions {
		if err := global.Mysql.Where("name = ?", perm.Name).FirstOrCreate(&perm).Error; err != nil {
			return err
		}
	}

	// Create roles
	adminRole := entities.Role{
		Name:        "admin",
		Description: "Administrator with full access",
	}
	userRole := entities.Role{
		Name:        "user",
		Description: "Regular user with limited access",
	}

	// Create admin role
	if err := global.Mysql.Where("name = ?", "admin").FirstOrCreate(&adminRole).Error; err != nil {
		return err
	}

	// Create user role
	if err := global.Mysql.Where("name = ?", "user").FirstOrCreate(&userRole).Error; err != nil {
		return err
	}

	// Assign all permissions to admin role
	var allPermissions []entities.Permission
	if err := global.Mysql.Find(&allPermissions).Error; err != nil {
		return err
	}

	if err := global.Mysql.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		return err
	}

	// Assign read permissions to user role
	var readPermissions []entities.Permission
	if err := global.Mysql.Where("action = ?", "read").Find(&readPermissions).Error; err != nil {
		return err
	}

	if err := global.Mysql.Model(&userRole).Association("Permissions").Replace(readPermissions); err != nil {
		return err
	}

	return nil
}

// CreateDefaultAdmin creates a default admin user
func (s *AuthService) CreateDefaultAdmin() error {
	// Get admin role
	var adminRole entities.Role
	if err := global.Mysql.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	// Check if admin user already exists
	var existingAdmin entities.User
	if err := global.Mysql.Where("username = ?", "admin").First(&existingAdmin).Error; err == nil {
		return nil // Admin already exists
	}

	// Create default admin user
	_, err := s.Register("admin", "admin@logdetect.com", "admin123", adminRole.ID)
	return err
}

// GetRoleByName retrieves a role by name
func GetRoleByName(name string) (*entities.Role, error) {
	var role entities.Role
	if err := global.Mysql.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// ListAllUsers retrieves all active users with their roles
func ListAllUsers(users *[]entities.User) error {
	return global.Mysql.Preload("Role").Where("is_active = ?", true).Find(users).Error
}
