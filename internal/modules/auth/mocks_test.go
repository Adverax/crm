package auth

import (
	"context"

	"github.com/google/uuid"
)

// --- Mock UserAuthRepository ---

type mockUserAuthRepo struct {
	getByUsernameFn func(ctx context.Context, username string) (*UserWithPassword, error)
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*UserWithPassword, error)
	getByEmailFn    func(ctx context.Context, email string) (*UserWithPassword, error)
	setPasswordFn   func(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

func (m *mockUserAuthRepo) GetByUsername(ctx context.Context, username string) (*UserWithPassword, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *mockUserAuthRepo) GetByID(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserAuthRepo) GetByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserAuthRepo) SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if m.setPasswordFn != nil {
		return m.setPasswordFn(ctx, userID, passwordHash)
	}
	return nil
}

// --- Mock RefreshTokenRepository ---

type mockRefreshTokenRepo struct {
	createFn          func(ctx context.Context, token *RefreshToken) error
	getByTokenHashFn  func(ctx context.Context, tokenHash string) (*RefreshToken, error)
	deleteByHashFn    func(ctx context.Context, tokenHash string) error
	deleteAllByUserFn func(ctx context.Context, userID uuid.UUID) error
	deleteExpiredFn   func(ctx context.Context) (int64, error)
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *RefreshToken) error {
	if m.createFn != nil {
		return m.createFn(ctx, token)
	}
	return nil
}

func (m *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(ctx, tokenHash)
	}
	return nil, nil
}

func (m *mockRefreshTokenRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if m.deleteByHashFn != nil {
		return m.deleteByHashFn(ctx, tokenHash)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.deleteAllByUserFn != nil {
		return m.deleteAllByUserFn(ctx, userID)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return 0, nil
}

// --- Mock PasswordResetTokenRepository ---

type mockResetTokenRepo struct {
	createFn         func(ctx context.Context, token *PasswordResetToken) error
	getByTokenHashFn func(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	markUsedFn       func(ctx context.Context, id uuid.UUID) error
	deleteExpiredFn  func(ctx context.Context) (int64, error)
}

func (m *mockResetTokenRepo) Create(ctx context.Context, token *PasswordResetToken) error {
	if m.createFn != nil {
		return m.createFn(ctx, token)
	}
	return nil
}

func (m *mockResetTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(ctx, tokenHash)
	}
	return nil, nil
}

func (m *mockResetTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, id)
	}
	return nil
}

func (m *mockResetTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return 0, nil
}

// --- Mock EmailSender ---

type mockEmailSender struct {
	sendResetFn func(ctx context.Context, email, resetURL string) error
}

func (m *mockEmailSender) SendPasswordReset(ctx context.Context, email, resetURL string) error {
	if m.sendResetFn != nil {
		return m.sendResetFn(ctx, email, resetURL)
	}
	return nil
}
