package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestProfileService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   security.CreateProfileInput
		setup   func(*mockProfileRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates profile with auto base PS",
			input: security.CreateProfileInput{
				APIName: "sales_profile",
				Label:   "Sales Profile",
			},
			wantErr: false,
		},
		{
			name: "returns validation error for empty api_name",
			input: security.CreateProfileInput{
				APIName: "",
				Label:   "Bad",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for empty label",
			input: security.CreateProfileInput{
				APIName: "good_name",
				Label:   "",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns conflict for duplicate api_name",
			input: security.CreateProfileInput{
				APIName: "existing_profile",
				Label:   "Existing",
			},
			setup: func(r *mockProfileRepo) {
				r.byName["existing_profile"] = &security.Profile{ID: uuid.New(), APIName: "existing_profile"}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := newMockProfileRepo()
			psRepo := newMockPSRepo()
			if tt.setup != nil {
				tt.setup(profileRepo)
			}

			svc := security.NewProfileService(&mockTxBeginner{}, profileRepo, psRepo)
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
			if result.BasePermissionSetID == uuid.Nil {
				t.Error("expected base permission set to be created")
			}
		})
	}
}

func TestProfileService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockProfileRepo)
		wantErr bool
	}{
		{
			name: "returns profile when exists",
			id:   existingID,
			setup: func(r *mockProfileRepo) {
				r.profiles[existingID] = &security.Profile{ID: existingID, APIName: "test"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent profile",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := newMockProfileRepo()
			if tt.setup != nil {
				tt.setup(profileRepo)
			}

			svc := security.NewProfileService(&mockTxBeginner{}, profileRepo, newMockPSRepo())
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

func TestProfileService_List(t *testing.T) {
	profileRepo := newMockProfileRepo()
	id := uuid.New()
	profileRepo.profiles[id] = &security.Profile{ID: id, APIName: "admin"}

	svc := security.NewProfileService(&mockTxBeginner{}, profileRepo, newMockPSRepo())
	profiles, total, err := svc.List(context.Background(), 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestProfileService_Update(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		input   security.UpdateProfileInput
		setup   func(*mockProfileRepo)
		wantErr bool
	}{
		{
			name: "updates profile successfully",
			id:   existingID,
			input: security.UpdateProfileInput{
				Label: "Updated Profile",
			},
			setup: func(r *mockProfileRepo) {
				r.profiles[existingID] = &security.Profile{ID: existingID, APIName: "test", Label: "Original"}
			},
			wantErr: false,
		},
		{
			name: "returns not found for non-existent profile",
			id:   uuid.New(),
			input: security.UpdateProfileInput{
				Label: "Updated",
			},
			wantErr: true,
		},
		{
			name: "returns validation error for empty label",
			id:   existingID,
			input: security.UpdateProfileInput{
				Label: "",
			},
			setup: func(r *mockProfileRepo) {
				r.profiles[existingID] = &security.Profile{ID: existingID, APIName: "test"}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := newMockProfileRepo()
			if tt.setup != nil {
				tt.setup(profileRepo)
			}

			svc := security.NewProfileService(&mockTxBeginner{}, profileRepo, newMockPSRepo())
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

func TestProfileService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockProfileRepo)
		wantErr bool
	}{
		{
			name: "deletes profile successfully",
			id:   existingID,
			setup: func(r *mockProfileRepo) {
				r.profiles[existingID] = &security.Profile{ID: existingID, APIName: "to_delete"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent profile",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := newMockProfileRepo()
			if tt.setup != nil {
				tt.setup(profileRepo)
			}

			svc := security.NewProfileService(&mockTxBeginner{}, profileRepo, newMockPSRepo())
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
