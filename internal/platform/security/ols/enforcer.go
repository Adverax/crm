package ols

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// Enforcer checks object-level security permissions.
type Enforcer interface {
	CanRead(ctx context.Context, userID, objectID uuid.UUID) error
	CanCreate(ctx context.Context, userID, objectID uuid.UUID) error
	CanUpdate(ctx context.Context, userID, objectID uuid.UUID) error
	CanDelete(ctx context.Context, userID, objectID uuid.UUID) error
	GetPermissions(ctx context.Context, userID, objectID uuid.UUID) (int, error)
}

type enforcerImpl struct {
	effectiveRepo security.EffectivePermissionRepository
}

// NewEnforcer creates a new OLS Enforcer.
func NewEnforcer(effectiveRepo security.EffectivePermissionRepository) Enforcer {
	return &enforcerImpl{effectiveRepo: effectiveRepo}
}

func (e *enforcerImpl) CanRead(ctx context.Context, userID, objectID uuid.UUID) error {
	perms, err := e.GetPermissions(ctx, userID, objectID)
	if err != nil {
		return fmt.Errorf("olsEnforcer.CanRead: %w", err)
	}
	if !security.HasOLS(perms, security.OLSRead) {
		return fmt.Errorf("olsEnforcer.CanRead: %w",
			apperror.Forbidden("insufficient object-level permissions: read"))
	}
	return nil
}

func (e *enforcerImpl) CanCreate(ctx context.Context, userID, objectID uuid.UUID) error {
	perms, err := e.GetPermissions(ctx, userID, objectID)
	if err != nil {
		return fmt.Errorf("olsEnforcer.CanCreate: %w", err)
	}
	if !security.HasOLS(perms, security.OLSCreate) {
		return fmt.Errorf("olsEnforcer.CanCreate: %w",
			apperror.Forbidden("insufficient object-level permissions: create"))
	}
	return nil
}

func (e *enforcerImpl) CanUpdate(ctx context.Context, userID, objectID uuid.UUID) error {
	perms, err := e.GetPermissions(ctx, userID, objectID)
	if err != nil {
		return fmt.Errorf("olsEnforcer.CanUpdate: %w", err)
	}
	if !security.HasOLS(perms, security.OLSUpdate) {
		return fmt.Errorf("olsEnforcer.CanUpdate: %w",
			apperror.Forbidden("insufficient object-level permissions: update"))
	}
	return nil
}

func (e *enforcerImpl) CanDelete(ctx context.Context, userID, objectID uuid.UUID) error {
	perms, err := e.GetPermissions(ctx, userID, objectID)
	if err != nil {
		return fmt.Errorf("olsEnforcer.CanDelete: %w", err)
	}
	if !security.HasOLS(perms, security.OLSDelete) {
		return fmt.Errorf("olsEnforcer.CanDelete: %w",
			apperror.Forbidden("insufficient object-level permissions: delete"))
	}
	return nil
}

func (e *enforcerImpl) GetPermissions(ctx context.Context, userID, objectID uuid.UUID) (int, error) {
	eff, err := e.effectiveRepo.GetOLS(ctx, userID, objectID)
	if err != nil {
		return 0, fmt.Errorf("olsEnforcer.GetPermissions: %w", err)
	}
	if eff == nil {
		return 0, nil
	}
	return eff.Permissions, nil
}
