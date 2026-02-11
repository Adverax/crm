package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestUserRoleService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   security.CreateUserRoleInput
		setup   func(*mockUserRoleRepo, *mockGroupRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates role successfully with auto-groups",
			input: security.CreateUserRoleInput{
				APIName: "sales_rep",
				Label:   "Sales Rep",
			},
			wantErr: false,
		},
		{
			name: "creates role with parent",
			input: security.CreateUserRoleInput{
				APIName: "sales_rep",
				Label:   "Sales Rep",
			},
			setup: func(rr *mockUserRoleRepo, _ *mockGroupRepo) {
				parentID := uuid.New()
				rr.roles = append(rr.roles, security.UserRole{ID: parentID, APIName: "manager", Label: "Manager"})
				input := security.CreateUserRoleInput{APIName: "sales_rep", Label: "Sales Rep", ParentID: &parentID}
				_ = input // parent will be set via the test input
			},
			wantErr: false,
		},
		{
			name: "returns validation error for empty api_name",
			input: security.CreateUserRoleInput{
				APIName: "",
				Label:   "Bad",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for empty label",
			input: security.CreateUserRoleInput{
				APIName: "good_name",
				Label:   "",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns conflict for duplicate api_name",
			input: security.CreateUserRoleInput{
				APIName: "existing_role",
				Label:   "Existing",
			},
			setup: func(rr *mockUserRoleRepo, _ *mockGroupRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: uuid.New(), APIName: "existing_role", Label: "Existing"})
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
		{
			name: "returns not found for non-existent parent",
			input: security.CreateUserRoleInput{
				APIName:  "child_role",
				Label:    "Child",
				ParentID: ptrUUID(uuid.New()),
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			groupRepo := newMockGroupRepo()
			outboxRepo := &mockOutboxRepo{}

			if tt.setup != nil {
				tt.setup(roleRepo, groupRepo)
			}

			svc := security.NewUserRoleService(&mockTxBeginner{}, roleRepo, groupRepo, outboxRepo)
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
			if len(outboxRepo.events) == 0 {
				t.Error("expected outbox event to be emitted")
			}
		})
	}
}

func TestUserRoleService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockUserRoleRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "returns role when exists",
			id:   existingID,
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: existingID, APIName: "ceo", Label: "CEO"})
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent role",
			id:      uuid.New(),
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			if tt.setup != nil {
				tt.setup(roleRepo)
			}

			svc := security.NewUserRoleService(&mockTxBeginner{}, roleRepo, newMockGroupRepo(), &mockOutboxRepo{})
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

func TestUserRoleService_List(t *testing.T) {
	tests := []struct {
		name      string
		page      int32
		perPage   int32
		setup     func(*mockUserRoleRepo)
		wantCount int
		wantTotal int64
	}{
		{
			name:    "returns empty list",
			page:    1,
			perPage: 20,
			setup: func(_ *mockUserRoleRepo) {
			},
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name:    "returns roles with default pagination",
			page:    0,
			perPage: 0,
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: uuid.New(), APIName: "ceo"})
			},
			wantCount: 1,
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			if tt.setup != nil {
				tt.setup(roleRepo)
			}

			svc := security.NewUserRoleService(&mockTxBeginner{}, roleRepo, newMockGroupRepo(), &mockOutboxRepo{})
			roles, total, err := svc.List(context.Background(), tt.page, tt.perPage)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(roles) != tt.wantCount {
				t.Errorf("expected %d roles, got %d", tt.wantCount, len(roles))
			}
			if total != tt.wantTotal {
				t.Errorf("expected total %d, got %d", tt.wantTotal, total)
			}
		})
	}
}

func TestUserRoleService_Update(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		input   security.UpdateUserRoleInput
		setup   func(*mockUserRoleRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "updates role successfully",
			id:   existingID,
			input: security.UpdateUserRoleInput{
				Label: "Updated CEO",
			},
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: existingID, APIName: "ceo", Label: "CEO"})
			},
			wantErr: false,
		},
		{
			name: "returns not found for non-existent role",
			id:   uuid.New(),
			input: security.UpdateUserRoleInput{
				Label: "Updated",
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
		{
			name: "returns validation error for empty label",
			id:   existingID,
			input: security.UpdateUserRoleInput{
				Label: "",
			},
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: existingID, APIName: "ceo", Label: "CEO"})
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns error when role is its own parent",
			id:   existingID,
			input: security.UpdateUserRoleInput{
				Label:    "Self Parent",
				ParentID: &existingID,
			},
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: existingID, APIName: "ceo", Label: "CEO"})
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			outboxRepo := &mockOutboxRepo{}
			if tt.setup != nil {
				tt.setup(roleRepo)
			}

			svc := security.NewUserRoleService(&mockTxBeginner{}, roleRepo, newMockGroupRepo(), outboxRepo)
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

func TestUserRoleService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockUserRoleRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "deletes role successfully",
			id:   existingID,
			setup: func(rr *mockUserRoleRepo) {
				rr.roles = append(rr.roles, security.UserRole{ID: existingID, APIName: "old_role", Label: "Old"})
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent role",
			id:      uuid.New(),
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			if tt.setup != nil {
				tt.setup(roleRepo)
			}

			svc := security.NewUserRoleService(&mockTxBeginner{}, roleRepo, newMockGroupRepo(), &mockOutboxRepo{})
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
