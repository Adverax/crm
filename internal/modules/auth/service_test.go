package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestService(
	userRepo *mockUserAuthRepo,
	refreshRepo *mockRefreshTokenRepo,
	resetRepo *mockResetTokenRepo,
	emailSender *mockEmailSender,
) *ServiceImpl {
	if userRepo == nil {
		userRepo = &mockUserAuthRepo{}
	}
	if refreshRepo == nil {
		refreshRepo = &mockRefreshTokenRepo{}
	}
	if resetRepo == nil {
		resetRepo = &mockResetTokenRepo{}
	}
	if emailSender == nil {
		emailSender = &mockEmailSender{}
	}
	return NewService(ServiceConfig{
		UserRepo:     userRepo,
		RefreshRepo:  refreshRepo,
		ResetRepo:    resetRepo,
		EmailSender:  emailSender,
		JWTSecret:    "test-secret-key-for-testing-only",
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   7 * 24 * time.Hour,
		ResetBaseURL: "http://localhost:5173/reset-password",
	})
}

func testUser() *UserWithPassword {
	hash, _ := HashPassword("correctpassword")
	return &UserWithPassword{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		ProfileID:    uuid.New(),
		RoleID:       nil,
		IsActive:     true,
		PasswordHash: hash,
	}
}

func TestServiceImpl_Login(t *testing.T) {
	t.Parallel()

	user := testUser()

	tests := []struct {
		name      string
		input     LoginInput
		userRepo  *mockUserAuthRepo
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "successful login",
			input: LoginInput{Username: user.Username, Password: "correctpassword"},
			userRepo: &mockUserAuthRepo{
				getByUsernameFn: func(_ context.Context, username string) (*UserWithPassword, error) {
					if username == user.Username {
						return user, nil
					}
					return nil, nil
				},
			},
			wantErr: false,
		},
		{
			name:      "empty username",
			input:     LoginInput{Username: "", Password: "password"},
			wantErr:   true,
			errSubstr: "username is required",
		},
		{
			name:      "empty password",
			input:     LoginInput{Username: "admin", Password: ""},
			wantErr:   true,
			errSubstr: "password is required",
		},
		{
			name:  "user not found",
			input: LoginInput{Username: "nonexistent", Password: "password"},
			userRepo: &mockUserAuthRepo{
				getByUsernameFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					return nil, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid credentials",
		},
		{
			name:  "wrong password",
			input: LoginInput{Username: user.Username, Password: "wrongpassword"},
			userRepo: &mockUserAuthRepo{
				getByUsernameFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					return user, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid credentials",
		},
		{
			name:  "inactive user",
			input: LoginInput{Username: "inactive", Password: "password"},
			userRepo: &mockUserAuthRepo{
				getByUsernameFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					u := *user
					u.IsActive = false
					return &u, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid credentials",
		},
		{
			name:  "password not set",
			input: LoginInput{Username: "nopwd", Password: "password"},
			userRepo: &mockUserAuthRepo{
				getByUsernameFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					u := *user
					u.PasswordHash = ""
					return &u, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid credentials",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(tt.userRepo, nil, nil, nil)
			pair, err := svc.Login(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Login() error = %v, want containing %q", err, tt.errSubstr)
				}
			}
			if !tt.wantErr {
				if pair == nil {
					t.Fatal("Login() returned nil pair")
				}
				if pair.AccessToken == "" {
					t.Error("Login() access token is empty")
				}
				if pair.RefreshToken == "" {
					t.Error("Login() refresh token is empty")
				}
			}
		})
	}
}

func TestServiceImpl_Refresh(t *testing.T) {
	t.Parallel()

	user := testUser()
	rawToken := "valid-refresh-token-hex-string"
	tokenHash := HashToken(rawToken)

	tests := []struct {
		name        string
		input       RefreshInput
		refreshRepo *mockRefreshTokenRepo
		userRepo    *mockUserAuthRepo
		wantErr     bool
		errSubstr   string
	}{
		{
			name:  "successful refresh",
			input: RefreshInput{RefreshToken: rawToken},
			refreshRepo: &mockRefreshTokenRepo{
				getByTokenHashFn: func(_ context.Context, hash string) (*RefreshToken, error) {
					if hash == tokenHash {
						return &RefreshToken{
							ID:        uuid.New(),
							UserID:    user.ID,
							TokenHash: hash,
							ExpiresAt: time.Now().Add(time.Hour),
						}, nil
					}
					return nil, nil
				},
			},
			userRepo: &mockUserAuthRepo{
				getByIDFn: func(_ context.Context, id uuid.UUID) (*UserWithPassword, error) {
					if id == user.ID {
						return user, nil
					}
					return nil, nil
				},
			},
			wantErr: false,
		},
		{
			name:      "empty refresh token",
			input:     RefreshInput{RefreshToken: ""},
			wantErr:   true,
			errSubstr: "refresh_token is required",
		},
		{
			name:  "invalid refresh token",
			input: RefreshInput{RefreshToken: "nonexistent"},
			refreshRepo: &mockRefreshTokenRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*RefreshToken, error) {
					return nil, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid refresh token",
		},
		{
			name:  "expired refresh token",
			input: RefreshInput{RefreshToken: rawToken},
			refreshRepo: &mockRefreshTokenRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*RefreshToken, error) {
					return &RefreshToken{
						ID:        uuid.New(),
						UserID:    user.ID,
						ExpiresAt: time.Now().Add(-time.Hour),
					}, nil
				},
			},
			wantErr:   true,
			errSubstr: "expired",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(tt.userRepo, tt.refreshRepo, nil, nil)
			pair, err := svc.Refresh(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Refresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Refresh() error = %v, want containing %q", err, tt.errSubstr)
				}
			}
			if !tt.wantErr && pair == nil {
				t.Error("Refresh() returned nil pair")
			}
		})
	}
}

func TestServiceImpl_Me(t *testing.T) {
	t.Parallel()

	user := testUser()

	tests := []struct {
		name      string
		userID    uuid.UUID
		userRepo  *mockUserAuthRepo
		wantErr   bool
		errSubstr string
	}{
		{
			name:   "returns user info",
			userID: user.ID,
			userRepo: &mockUserAuthRepo{
				getByIDFn: func(_ context.Context, id uuid.UUID) (*UserWithPassword, error) {
					if id == user.ID {
						return user, nil
					}
					return nil, nil
				},
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: uuid.New(),
			userRepo: &mockUserAuthRepo{
				getByIDFn: func(_ context.Context, _ uuid.UUID) (*UserWithPassword, error) {
					return nil, nil
				},
			},
			wantErr:   true,
			errSubstr: "NOT_FOUND",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(tt.userRepo, nil, nil, nil)
			info, err := svc.Me(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Me() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if info == nil {
					t.Fatal("Me() returned nil")
				}
				if info.Username != user.Username {
					t.Errorf("Me() username = %q, want %q", info.Username, user.Username)
				}
			}
		})
	}
}

func TestServiceImpl_SetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		password  string
		wantErr   bool
		errSubstr string
	}{
		{name: "valid password", password: "newpassword123", wantErr: false},
		{name: "too short", password: "short", wantErr: true, errSubstr: "at least 8"},
		{name: "too long", password: strings.Repeat("a", 129), wantErr: true, errSubstr: "at most 128"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(nil, nil, nil, nil)
			err := svc.SetPassword(context.Background(), uuid.New(), tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("SetPassword() error = %v, want containing %q", err, tt.errSubstr)
				}
			}
		})
	}
}

func TestServiceImpl_ForgotPassword(t *testing.T) {
	t.Parallel()

	user := testUser()

	tests := []struct {
		name      string
		input     ForgotPasswordInput
		userRepo  *mockUserAuthRepo
		emailSent bool
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "sends reset email for existing user",
			input: ForgotPasswordInput{Email: user.Email},
			userRepo: &mockUserAuthRepo{
				getByEmailFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					return user, nil
				},
			},
			emailSent: true,
			wantErr:   false,
		},
		{
			name:  "returns success for non-existent email",
			input: ForgotPasswordInput{Email: "nobody@example.com"},
			userRepo: &mockUserAuthRepo{
				getByEmailFn: func(_ context.Context, _ string) (*UserWithPassword, error) {
					return nil, nil
				},
			},
			emailSent: false,
			wantErr:   false,
		},
		{
			name:      "validates empty email",
			input:     ForgotPasswordInput{Email: ""},
			wantErr:   true,
			errSubstr: "email is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var emailCalled bool
			email := &mockEmailSender{
				sendResetFn: func(_ context.Context, _, _ string) error {
					emailCalled = true
					return nil
				},
			}
			svc := newTestService(tt.userRepo, nil, nil, email)
			err := svc.ForgotPassword(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ForgotPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.emailSent != emailCalled {
				t.Errorf("ForgotPassword() emailSent = %v, want %v", emailCalled, tt.emailSent)
			}
		})
	}
}

func TestServiceImpl_ResetPassword(t *testing.T) {
	t.Parallel()

	rawToken := "reset-token-hex"
	tokenHash := HashToken(rawToken)
	userID := uuid.New()

	tests := []struct {
		name      string
		input     ResetPasswordInput
		resetRepo *mockResetTokenRepo
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "successful reset",
			input: ResetPasswordInput{Token: rawToken, Password: "newsecurepassword"},
			resetRepo: &mockResetTokenRepo{
				getByTokenHashFn: func(_ context.Context, hash string) (*PasswordResetToken, error) {
					if hash == tokenHash {
						return &PasswordResetToken{
							ID:        uuid.New(),
							UserID:    userID,
							TokenHash: hash,
							ExpiresAt: time.Now().Add(time.Hour),
							UsedAt:    nil,
						}, nil
					}
					return nil, nil
				},
			},
			wantErr: false,
		},
		{
			name:      "empty token",
			input:     ResetPasswordInput{Token: "", Password: "newpassword"},
			wantErr:   true,
			errSubstr: "token is required",
		},
		{
			name:      "short password",
			input:     ResetPasswordInput{Token: "abc", Password: "short"},
			wantErr:   true,
			errSubstr: "at least 8",
		},
		{
			name:  "invalid token",
			input: ResetPasswordInput{Token: "nonexistent", Password: "newpassword1"},
			resetRepo: &mockResetTokenRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*PasswordResetToken, error) {
					return nil, nil
				},
			},
			wantErr:   true,
			errSubstr: "invalid or expired",
		},
		{
			name:  "expired token",
			input: ResetPasswordInput{Token: rawToken, Password: "newpassword1"},
			resetRepo: &mockResetTokenRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*PasswordResetToken, error) {
					return &PasswordResetToken{
						ID:        uuid.New(),
						UserID:    userID,
						ExpiresAt: time.Now().Add(-time.Hour),
					}, nil
				},
			},
			wantErr:   true,
			errSubstr: "expired",
		},
		{
			name:  "already used token",
			input: ResetPasswordInput{Token: rawToken, Password: "newpassword1"},
			resetRepo: &mockResetTokenRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*PasswordResetToken, error) {
					usedAt := time.Now().Add(-time.Minute)
					return &PasswordResetToken{
						ID:        uuid.New(),
						UserID:    userID,
						ExpiresAt: time.Now().Add(time.Hour),
						UsedAt:    &usedAt,
					}, nil
				},
			},
			wantErr:   true,
			errSubstr: "already been used",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(nil, nil, tt.resetRepo, nil)
			err := svc.ResetPassword(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ResetPassword() error = %v, want containing %q", err, tt.errSubstr)
				}
			}
		})
	}
}

func TestServiceImpl_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		refreshToken string
		wantErr      bool
	}{
		{name: "successful logout", refreshToken: "some-token", wantErr: false},
		{name: "empty token is no-op", refreshToken: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := newTestService(nil, nil, nil, nil)
			err := svc.Logout(context.Background(), tt.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
