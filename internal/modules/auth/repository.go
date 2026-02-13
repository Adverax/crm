package auth

import (
	"context"

	"github.com/google/uuid"
)

// RefreshTokenRepository manages refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// PasswordResetTokenRepository manages password reset token persistence.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// UserAuthRepository provides user queries needed for authentication.
type UserAuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*UserWithPassword, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserWithPassword, error)
	GetByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}
