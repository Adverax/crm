package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestPermissionSetService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   security.CreatePermissionSetInput
		setup   func(*mockPSRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates grant PS successfully",
			input: security.CreatePermissionSetInput{
				APIName: "sales_access",
				Label:   "Sales Access",
				PSType:  security.PSTypeGrant,
			},
			wantErr: false,
		},
		{
			name: "creates deny PS successfully",
			input: security.CreatePermissionSetInput{
				APIName: "restrict_delete",
				Label:   "Restrict Delete",
				PSType:  security.PSTypeDeny,
			},
			wantErr: false,
		},
		{
			name: "returns validation error for empty api_name",
			input: security.CreatePermissionSetInput{
				APIName: "",
				Label:   "Bad",
				PSType:  security.PSTypeGrant,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for empty label",
			input: security.CreatePermissionSetInput{
				APIName: "good_name",
				Label:   "",
				PSType:  security.PSTypeGrant,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for invalid ps_type",
			input: security.CreatePermissionSetInput{
				APIName: "good_name",
				Label:   "Good",
				PSType:  "invalid",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns conflict for duplicate api_name",
			input: security.CreatePermissionSetInput{
				APIName: "existing_ps",
				Label:   "Existing",
				PSType:  security.PSTypeGrant,
			},
			setup: func(r *mockPSRepo) {
				r.byName["existing_ps"] = &security.PermissionSet{ID: uuid.New(), APIName: "existing_ps"}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psRepo := newMockPSRepo()
			if tt.setup != nil {
				tt.setup(psRepo)
			}

			svc := security.NewPermissionSetService(&mockTxBeginner{}, psRepo)
			result, err := svc.Create(context.Background(), tt.input)

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
			if result.APIName != tt.input.APIName {
				t.Errorf("expected api_name %s, got %s", tt.input.APIName, result.APIName)
			}
		})
	}
}

func TestPermissionSetService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockPSRepo)
		wantErr bool
	}{
		{
			name: "returns PS when exists",
			id:   existingID,
			setup: func(r *mockPSRepo) {
				r.sets[existingID] = &security.PermissionSet{ID: existingID, APIName: "test_ps"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent PS",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psRepo := newMockPSRepo()
			if tt.setup != nil {
				tt.setup(psRepo)
			}

			svc := security.NewPermissionSetService(&mockTxBeginner{}, psRepo)
			result, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, result.ID)
			}
		})
	}
}

func TestPermissionSetService_List(t *testing.T) {
	psRepo := newMockPSRepo()
	psRepo.sets[uuid.New()] = &security.PermissionSet{ID: uuid.New(), APIName: "ps1"}

	svc := security.NewPermissionSetService(&mockTxBeginner{}, psRepo)
	sets, total, err := svc.List(context.Background(), 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 1 {
		t.Errorf("expected 1 set, got %d", len(sets))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestPermissionSetService_Update(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		input   security.UpdatePermissionSetInput
		setup   func(*mockPSRepo)
		wantErr bool
	}{
		{
			name: "updates PS successfully",
			id:   existingID,
			input: security.UpdatePermissionSetInput{
				Label: "Updated PS",
			},
			setup: func(r *mockPSRepo) {
				r.sets[existingID] = &security.PermissionSet{ID: existingID, APIName: "test_ps", Label: "Original"}
			},
			wantErr: false,
		},
		{
			name: "returns not found for non-existent PS",
			id:   uuid.New(),
			input: security.UpdatePermissionSetInput{
				Label: "Updated",
			},
			wantErr: true,
		},
		{
			name: "returns validation error for empty label",
			id:   existingID,
			input: security.UpdatePermissionSetInput{
				Label: "",
			},
			setup: func(r *mockPSRepo) {
				r.sets[existingID] = &security.PermissionSet{ID: existingID, APIName: "test_ps"}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psRepo := newMockPSRepo()
			if tt.setup != nil {
				tt.setup(psRepo)
			}

			svc := security.NewPermissionSetService(&mockTxBeginner{}, psRepo)
			result, err := svc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Label != tt.input.Label {
				t.Errorf("expected label %s, got %s", tt.input.Label, result.Label)
			}
		})
	}
}

func TestPermissionSetService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockPSRepo)
		wantErr bool
	}{
		{
			name: "deletes PS successfully",
			id:   existingID,
			setup: func(r *mockPSRepo) {
				r.sets[existingID] = &security.PermissionSet{ID: existingID, APIName: "to_delete"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent PS",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psRepo := newMockPSRepo()
			if tt.setup != nil {
				tt.setup(psRepo)
			}

			svc := security.NewPermissionSetService(&mockTxBeginner{}, psRepo)
			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
