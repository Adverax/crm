package auth

import (
	"time"

	"github.com/google/uuid"
)

// TokenPair holds access and refresh tokens returned to the client.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken represents a stored refresh token record.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PasswordResetToken represents a stored password reset token record.
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// LoginInput represents the login request body.
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshInput represents the token refresh request body.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordInput represents the forgot password request body.
type ForgotPasswordInput struct {
	Email string `json:"email"`
}

// ResetPasswordInput represents the reset password request body.
type ResetPasswordInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// SetPasswordInput represents the admin set-password request body.
type SetPasswordInput struct {
	Password string `json:"password"`
}

// UserInfo represents the current user info returned by /me.
type UserInfo struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	ProfileID uuid.UUID  `json:"profile_id"`
	RoleID    *uuid.UUID `json:"role_id"`
	IsActive  bool       `json:"is_active"`
}

// UserWithPassword holds user data including password hash for auth checks.
type UserWithPassword struct {
	ID           uuid.UUID
	Username     string
	Email        string
	FirstName    string
	LastName     string
	ProfileID    uuid.UUID
	RoleID       *uuid.UUID
	IsActive     bool
	PasswordHash string
}

// AccessTokenClaims represents the JWT claims for access tokens.
type AccessTokenClaims struct {
	Sub string `json:"sub"`
	Pid string `json:"pid"`
	Rid string `json:"rid,omitempty"`
}
