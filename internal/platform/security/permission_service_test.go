package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestPermissionService_SetObjectPermission(t *testing.T) {
	psID := uuid.New()
	objectID := uuid.New()

	tests := []struct {
		name    string
		input   security.SetObjectPermissionInput
		wantErr bool
		errCode string
	}{
		{
			name: "sets object permission successfully",
			input: security.SetObjectPermissionInput{
				ObjectID:    objectID,
				Permissions: 15, // OLSAll
			},
			wantErr: false,
		},
		{
			name: "sets read-only permission",
			input: security.SetObjectPermissionInput{
				ObjectID:    objectID,
				Permissions: 1, // OLSRead
			},
			wantErr: false,
		},
		{
			name: "returns validation error for negative permissions",
			input: security.SetObjectPermissionInput{
				ObjectID:    objectID,
				Permissions: -1,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for permissions > OLSAll",
			input: security.SetObjectPermissionInput{
				ObjectID:    objectID,
				Permissions: 16,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objPermRepo := newMockObjPermRepo()
			fieldPermRepo := newMockFieldPermRepo()
			outboxRepo := &mockOutboxRepo{}

			svc := security.NewPermissionService(&mockTxBeginner{}, objPermRepo, fieldPermRepo, outboxRepo)
			result, err := svc.SetObjectPermission(context.Background(), psID, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errCode != "" {
					var appErr *apperror.AppError
					if errors.As(err, &appErr) {
						if string(appErr.Code) != tt.errCode {
							t.Errorf("expected error code %s, got %s", tt.errCode, appErr.Code)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Permissions != tt.input.Permissions {
				t.Errorf("expected permissions %d, got %d", tt.input.Permissions, result.Permissions)
			}
			if len(outboxRepo.events) == 0 {
				t.Error("expected outbox event to be emitted")
			}
		})
	}
}

func TestPermissionService_ListObjectPermissions(t *testing.T) {
	psID := uuid.New()
	objectID := uuid.New()

	objPermRepo := newMockObjPermRepo()
	objPermRepo.perms[permKey(psID, objectID)] = &security.ObjectPermission{
		ID: uuid.New(), PermissionSetID: psID, ObjectID: objectID, Permissions: 15,
	}

	svc := security.NewPermissionService(&mockTxBeginner{}, objPermRepo, newMockFieldPermRepo(), &mockOutboxRepo{})
	perms, err := svc.ListObjectPermissions(context.Background(), psID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 1 {
		t.Errorf("expected 1 permission, got %d", len(perms))
	}
}

func TestPermissionService_RemoveObjectPermission(t *testing.T) {
	psID := uuid.New()
	objectID := uuid.New()

	objPermRepo := newMockObjPermRepo()
	outboxRepo := &mockOutboxRepo{}

	svc := security.NewPermissionService(&mockTxBeginner{}, objPermRepo, newMockFieldPermRepo(), outboxRepo)
	err := svc.RemoveObjectPermission(context.Background(), psID, objectID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outboxRepo.events) == 0 {
		t.Error("expected outbox event to be emitted")
	}
}

func TestPermissionService_SetFieldPermission(t *testing.T) {
	psID := uuid.New()
	fieldID := uuid.New()

	tests := []struct {
		name    string
		input   security.SetFieldPermissionInput
		wantErr bool
		errCode string
	}{
		{
			name: "sets field permission successfully",
			input: security.SetFieldPermissionInput{
				FieldID:     fieldID,
				Permissions: 3, // FLSAll
			},
			wantErr: false,
		},
		{
			name: "sets read-only field permission",
			input: security.SetFieldPermissionInput{
				FieldID:     fieldID,
				Permissions: 1, // FLSRead
			},
			wantErr: false,
		},
		{
			name: "returns validation error for negative permissions",
			input: security.SetFieldPermissionInput{
				FieldID:     fieldID,
				Permissions: -1,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for permissions > FLSAll",
			input: security.SetFieldPermissionInput{
				FieldID:     fieldID,
				Permissions: 4,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objPermRepo := newMockObjPermRepo()
			fieldPermRepo := newMockFieldPermRepo()
			outboxRepo := &mockOutboxRepo{}

			svc := security.NewPermissionService(&mockTxBeginner{}, objPermRepo, fieldPermRepo, outboxRepo)
			result, err := svc.SetFieldPermission(context.Background(), psID, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errCode != "" {
					var appErr *apperror.AppError
					if errors.As(err, &appErr) {
						if string(appErr.Code) != tt.errCode {
							t.Errorf("expected error code %s, got %s", tt.errCode, appErr.Code)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Permissions != tt.input.Permissions {
				t.Errorf("expected permissions %d, got %d", tt.input.Permissions, result.Permissions)
			}
		})
	}
}

func TestPermissionService_ListFieldPermissions(t *testing.T) {
	psID := uuid.New()
	fieldID := uuid.New()

	fieldPermRepo := newMockFieldPermRepo()
	fieldPermRepo.perms[permKey(psID, fieldID)] = &security.FieldPermission{
		ID: uuid.New(), PermissionSetID: psID, FieldID: fieldID, Permissions: 3,
	}

	svc := security.NewPermissionService(&mockTxBeginner{}, newMockObjPermRepo(), fieldPermRepo, &mockOutboxRepo{})
	perms, err := svc.ListFieldPermissions(context.Background(), psID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 1 {
		t.Errorf("expected 1 permission, got %d", len(perms))
	}
}

func TestPermissionService_RemoveFieldPermission(t *testing.T) {
	psID := uuid.New()
	fieldID := uuid.New()

	outboxRepo := &mockOutboxRepo{}

	svc := security.NewPermissionService(&mockTxBeginner{}, newMockObjPermRepo(), newMockFieldPermRepo(), outboxRepo)
	err := svc.RemoveFieldPermission(context.Background(), psID, fieldID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outboxRepo.events) == 0 {
		t.Error("expected outbox event to be emitted")
	}
}
