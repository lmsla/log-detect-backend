package controller

import (
	"log-detect/entities"
	"log-detect/middleware"
	"log-detect/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary User Login
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param login body entities.LoginRequest true "Login credentials"
// @Success 200 {object} entities.LoginResponse
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var loginReq entities.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	authService := services.NewAuthService()
	response, err := authService.Login(loginReq.Username, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Register User
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param user body entities.User true "User registration data"
// @Success 200 {object} entities.User
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var user entities.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Check if current user has permission to create users
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	authService := services.NewAuthService()
	hasPermission, err := authService.CheckPermission(currentUser.ID, "user", "create")
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create users"})
		return
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username, email, and password are required"})
		return
	}

	if user.RoleID == 0 {
		// Default to user role if not specified
		userRole, err := services.GetRoleByName("user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get default user role"})
			return
		}
		user.RoleID = userRole.ID
	}

	registeredUser, err := authService.Register(user.Username, user.Email, user.Password, user.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't return password in response
	registeredUser.Password = ""

	c.JSON(http.StatusOK, registeredUser)
}

// @Summary Get Current User Profile
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} entities.User
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/profile [get]
func GetProfile(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Don't return password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// @Summary Refresh Token
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} entities.LoginResponse
// @Failure 401 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	authService := services.NewAuthService()
	token, err := authService.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := entities.LoginResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary List Users
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {array} entities.User
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/users [get]
func ListUsers(c *gin.Context) {
	// Check if current user has permission to list users
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	authService := services.NewAuthService()
	hasPermission, err := authService.CheckPermission(currentUser.ID, "user", "read")
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to list users"})
		return
	}

	var users []entities.User
	if err := services.ListAllUsers(&users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Get User by ID
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} entities.User
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 404 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/users/{id} [get]
func GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if current user has permission to read users
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	authService := services.NewAuthService()
	hasPermission, err := authService.CheckPermission(currentUser.ID, "user", "read")
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read users"})
		return
	}

	// Allow users to read their own profile or admins can read any user
	if uint(userID) != currentUser.ID && currentUser.Role.Name != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only access own profile"})
		return
	}

	user, err := authService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// @Summary Update User
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Param user body entities.User true "Updated user data"
// @Success 200 {object} entities.User
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/users/{id} [put]
func UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateData entities.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Check if current user has permission to update users
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	authService := services.NewAuthService()

	// Users can only update their own profile, admins can update any user
	canUpdate := false
	if uint(userID) == currentUser.ID {
		canUpdate = true
	} else {
		hasPermission, err := authService.CheckPermission(currentUser.ID, "user", "update")
		if err == nil && hasPermission {
			canUpdate = true
		}
	}

	if !canUpdate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update user"})
		return
	}

	// Update user logic would go here
	// For now, return success
	c.JSON(http.StatusOK, gin.H{"message": "User update not implemented yet"})
}

// @Summary Delete User
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Security ApiKeyAuth
// @Router /auth/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if current user has permission to delete users
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	authService := services.NewAuthService()
	hasPermission, err := authService.CheckPermission(currentUser.ID, "user", "delete")
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete users"})
		return
	}

	// Users cannot delete themselves
	if uint(userID) == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete own account"})
		return
	}

	// Delete user logic would go here
	// For now, return success
	c.JSON(http.StatusOK, gin.H{"message": "User deletion not implemented yet"})
}
