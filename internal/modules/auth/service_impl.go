package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// ServiceImpl implements the Service interface.
type ServiceImpl struct {
	userRepo     UserAuthRepository
	refreshRepo  RefreshTokenRepository
	resetRepo    PasswordResetTokenRepository
	emailSender  EmailSender
	jwtSecret    []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
	resetBaseURL string
}

// ServiceConfig holds configuration for creating a ServiceImpl.
type ServiceConfig struct {
	UserRepo     UserAuthRepository
	RefreshRepo  RefreshTokenRepository
	ResetRepo    PasswordResetTokenRepository
	EmailSender  EmailSender
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	ResetBaseURL string
}

// NewService creates a new auth service.
func NewService(cfg ServiceConfig) *ServiceImpl {
	return &ServiceImpl{
		userRepo:     cfg.UserRepo,
		refreshRepo:  cfg.RefreshRepo,
		resetRepo:    cfg.ResetRepo,
		emailSender:  cfg.EmailSender,
		jwtSecret:    []byte(cfg.JWTSecret),
		accessTTL:    cfg.AccessTTL,
		refreshTTL:   cfg.RefreshTTL,
		resetBaseURL: cfg.ResetBaseURL,
	}
}

func (s *ServiceImpl) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	if err := ValidateLogin(input); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}

	// Constant-time comparison even when user doesn't exist (prevents username enumeration).
	// This is a valid bcrypt cost-12 hash to ensure the comparison takes the same time.
	const dummyHash = "$2a$12$rJiQbXxC1tSSSBJk/g6n8uzcvtLHBId8/e4lCrAusHvhyo8fCOin."
	if user == nil {
		_ = CheckPassword(dummyHash, input.Password)
		return nil, apperror.Unauthorized("invalid credentials")
	}

	if !user.IsActive {
		_ = CheckPassword(dummyHash, input.Password)
		return nil, apperror.Unauthorized("invalid credentials")
	}

	if user.PasswordHash == "" {
		_ = CheckPassword(dummyHash, input.Password)
		return nil, apperror.Unauthorized("invalid credentials")
	}

	if err := CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	return s.issueTokenPair(ctx, user)
}

func (s *ServiceImpl) Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	if err := ValidateRefresh(input); err != nil {
		return nil, err
	}

	tokenHash := HashToken(input.RefreshToken)
	stored, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("authService.Refresh: %w", err)
	}
	if stored == nil {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = s.refreshRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, apperror.Unauthorized("refresh token expired")
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("authService.Refresh: %w", err)
	}
	if user == nil || !user.IsActive {
		_ = s.refreshRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, apperror.Unauthorized("user not found or inactive")
	}

	// Token rotation: delete old, issue new.
	if err := s.refreshRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("authService.Refresh: %w", err)
	}

	return s.issueTokenPair(ctx, user)
}

func (s *ServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	tokenHash := HashToken(refreshToken)
	if err := s.refreshRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("authService.Logout: %w", err)
	}
	return nil
}

func (s *ServiceImpl) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.refreshRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("authService.LogoutAll: %w", err)
	}
	return nil
}

func (s *ServiceImpl) Me(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("authService.Me: %w", err)
	}
	if user == nil {
		return nil, apperror.NotFound("User", userID.String())
	}
	return &UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ProfileID: user.ProfileID,
		RoleID:    user.RoleID,
		IsActive:  user.IsActive,
	}, nil
}

func (s *ServiceImpl) SetPassword(ctx context.Context, userID uuid.UUID, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return apperror.Internal("failed to hash password")
	}

	if err := s.userRepo.SetPassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("authService.SetPassword: %w", err)
	}
	return nil
}

func (s *ServiceImpl) ForgotPassword(ctx context.Context, input ForgotPasswordInput) error {
	if err := ValidateForgotPassword(input); err != nil {
		return err
	}

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("authService.ForgotPassword: %w", err)
	}

	// Always return success to not reveal email existence.
	if user == nil || !user.IsActive {
		return nil
	}

	rawToken, err := GenerateToken()
	if err != nil {
		return fmt.Errorf("authService.ForgotPassword: %w", err)
	}

	now := time.Now()
	resetToken := &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: HashToken(rawToken),
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}

	if err := s.resetRepo.Create(ctx, resetToken); err != nil {
		return fmt.Errorf("authService.ForgotPassword: %w", err)
	}

	resetURL := fmt.Sprintf("%s?token=%s", s.resetBaseURL, rawToken)
	if err := s.emailSender.SendPasswordReset(ctx, user.Email, resetURL); err != nil {
		return fmt.Errorf("authService.ForgotPassword: %w", err)
	}

	return nil
}

func (s *ServiceImpl) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	if err := ValidateResetPassword(input); err != nil {
		return err
	}

	tokenHash := HashToken(input.Token)
	stored, err := s.resetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}
	if stored == nil {
		return apperror.BadRequest("invalid or expired reset token")
	}
	if time.Now().After(stored.ExpiresAt) {
		return apperror.BadRequest("reset token has expired")
	}
	if stored.UsedAt != nil {
		return apperror.BadRequest("reset token has already been used")
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return apperror.Internal("failed to hash password")
	}

	if err := s.userRepo.SetPassword(ctx, stored.UserID, hash); err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}

	if err := s.resetRepo.MarkUsed(ctx, stored.ID); err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}

	// Force re-login by deleting all refresh tokens.
	if err := s.refreshRepo.DeleteAllByUserID(ctx, stored.UserID); err != nil {
		return fmt.Errorf("authService.ResetPassword: %w", err)
	}

	return nil
}

func (s *ServiceImpl) issueTokenPair(ctx context.Context, user *UserWithPassword) (*TokenPair, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"pid": user.ProfileID.String(),
		"exp": now.Add(s.accessTTL).Unix(),
		"iat": now.Unix(),
	}
	if user.RoleID != nil {
		claims["rid"] = user.RoleID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("authService.issueTokenPair: %w", err)
	}

	rawRefresh, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("authService.issueTokenPair: %w", err)
	}

	refreshRecord := &RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: HashToken(rawRefresh),
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.refreshRepo.Create(ctx, refreshRecord); err != nil {
		return nil, fmt.Errorf("authService.issueTokenPair: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
