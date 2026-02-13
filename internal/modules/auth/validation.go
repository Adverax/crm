package auth

import (
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// ValidateLogin validates a login request.
func ValidateLogin(input LoginInput) error {
	if strings.TrimSpace(input.Username) == "" {
		return apperror.Validation("username is required")
	}
	if input.Password == "" {
		return apperror.Validation("password is required")
	}
	return nil
}

// ValidateRefresh validates a token refresh request.
func ValidateRefresh(input RefreshInput) error {
	if input.RefreshToken == "" {
		return apperror.Validation("refresh_token is required")
	}
	return nil
}

// ValidatePassword validates a password meets minimum requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return apperror.Validation("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return apperror.Validation("password must be at most 128 characters")
	}
	return nil
}

// ValidateForgotPassword validates a forgot password request.
func ValidateForgotPassword(input ForgotPasswordInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return apperror.Validation("email is required")
	}
	if !strings.Contains(input.Email, "@") {
		return apperror.Validation("email must be a valid email address")
	}
	return nil
}

// ValidateResetPassword validates a password reset request.
func ValidateResetPassword(input ResetPasswordInput) error {
	if input.Token == "" {
		return apperror.Validation("token is required")
	}
	return ValidatePassword(input.Password)
}
