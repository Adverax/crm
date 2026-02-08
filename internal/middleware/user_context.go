package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/security"
)

const userContextKey = "user_context"

// SetUserContext stores UserContext in the Gin context.
func SetUserContext(c *gin.Context, uc security.UserContext) {
	c.Set(userContextKey, uc)
}

// GetUserContext retrieves UserContext from the Gin context.
func GetUserContext(c *gin.Context) (security.UserContext, bool) {
	val, exists := c.Get(userContextKey)
	if !exists {
		return security.UserContext{}, false
	}
	uc, ok := val.(security.UserContext)
	return uc, ok
}

// MustGetUserContext retrieves UserContext or returns zero value.
func MustGetUserContext(c *gin.Context) security.UserContext {
	uc, _ := GetUserContext(c)
	return uc
}

// GetUserID retrieves the current user ID from Gin context.
func GetUserID(c *gin.Context) uuid.UUID {
	uc, _ := GetUserContext(c)
	return uc.UserID
}
