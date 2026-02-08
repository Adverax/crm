package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type userRoleServiceImpl struct {
	txBeginner TxBeginner
	repo       UserRoleRepository
}

// NewUserRoleService creates a new UserRoleService.
func NewUserRoleService(txBeginner TxBeginner, repo UserRoleRepository) UserRoleService {
	return &userRoleServiceImpl{txBeginner: txBeginner, repo: repo}
}

func (s *userRoleServiceImpl) Create(ctx context.Context, input CreateUserRoleInput) (*UserRole, error) {
	if err := ValidateCreateUserRole(input); err != nil {
		return nil, fmt.Errorf("userRoleService.Create: %w", err)
	}

	existing, _ := s.repo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("userRoleService.Create: %w",
			apperror.Conflict(fmt.Sprintf("role with api_name '%s' already exists", input.APIName)))
	}

	if input.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("userRoleService.Create: lookup parent: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("userRoleService.Create: %w",
				apperror.NotFound("UserRole", input.ParentID.String()))
		}
	}

	var result *UserRole
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.repo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("userRoleService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *userRoleServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userRoleService.GetByID: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("userRoleService.GetByID: %w",
			apperror.NotFound("UserRole", id.String()))
	}
	return role, nil
}

func (s *userRoleServiceImpl) List(ctx context.Context, page, perPage int32) ([]UserRole, int64, error) {
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

	roles, err := s.repo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("userRoleService.List: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("userRoleService.List: count: %w", err)
	}

	return roles, total, nil
}

func (s *userRoleServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateUserRoleInput) (*UserRole, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userRoleService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("userRoleService.Update: %w",
			apperror.NotFound("UserRole", id.String()))
	}

	if err := ValidateUpdateUserRole(input); err != nil {
		return nil, fmt.Errorf("userRoleService.Update: %w", err)
	}

	if input.ParentID != nil && *input.ParentID == id {
		return nil, fmt.Errorf("userRoleService.Update: %w",
			apperror.Validation("role cannot be its own parent"))
	}

	var result *UserRole
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.repo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("userRoleService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *userRoleServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("userRoleService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("userRoleService.Delete: %w",
			apperror.NotFound("UserRole", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.repo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("userRoleService.Delete: %w", err)
		}
		return nil
	})
}
