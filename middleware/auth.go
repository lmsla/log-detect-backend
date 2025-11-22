package middleware

import (
	"log-detect/entities"
	"log-detect/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate JWT token
		authService := services.NewAuthService()
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_id", claims.RoleID)

		c.Next()
	}
}

// PermissionMiddleware checks if user has required permission
func PermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		authService := services.NewAuthService()
		hasPermission, err := authService.CheckPermission(userID.(uint), resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware optionally validates JWT tokens but doesn't abort if missing
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString := tokenParts[1]

				authService := services.NewAuthService()
				claims, err := authService.ValidateJWT(tokenString)
				if err == nil {
					// Set user information in context if token is valid
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("role_id", claims.RoleID)
				}
			}
		}
		c.Next()
	}
}

// GetCurrentUser retrieves current user from context
func GetCurrentUser(c *gin.Context) (*entities.User, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, false
	}

	authService := services.NewAuthService()
	user, err := authService.GetUserByID(userID.(uint))
	if err != nil {
		return nil, false
	}

	return user, true
}

// RequireRole checks if user has one of the required roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user's role is in the allowed roles
		for _, role := range roles {
			if user.Role.Name == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role permissions"})
		c.Abort()
	}
}

// AdminOnly middleware allows only admin users
func AdminOnly() gin.HandlerFunc {
	return RequireRole("admin")
}
