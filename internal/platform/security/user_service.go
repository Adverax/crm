package security

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type userServiceImpl struct {
	txBeginner  TxBeginner
	userRepo    UserRepository
	profileRepo ProfileRepository
	roleRepo    UserRoleRepository
	psToUserRepo PermissionSetToUserRepository
	outboxRepo  OutboxRepository
}

// NewUserService creates a new UserService.
func NewUserService(
	txBeginner TxBeginner,
	userRepo UserRepository,
	profileRepo ProfileRepository,
	roleRepo UserRoleRepository,
	psToUserRepo PermissionSetToUserRepository,
	outboxRepo OutboxRepository,
) UserService {
	return &userServiceImpl{
		txBeginner:   txBeginner,
		userRepo:     userRepo,
		profileRepo:  profileRepo,
		roleRepo:     roleRepo,
		psToUserRepo: psToUserRepo,
		outboxRepo:   outboxRepo,
	}
}

func (s *userServiceImpl) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if err := ValidateCreateUser(input); err != nil {
		return nil, fmt.Errorf("userService.Create: %w", err)
	}

	existing, _ := s.userRepo.GetByUsername(ctx, input.Username)
	if existing != nil {
		return nil, fmt.Errorf("userService.Create: %w",
			apperror.Conflict(fmt.Sprintf("user with username '%s' already exists", input.Username)))
	}

	profile, err := s.profileRepo.GetByID(ctx, input.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("userService.Create: lookup profile: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("userService.Create: %w",
			apperror.NotFound("Profile", input.ProfileID.String()))
	}

	if input.RoleID != nil {
		role, err := s.roleRepo.GetByID(ctx, *input.RoleID)
		if err != nil {
			return nil, fmt.Errorf("userService.Create: lookup role: %w", err)
		}
		if role == nil {
			return nil, fmt.Errorf("userService.Create: %w",
				apperror.NotFound("UserRole", input.RoleID.String()))
		}
	}

	var result *User
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.userRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("userService.Create: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "create"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "user_changed",
			EntityType: "user",
			EntityID:   created.ID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("userService.Create: outbox: %w", err)
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *userServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userService.GetByID: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("userService.GetByID: %w",
			apperror.NotFound("User", id.String()))
	}
	return user, nil
}

func (s *userServiceImpl) List(ctx context.Context, page, perPage int32) ([]User, int64, error) {
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

	users, err := s.userRepo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("userService.List: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("userService.List: count: %w", err)
	}

	return users, total, nil
}

func (s *userServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error) {
	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("userService.Update: %w",
			apperror.NotFound("User", id.String()))
	}

	if err := ValidateUpdateUser(input); err != nil {
		return nil, fmt.Errorf("userService.Update: %w", err)
	}

	profileChanged := existing.ProfileID != input.ProfileID
	roleChanged := !uuidPtrEqual(existing.RoleID, input.RoleID)

	var result *User
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.userRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("userService.Update: %w", err)
		}

		if profileChanged || roleChanged {
			payload, _ := json.Marshal(map[string]string{"action": "update"})
			if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
				EventType:  "user_changed",
				EntityType: "user",
				EntityID:   id,
				Payload:    payload,
			}); err != nil {
				return fmt.Errorf("userService.Update: outbox: %w", err)
			}
		}

		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *userServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("userService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("userService.Delete: %w",
			apperror.NotFound("User", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.userRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("userService.Delete: %w", err)
		}
		return nil
	})
}

func (s *userServiceImpl) AssignPermissionSet(ctx context.Context, userID, psID uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		_, err := s.psToUserRepo.Assign(ctx, tx, psID, userID)
		if err != nil {
			return fmt.Errorf("userService.AssignPermissionSet: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "assign"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("userService.AssignPermissionSet: outbox: %w", err)
		}

		return nil
	})
}

func (s *userServiceImpl) RevokePermissionSet(ctx context.Context, userID, psID uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.psToUserRepo.Revoke(ctx, tx, psID, userID); err != nil {
			return fmt.Errorf("userService.RevokePermissionSet: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "revoke"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("userService.RevokePermissionSet: outbox: %w", err)
		}

		return nil
	})
}

func (s *userServiceImpl) ListPermissionSets(ctx context.Context, userID uuid.UUID) ([]PermissionSetToUser, error) {
	assignments, err := s.psToUserRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("userService.ListPermissionSets: %w", err)
	}
	return assignments, nil
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
