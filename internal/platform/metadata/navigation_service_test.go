package metadata

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"errors"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// mockNavigationRepo is a test double for NavigationRepository.
type mockNavigationRepo struct {
	createFn       func(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error)
	getByProfileFn func(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error)
	listAllFn      func(ctx context.Context) ([]ProfileNavigation, error)
	updateFn       func(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockNavigationRepo) Create(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error) {
	return m.createFn(ctx, input)
}

func (m *mockNavigationRepo) GetByID(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockNavigationRepo) GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error) {
	return m.getByProfileFn(ctx, profileID)
}

func (m *mockNavigationRepo) ListAll(ctx context.Context) ([]ProfileNavigation, error) {
	return m.listAllFn(ctx)
}

func (m *mockNavigationRepo) Update(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error) {
	return m.updateFn(ctx, id, input)
}

func (m *mockNavigationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}

func validNavConfig() NavConfig {
	return NavConfig{
		Groups: []NavGroup{
			{
				Key:   "sales",
				Label: "Sales",
				Icon:  "briefcase",
				Items: []NavItem{
					{Type: "object", ObjectAPIName: "Account"},
					{Type: "link", Label: "Reports", URL: "/reports", Icon: "bar-chart-2"},
					{Type: "divider"},
				},
			},
		},
	}
}

func TestProfileNavigationService_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name      string
		input     CreateProfileNavigationInput
		mockSetup func(*mockNavigationRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name: "creates navigation successfully",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config:    validNavConfig(),
			},
			mockSetup: func(m *mockNavigationRepo) {
				m.createFn = func(_ context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error) {
					return &ProfileNavigation{
						ID:        uuid.New(),
						ProfileID: input.ProfileID,
						Config:    input.Config,
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when profile_id is nil",
			input: CreateProfileNavigationInput{
				ProfileID: uuid.Nil,
				Config:    validNavConfig(),
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when group key is empty",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{{Key: "", Label: "X"}},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when duplicate group keys",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{
						{Key: "a", Label: "A"},
						{Key: "a", Label: "B"},
					},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when object item missing api_name",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{
						{Key: "g", Label: "G", Items: []NavItem{{Type: "object"}}},
					},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when link item missing url",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{
						{Key: "g", Label: "G", Items: []NavItem{{Type: "link", Label: "X"}}},
					},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when link has javascript URL",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{
						{Key: "g", Label: "G", Items: []NavItem{{Type: "link", Label: "X", URL: "javascript:alert(1)"}}},
					},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when invalid item type",
			input: CreateProfileNavigationInput{
				ProfileID: profileID,
				Config: NavConfig{
					Groups: []NavGroup{
						{Key: "g", Label: "G", Items: []NavItem{{Type: "unknown"}}},
					},
				},
			},
			mockSetup: func(_ *mockNavigationRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockNavigationRepo{}
			tt.mockSetup(repo)
			svc := NewProfileNavigationService(repo)

			result, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errCode != "" {
					var appErr *apperror.AppError
					if errors.As(err, &appErr) && string(appErr.Code) != tt.errCode {
						t.Errorf("expected error code %s, got: %s", tt.errCode, appErr.Code)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
		})
	}
}

func TestProfileNavigationService_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockNavigationRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name: "returns navigation when exists",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return &ProfileNavigation{ID: id}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "returns NotFound when not exists",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return nil, nil
				}
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockNavigationRepo{}
			tt.mockSetup(repo)
			svc := NewProfileNavigationService(repo)

			result, err := svc.GetByID(context.Background(), id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
		})
	}
}

func TestProfileNavigationService_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		input     UpdateProfileNavigationInput
		mockSetup func(*mockNavigationRepo)
		wantErr   bool
	}{
		{
			name:  "updates successfully",
			input: UpdateProfileNavigationInput{Config: validNavConfig()},
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return &ProfileNavigation{ID: id}, nil
				}
				m.updateFn = func(_ context.Context, _ uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error) {
					return &ProfileNavigation{ID: id, Config: input.Config}, nil
				}
			},
			wantErr: false,
		},
		{
			name:  "returns NotFound when not exists",
			input: UpdateProfileNavigationInput{Config: validNavConfig()},
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockNavigationRepo{}
			tt.mockSetup(repo)
			svc := NewProfileNavigationService(repo)

			result, err := svc.Update(context.Background(), id, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
		})
	}
}

func TestProfileNavigationService_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockNavigationRepo)
		wantErr   bool
	}{
		{
			name: "deletes successfully",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return &ProfileNavigation{ID: id}, nil
				}
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "returns NotFound when not exists",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockNavigationRepo{}
			tt.mockSetup(repo)
			svc := NewProfileNavigationService(repo)

			err := svc.Delete(context.Background(), id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestProfileNavigationService_ResolveForProfile(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockNavigationRepo)
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "returns navigation when exists",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByProfileFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return &ProfileNavigation{ProfileID: profileID, Config: validNavConfig()}, nil
				}
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns nil when not found",
			mockSetup: func(m *mockNavigationRepo) {
				m.getByProfileFn = func(_ context.Context, _ uuid.UUID) (*ProfileNavigation, error) {
					return nil, nil
				}
			},
			wantNil: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockNavigationRepo{}
			tt.mockSetup(repo)
			svc := NewProfileNavigationService(repo)

			result, err := svc.ResolveForProfile(context.Background(), profileID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Fatal("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Fatal("expected result, got nil")
			}
		})
	}
}
