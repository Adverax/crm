package security

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type permissionServiceImpl struct {
	txBeginner    TxBeginner
	objectPermRepo ObjectPermissionRepository
	fieldPermRepo  FieldPermissionRepository
	outboxRepo     OutboxRepository
}

// NewPermissionService creates a new PermissionService.
func NewPermissionService(
	txBeginner TxBeginner,
	objectPermRepo ObjectPermissionRepository,
	fieldPermRepo FieldPermissionRepository,
	outboxRepo OutboxRepository,
) PermissionService {
	return &permissionServiceImpl{
		txBeginner:     txBeginner,
		objectPermRepo: objectPermRepo,
		fieldPermRepo:  fieldPermRepo,
		outboxRepo:     outboxRepo,
	}
}

func (s *permissionServiceImpl) SetObjectPermission(ctx context.Context, psID uuid.UUID, input SetObjectPermissionInput) (*ObjectPermission, error) {
	if err := ValidateSetObjectPermission(input); err != nil {
		return nil, fmt.Errorf("permissionService.SetObjectPermission: %w", err)
	}

	var result *ObjectPermission
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		op, err := s.objectPermRepo.Upsert(ctx, tx, psID, input.ObjectID, input.Permissions)
		if err != nil {
			return fmt.Errorf("permissionService.SetObjectPermission: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "set_object_permission"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("permissionService.SetObjectPermission: outbox: %w", err)
		}

		result = op
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionServiceImpl) ListObjectPermissions(ctx context.Context, psID uuid.UUID) ([]ObjectPermission, error) {
	perms, err := s.objectPermRepo.ListByPermissionSetID(ctx, psID)
	if err != nil {
		return nil, fmt.Errorf("permissionService.ListObjectPermissions: %w", err)
	}
	return perms, nil
}

func (s *permissionServiceImpl) RemoveObjectPermission(ctx context.Context, psID, objectID uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.objectPermRepo.Delete(ctx, tx, psID, objectID); err != nil {
			return fmt.Errorf("permissionService.RemoveObjectPermission: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "remove_object_permission"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("permissionService.RemoveObjectPermission: outbox: %w", err)
		}

		return nil
	})
}

func (s *permissionServiceImpl) SetFieldPermission(ctx context.Context, psID uuid.UUID, input SetFieldPermissionInput) (*FieldPermission, error) {
	if err := ValidateSetFieldPermission(input); err != nil {
		return nil, fmt.Errorf("permissionService.SetFieldPermission: %w", err)
	}

	var result *FieldPermission
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		fp, err := s.fieldPermRepo.Upsert(ctx, tx, psID, input.FieldID, input.Permissions)
		if err != nil {
			return fmt.Errorf("permissionService.SetFieldPermission: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "set_field_permission"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("permissionService.SetFieldPermission: outbox: %w", err)
		}

		result = fp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionServiceImpl) ListFieldPermissions(ctx context.Context, psID uuid.UUID) ([]FieldPermission, error) {
	perms, err := s.fieldPermRepo.ListByPermissionSetID(ctx, psID)
	if err != nil {
		return nil, fmt.Errorf("permissionService.ListFieldPermissions: %w", err)
	}
	return perms, nil
}

func (s *permissionServiceImpl) RemoveFieldPermission(ctx context.Context, psID, fieldID uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.fieldPermRepo.Delete(ctx, tx, psID, fieldID); err != nil {
			return fmt.Errorf("permissionService.RemoveFieldPermission: %w", err)
		}

		payload, _ := json.Marshal(map[string]string{"action": "remove_field_permission"})
		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "permission_set_changed",
			EntityType: "permission_set",
			EntityID:   psID,
			Payload:    payload,
		}); err != nil {
			return fmt.Errorf("permissionService.RemoveFieldPermission: outbox: %w", err)
		}

		return nil
	})
}
