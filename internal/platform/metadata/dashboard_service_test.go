package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// mockDashboardRepo is a test double for DashboardRepository.
type mockDashboardRepo struct {
	createFn       func(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error)
	getByProfileFn func(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error)
	listAllFn      func(ctx context.Context) ([]ProfileDashboard, error)
	updateFn       func(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockDashboardRepo) Create(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error) {
	return m.createFn(ctx, input)
}

func (m *mockDashboardRepo) GetByID(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockDashboardRepo) GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error) {
	return m.getByProfileFn(ctx, profileID)
}

func (m *mockDashboardRepo) ListAll(ctx context.Context) ([]ProfileDashboard, error) {
	return m.listAllFn(ctx)
}

func (m *mockDashboardRepo) Update(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error) {
	return m.updateFn(ctx, id, input)
}

func (m *mockDashboardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}

func validDashboardConfig() DashboardConfig {
	return DashboardConfig{
		Widgets: []DashboardWidget{
			{
				Key:           "tasks",
				Type:          "list",
				Label:         "My Tasks",
				Size:          "half",
				Query:         "SELECT Id, subject FROM Task LIMIT 10",
				Columns:       []string{"subject"},
				ObjectAPIName: "Task",
			},
			{
				Key:    "deals",
				Type:   "metric",
				Label:  "Deals Count",
				Size:   "third",
				Query:  "SELECT COUNT(Id) FROM Opportunity",
				Format: "number",
			},
			{
				Key:   "quick",
				Type:  "link_list",
				Label: "Quick Actions",
				Size:  "third",
				Links: []DashLink{
					{Label: "New Account", URL: "/app/Account/new", Icon: "building"},
				},
			},
		},
	}
}

func TestProfileDashboardService_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name      string
		input     CreateProfileDashboardInput
		mockSetup func(*mockDashboardRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name: "creates dashboard successfully",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config:    validDashboardConfig(),
			},
			mockSetup: func(m *mockDashboardRepo) {
				m.createFn = func(_ context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error) {
					return &ProfileDashboard{
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
			input: CreateProfileDashboardInput{
				ProfileID: uuid.Nil,
				Config:    validDashboardConfig(),
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when widget key is empty",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "", Type: "metric", Label: "X", Query: "SELECT 1"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when duplicate widget keys",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{
						{Key: "a", Type: "metric", Label: "A", Query: "SELECT 1"},
						{Key: "a", Type: "metric", Label: "B", Query: "SELECT 1"},
					},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when invalid widget type",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "unknown", Label: "W"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when list widget missing columns",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "list", Label: "W", Query: "SELECT 1", ObjectAPIName: "Task"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when metric widget missing query",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "metric", Label: "W"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when link_list widget missing links",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "link_list", Label: "W"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when invalid size",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "metric", Label: "W", Query: "SELECT 1", Size: "giant"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
		{
			name: "returns error when invalid format",
			input: CreateProfileDashboardInput{
				ProfileID: profileID,
				Config: DashboardConfig{
					Widgets: []DashboardWidget{{Key: "w", Type: "metric", Label: "W", Query: "SELECT 1", Format: "invalid"}},
				},
			},
			mockSetup: func(_ *mockDashboardRepo) {},
			wantErr:   true,
			errCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockDashboardRepo{}
			tt.mockSetup(repo)
			svc := NewProfileDashboardService(repo)

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

func TestProfileDashboardService_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockDashboardRepo)
		wantErr   bool
	}{
		{
			name: "returns dashboard when exists",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return &ProfileDashboard{ID: id}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "returns NotFound when not exists",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockDashboardRepo{}
			tt.mockSetup(repo)
			svc := NewProfileDashboardService(repo)

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

func TestProfileDashboardService_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		input     UpdateProfileDashboardInput
		mockSetup func(*mockDashboardRepo)
		wantErr   bool
	}{
		{
			name:  "updates successfully",
			input: UpdateProfileDashboardInput{Config: validDashboardConfig()},
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return &ProfileDashboard{ID: id}, nil
				}
				m.updateFn = func(_ context.Context, _ uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error) {
					return &ProfileDashboard{ID: id, Config: input.Config}, nil
				}
			},
			wantErr: false,
		},
		{
			name:  "returns NotFound when not exists",
			input: UpdateProfileDashboardInput{Config: validDashboardConfig()},
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockDashboardRepo{}
			tt.mockSetup(repo)
			svc := NewProfileDashboardService(repo)

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

func TestProfileDashboardService_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockDashboardRepo)
		wantErr   bool
	}{
		{
			name: "deletes successfully",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return &ProfileDashboard{ID: id}, nil
				}
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "returns NotFound when not exists",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockDashboardRepo{}
			tt.mockSetup(repo)
			svc := NewProfileDashboardService(repo)

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

func TestProfileDashboardService_ResolveForProfile(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(*mockDashboardRepo)
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "returns dashboard when exists",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByProfileFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
					return &ProfileDashboard{ProfileID: profileID, Config: validDashboardConfig()}, nil
				}
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns nil when not found",
			mockSetup: func(m *mockDashboardRepo) {
				m.getByProfileFn = func(_ context.Context, _ uuid.UUID) (*ProfileDashboard, error) {
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
			repo := &mockDashboardRepo{}
			tt.mockSetup(repo)
			svc := NewProfileDashboardService(repo)

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
