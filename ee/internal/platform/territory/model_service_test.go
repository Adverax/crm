//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/ee/internal/platform/territory"
	"github.com/adverax/crm/internal/pkg/apperror"
)

// --- Mock helpers ---

type mockTxBeginner struct{}

func (m *mockTxBeginner) Begin(_ context.Context) (pgx.Tx, error) {
	return &mockTx{}, nil
}

type mockTx struct{ pgx.Tx }

func (t *mockTx) Commit(_ context.Context) error   { return nil }
func (t *mockTx) Rollback(_ context.Context) error { return nil }

// --- Mock ModelRepository ---

type mockModelRepo struct {
	models map[uuid.UUID]*territory.TerritoryModel
	byName map[string]*territory.TerritoryModel
}

func newMockModelRepo() *mockModelRepo {
	return &mockModelRepo{
		models: make(map[uuid.UUID]*territory.TerritoryModel),
		byName: make(map[string]*territory.TerritoryModel),
	}
}

func (r *mockModelRepo) Create(_ context.Context, _ pgx.Tx, input territory.CreateModelInput) (*territory.TerritoryModel, error) {
	m := &territory.TerritoryModel{
		ID:      uuid.New(),
		APIName: input.APIName,
		Label:   input.Label,
		Status:  territory.ModelStatusPlanning,
	}
	r.models[m.ID] = m
	r.byName[m.APIName] = m
	return m, nil
}

func (r *mockModelRepo) GetByID(_ context.Context, id uuid.UUID) (*territory.TerritoryModel, error) {
	return r.models[id], nil
}

func (r *mockModelRepo) GetByAPIName(_ context.Context, apiName string) (*territory.TerritoryModel, error) {
	return r.byName[apiName], nil
}

func (r *mockModelRepo) GetActive(_ context.Context) (*territory.TerritoryModel, error) {
	for _, m := range r.models {
		if m.Status == territory.ModelStatusActive {
			return m, nil
		}
	}
	return nil, nil
}

func (r *mockModelRepo) List(_ context.Context, _, _ int32) ([]territory.TerritoryModel, error) {
	result := make([]territory.TerritoryModel, 0, len(r.models))
	for _, m := range r.models {
		result = append(result, *m)
	}
	return result, nil
}

func (r *mockModelRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input territory.UpdateModelInput) (*territory.TerritoryModel, error) {
	m := r.models[id]
	if m == nil {
		return nil, errors.New("not found")
	}
	m.Label = input.Label
	m.Description = input.Description
	return m, nil
}

func (r *mockModelRepo) UpdateStatus(_ context.Context, _ pgx.Tx, id uuid.UUID, status territory.ModelStatus) error {
	m := r.models[id]
	if m == nil {
		return errors.New("not found")
	}
	m.Status = status
	return nil
}

func (r *mockModelRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.models, id)
	return nil
}

func (r *mockModelRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.models)), nil
}

// --- Mock EffectiveRepository ---

type mockEffectiveRepo struct {
	activateCalled  bool
	activateErr     error
	userTerritories map[uuid.UUID][]uuid.UUID
	groupIDs        []uuid.UUID
}

func newMockEffectiveRepo() *mockEffectiveRepo {
	return &mockEffectiveRepo{
		userTerritories: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *mockEffectiveRepo) RebuildHierarchy(_ context.Context, _ pgx.Tx, _ uuid.UUID) error {
	return nil
}

func (r *mockEffectiveRepo) GenerateRecordShareEntries(_ context.Context, _ pgx.Tx, _, _, _ uuid.UUID, _ string) error {
	return nil
}

func (r *mockEffectiveRepo) ActivateModel(_ context.Context, _ pgx.Tx, _ uuid.UUID) error {
	r.activateCalled = true
	return r.activateErr
}

func (r *mockEffectiveRepo) GetUserTerritories(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return r.userTerritories[userID], nil
}

func (r *mockEffectiveRepo) GetTerritoryGroupIDs(_ context.Context, _ []uuid.UUID) ([]uuid.UUID, error) {
	return r.groupIDs, nil
}

// --- Tests ---

func TestModelService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   territory.CreateModelInput
		setup   func(*mockModelRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates model successfully",
			input: territory.CreateModelInput{
				APIName: "q1_2026",
				Label:   "Q1 2026",
			},
			wantErr: false,
		},
		{
			name: "returns error when api_name already exists",
			input: territory.CreateModelInput{
				APIName: "existing",
				Label:   "Existing",
			},
			setup: func(r *mockModelRepo) {
				r.byName["existing"] = &territory.TerritoryModel{ID: uuid.New(), APIName: "existing"}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
		{
			name: "returns validation error when label is empty",
			input: territory.CreateModelInput{
				APIName: "no_label",
				Label:   "",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error when api_name is empty",
			input: territory.CreateModelInput{
				APIName: "",
				Label:   "Some Label",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			if tt.setup != nil {
				tt.setup(modelRepo)
			}

			svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, newMockEffectiveRepo())
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
			if result.Status != territory.ModelStatusPlanning {
				t.Errorf("expected status planning, got %s", result.Status)
			}
		})
	}
}

func TestModelService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockModelRepo)
		wantErr bool
	}{
		{
			name: "returns model when exists",
			id:   existingID,
			setup: func(r *mockModelRepo) {
				r.models[existingID] = &territory.TerritoryModel{ID: existingID, APIName: "test"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent model",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			if tt.setup != nil {
				tt.setup(modelRepo)
			}

			svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, newMockEffectiveRepo())
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

func TestModelService_Activate(t *testing.T) {
	planningID := uuid.New()
	activeID := uuid.New()
	archivedID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockModelRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "activates planning model successfully",
			id:   planningID,
			setup: func(r *mockModelRepo) {
				r.models[planningID] = &territory.TerritoryModel{
					ID: planningID, APIName: "q1", Status: territory.ModelStatusPlanning,
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when model is already active",
			id:   activeID,
			setup: func(r *mockModelRepo) {
				r.models[activeID] = &territory.TerritoryModel{
					ID: activeID, APIName: "q2", Status: territory.ModelStatusActive,
				}
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns error when model is archived",
			id:   archivedID,
			setup: func(r *mockModelRepo) {
				r.models[archivedID] = &territory.TerritoryModel{
					ID: archivedID, APIName: "q3", Status: territory.ModelStatusArchived,
				}
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name:    "returns not found for non-existent model",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			effectiveRepo := newMockEffectiveRepo()
			if tt.setup != nil {
				tt.setup(modelRepo)
			}

			svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, effectiveRepo)
			err := svc.Activate(context.Background(), tt.id)

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
			if !effectiveRepo.activateCalled {
				t.Error("expected ActivateModel to be called on effective repo")
			}
		})
	}
}

func TestModelService_Archive(t *testing.T) {
	activeID := uuid.New()
	planningID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockModelRepo)
		wantErr bool
	}{
		{
			name: "archives active model successfully",
			id:   activeID,
			setup: func(r *mockModelRepo) {
				r.models[activeID] = &territory.TerritoryModel{
					ID: activeID, APIName: "q1", Status: territory.ModelStatusActive,
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when model is in planning",
			id:   planningID,
			setup: func(r *mockModelRepo) {
				r.models[planningID] = &territory.TerritoryModel{
					ID: planningID, APIName: "q2", Status: territory.ModelStatusPlanning,
				}
			},
			wantErr: true,
		},
		{
			name:    "returns not found for non-existent model",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			if tt.setup != nil {
				tt.setup(modelRepo)
			}

			svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, newMockEffectiveRepo())
			err := svc.Archive(context.Background(), tt.id)

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

func TestModelService_Delete(t *testing.T) {
	planningID := uuid.New()
	activeID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockModelRepo)
		wantErr bool
	}{
		{
			name: "deletes planning model",
			id:   planningID,
			setup: func(r *mockModelRepo) {
				r.models[planningID] = &territory.TerritoryModel{
					ID: planningID, APIName: "q1", Status: territory.ModelStatusPlanning,
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when deleting active model",
			id:   activeID,
			setup: func(r *mockModelRepo) {
				r.models[activeID] = &territory.TerritoryModel{
					ID: activeID, APIName: "q2", Status: territory.ModelStatusActive,
				}
			},
			wantErr: true,
		},
		{
			name:    "returns not found for non-existent model",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			if tt.setup != nil {
				tt.setup(modelRepo)
			}

			svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, newMockEffectiveRepo())
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

func TestModelService_List(t *testing.T) {
	modelRepo := newMockModelRepo()
	id := uuid.New()
	modelRepo.models[id] = &territory.TerritoryModel{ID: id, APIName: "test"}

	svc := territory.NewModelService(&mockTxBeginner{}, modelRepo, newMockEffectiveRepo())
	models, total, err := svc.List(context.Background(), 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 1 {
		t.Errorf("expected 1 model, got %d", len(models))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}
