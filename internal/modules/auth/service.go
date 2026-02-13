package auth

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the auth module business logic.
type Service interface {
	Login(ctx context.Context, input LoginInput) (*TokenPair, error)
	Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	Me(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
	SetPassword(ctx context.Context, userID uuid.UUID, password string) error
	ForgotPassword(ctx context.Context, input ForgotPasswordInput) error
	ResetPassword(ctx context.Context, input ResetPasswordInput) error
}
