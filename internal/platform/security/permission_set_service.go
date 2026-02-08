package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type permissionSetServiceImpl struct {
	txBeginner TxBeginner
	repo       PermissionSetRepository
}

// NewPermissionSetService creates a new PermissionSetService.
func NewPermissionSetService(txBeginner TxBeginner, repo PermissionSetRepository) PermissionSetService {
	return &permissionSetServiceImpl{txBeginner: txBeginner, repo: repo}
}

func (s *permissionSetServiceImpl) Create(ctx context.Context, input CreatePermissionSetInput) (*PermissionSet, error) {
	if err := ValidateCreatePermissionSet(input); err != nil {
		return nil, fmt.Errorf("permissionSetService.Create: %w", err)
	}

	existing, _ := s.repo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("permissionSetService.Create: %w",
			apperror.Conflict(fmt.Sprintf("permission set with api_name '%s' already exists", input.APIName)))
	}

	var result *PermissionSet
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.repo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("permissionSetService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionSetServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*PermissionSet, error) {
	ps, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permissionSetService.GetByID: %w", err)
	}
	if ps == nil {
		return nil, fmt.Errorf("permissionSetService.GetByID: %w",
			apperror.NotFound("PermissionSet", id.String()))
	}
	return ps, nil
}

func (s *permissionSetServiceImpl) List(ctx context.Context, page, perPage int32) ([]PermissionSet, int64, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	sets, err := s.repo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("permissionSetService.List: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("permissionSetService.List: count: %w", err)
	}

	return sets, total, nil
}

func (s *permissionSetServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdatePermissionSetInput) (*PermissionSet, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permissionSetService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("permissionSetService.Update: %w",
			apperror.NotFound("PermissionSet", id.String()))
	}

	if err := ValidateUpdatePermissionSet(input); err != nil {
		return nil, fmt.Errorf("permissionSetService.Update: %w", err)
	}

	var result *PermissionSet
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.repo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("permissionSetService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionSetServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("permissionSetService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("permissionSetService.Delete: %w",
			apperror.NotFound("PermissionSet", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.repo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("permissionSetService.Delete: %w", err)
		}
		return nil
	})
}
