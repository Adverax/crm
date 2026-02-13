package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// JWTAuth creates a JWT authentication middleware.
func JWTAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apperror.Respond(c, apperror.Unauthorized("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			apperror.Respond(c, apperror.Unauthorized("invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			apperror.Respond(c, apperror.Unauthorized("invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			apperror.Respond(c, apperror.Unauthorized("invalid token claims"))
			c.Abort()
			return
		}

		sub, _ := claims.GetSubject()
		userID, err := uuid.Parse(sub)
		if err != nil {
			apperror.Respond(c, apperror.Unauthorized("invalid user id in token"))
			c.Abort()
			return
		}

		pidStr, _ := claims["pid"].(string)
		profileID, err := uuid.Parse(pidStr)
		if err != nil {
			apperror.Respond(c, apperror.Unauthorized("invalid profile id in token"))
			c.Abort()
			return
		}

		var roleID *uuid.UUID
		if ridStr, ok := claims["rid"].(string); ok && ridStr != "" {
			parsed, err := uuid.Parse(ridStr)
			if err == nil {
				roleID = &parsed
			}
		}

		uc := security.UserContext{
			UserID:    userID,
			ProfileID: profileID,
			RoleID:    roleID,
		}
		SetUserContext(c, uc)
		c.Request = c.Request.WithContext(security.ContextWithUser(c.Request.Context(), uc))

		c.Next()
	}
}
