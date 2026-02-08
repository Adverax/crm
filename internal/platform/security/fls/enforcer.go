package fls

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// Enforcer checks field-level security permissions.
type Enforcer interface {
	CanReadField(ctx context.Context, userID, fieldID uuid.UUID) error
	CanWriteField(ctx context.Context, userID, fieldID uuid.UUID) error
	GetReadableFields(ctx context.Context, userID, objectID uuid.UUID) ([]string, error)
	GetWritableFields(ctx context.Context, userID, objectID uuid.UUID) ([]string, error)
}

type enforcerImpl struct {
	effectiveRepo security.EffectivePermissionRepository
}

// NewEnforcer creates a new FLS Enforcer.
func NewEnforcer(effectiveRepo security.EffectivePermissionRepository) Enforcer {
	return &enforcerImpl{effectiveRepo: effectiveRepo}
}

func (e *enforcerImpl) CanReadField(ctx context.Context, userID, fieldID uuid.UUID) error {
	eff, err := e.effectiveRepo.GetFLS(ctx, userID, fieldID)
	if err != nil {
		return fmt.Errorf("flsEnforcer.CanReadField: %w", err)
	}
	if eff == nil || !security.HasFLS(eff.Permissions, security.FLSRead) {
		return fmt.Errorf("flsEnforcer.CanReadField: %w",
			apperror.Forbidden("insufficient field-level permissions: read"))
	}
	return nil
}

func (e *enforcerImpl) CanWriteField(ctx context.Context, userID, fieldID uuid.UUID) error {
	eff, err := e.effectiveRepo.GetFLS(ctx, userID, fieldID)
	if err != nil {
		return fmt.Errorf("flsEnforcer.CanWriteField: %w", err)
	}
	if eff == nil || !security.HasFLS(eff.Permissions, security.FLSWrite) {
		return fmt.Errorf("flsEnforcer.CanWriteField: %w",
			apperror.Forbidden("insufficient field-level permissions: write"))
	}
	return nil
}

func (e *enforcerImpl) GetReadableFields(ctx context.Context, userID, objectID uuid.UUID) ([]string, error) {
	fl, err := e.effectiveRepo.GetFieldList(ctx, userID, objectID, security.FLSRead)
	if err != nil {
		return nil, fmt.Errorf("flsEnforcer.GetReadableFields: %w", err)
	}
	if fl == nil {
		return nil, nil
	}
	return fl.FieldNames, nil
}

func (e *enforcerImpl) GetWritableFields(ctx context.Context, userID, objectID uuid.UUID) ([]string, error) {
	fl, err := e.effectiveRepo.GetFieldList(ctx, userID, objectID, security.FLSWrite)
	if err != nil {
		return nil, fmt.Errorf("flsEnforcer.GetWritableFields: %w", err)
	}
	if fl == nil {
		return nil, nil
	}
	return fl.FieldNames, nil
}
