package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

const devUserIDHeader = "X-Dev-User-Id"

// DevAuth creates a development auth middleware that loads user identity from a header.
// In production this will be replaced by JWT middleware (Phase 5).
func DevAuth(userRepo security.UserRepository, defaultUserID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(devUserIDHeader)
		if userIDStr == "" {
			userIDStr = defaultUserID.String()
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid X-Dev-User-Id header"))
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			slog.Error("dev_auth: failed to load user", "user_id", userID, "error", err)
			apperror.Respond(c, apperror.Internal("failed to load user"))
			c.Abort()
			return
		}
		if user == nil {
			apperror.Respond(c, apperror.Unauthorized("user not found"))
			c.Abort()
			return
		}
		if !user.IsActive {
			apperror.Respond(c, apperror.Unauthorized("user is inactive"))
			c.Abort()
			return
		}

		uc := security.UserContext{
			UserID:    user.ID,
			ProfileID: user.ProfileID,
			RoleID:    user.RoleID,
		}
		SetUserContext(c, uc)
		c.Next()
	}
}
