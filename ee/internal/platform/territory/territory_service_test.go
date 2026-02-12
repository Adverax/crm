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

// --- Mock TerritoryRepository ---

type mockTerritoryRepo struct {
	territories map[uuid.UUID]*territory.Territory
	byName      map[string]*territory.Territory // key: modelID+apiName
}

func newMockTerritoryRepo() *mockTerritoryRepo {
	return &mockTerritoryRepo{
		territories: make(map[uuid.UUID]*territory.Territory),
		byName:      make(map[string]*territory.Territory),
	}
}

func (r *mockTerritoryRepo) Create(_ context.Context, _ pgx.Tx, input territory.CreateTerritoryInput) (*territory.Territory, error) {
	t := &territory.Territory{
		ID:       uuid.New(),
		ModelID:  input.ModelID,
		ParentID: input.ParentID,
		APIName:  input.APIName,
		Label:    input.Label,
	}
	r.territories[t.ID] = t
	r.byName[input.ModelID.String()+input.APIName] = t
	return t, nil
}

func (r *mockTerritoryRepo) GetByID(_ context.Context, id uuid.UUID) (*territory.Territory, error) {
	return r.territories[id], nil
}

func (r *mockTerritoryRepo) GetByAPIName(_ context.Context, modelID uuid.UUID, apiName string) (*territory.Territory, error) {
	return r.byName[modelID.String()+apiName], nil
}

func (r *mockTerritoryRepo) ListByModelID(_ context.Context, modelID uuid.UUID) ([]territory.Territory, error) {
	result := make([]territory.Territory, 0)
	for _, t := range r.territories {
		if t.ModelID == modelID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *mockTerritoryRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input territory.UpdateTerritoryInput) (*territory.Territory, error) {
	t := r.territories[id]
	if t == nil {
		return nil, errors.New("not found")
	}
	t.ParentID = input.ParentID
	t.Label = input.Label
	t.Description = input.Description
	return t, nil
}

func (r *mockTerritoryRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.territories, id)
	return nil
}

// --- Tests ---

func TestTerritoryService_Create(t *testing.T) {
	modelID := uuid.New()

	tests := []struct {
		name    string
		input   territory.CreateTerritoryInput
		setupMR func(*mockModelRepo)
		setupTR func(*mockTerritoryRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates territory successfully",
			input: territory.CreateTerritoryInput{
				ModelID: modelID,
				APIName: "north",
				Label:   "North",
			},
			setupMR: func(r *mockModelRepo) {
				r.models[modelID] = &territory.TerritoryModel{ID: modelID, APIName: "q1", Status: territory.ModelStatusPlanning}
			},
			wantErr: false,
		},
		{
			name: "creates territory with parent successfully",
			input: territory.CreateTerritoryInput{
				ModelID: modelID,
				APIName: "northeast",
				Label:   "Northeast",
			},
			setupMR: func(r *mockModelRepo) {
				r.models[modelID] = &territory.TerritoryModel{ID: modelID, APIName: "q1", Status: territory.ModelStatusPlanning}
			},
			setupTR: func(r *mockTerritoryRepo) {
				parentID := uuid.New()
				r.territories[parentID] = &territory.Territory{ID: parentID, ModelID: modelID, APIName: "north"}
			},
			wantErr: false,
		},
		{
			name: "returns error when model not found",
			input: territory.CreateTerritoryInput{
				ModelID: uuid.New(),
				APIName: "orphan",
				Label:   "Orphan",
			},
			wantErr: true,
		},
		{
			name: "returns conflict when api_name exists in model",
			input: territory.CreateTerritoryInput{
				ModelID: modelID,
				APIName: "duplicate",
				Label:   "Duplicate",
			},
			setupMR: func(r *mockModelRepo) {
				r.models[modelID] = &territory.TerritoryModel{ID: modelID, APIName: "q1"}
			},
			setupTR: func(r *mockTerritoryRepo) {
				r.byName[modelID.String()+"duplicate"] = &territory.Territory{ID: uuid.New(), APIName: "duplicate"}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
		{
			name: "returns validation error when label is empty",
			input: territory.CreateTerritoryInput{
				ModelID: modelID,
				APIName: "empty_label",
				Label:   "",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRepo := newMockModelRepo()
			territoryRepo := newMockTerritoryRepo()
			if tt.setupMR != nil {
				tt.setupMR(modelRepo)
			}
			if tt.setupTR != nil {
				tt.setupTR(territoryRepo)
			}

			svc := territory.NewTerritoryService(&mockTxBeginner{}, territoryRepo, modelRepo)
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

func TestTerritoryService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockTerritoryRepo)
		wantErr bool
	}{
		{
			name: "returns territory when exists",
			id:   existingID,
			setup: func(r *mockTerritoryRepo) {
				r.territories[existingID] = &territory.Territory{ID: existingID, APIName: "north"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent territory",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			territoryRepo := newMockTerritoryRepo()
			if tt.setup != nil {
				tt.setup(territoryRepo)
			}

			svc := territory.NewTerritoryService(&mockTxBeginner{}, territoryRepo, newMockModelRepo())
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

func TestTerritoryService_Update(t *testing.T) {
	existingID := uuid.New()
	modelID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		input   territory.UpdateTerritoryInput
		setup   func(*mockTerritoryRepo)
		wantErr bool
	}{
		{
			name: "updates territory successfully",
			id:   existingID,
			input: territory.UpdateTerritoryInput{
				Label: "Updated North",
			},
			setup: func(r *mockTerritoryRepo) {
				r.territories[existingID] = &territory.Territory{
					ID: existingID, ModelID: modelID, APIName: "north", Label: "North",
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when self-referencing parent",
			id:   existingID,
			input: territory.UpdateTerritoryInput{
				ParentID: &existingID,
				Label:    "Self Parent",
			},
			setup: func(r *mockTerritoryRepo) {
				r.territories[existingID] = &territory.Territory{
					ID: existingID, ModelID: modelID, APIName: "north",
				}
			},
			wantErr: true,
		},
		{
			name:    "returns not found for non-existent territory",
			id:      uuid.New(),
			input:   territory.UpdateTerritoryInput{Label: "X"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			territoryRepo := newMockTerritoryRepo()
			if tt.setup != nil {
				tt.setup(territoryRepo)
			}

			svc := territory.NewTerritoryService(&mockTxBeginner{}, territoryRepo, newMockModelRepo())
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

func TestTerritoryService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockTerritoryRepo)
		wantErr bool
	}{
		{
			name: "deletes territory successfully",
			id:   existingID,
			setup: func(r *mockTerritoryRepo) {
				r.territories[existingID] = &territory.Territory{ID: existingID, APIName: "north"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent territory",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			territoryRepo := newMockTerritoryRepo()
			if tt.setup != nil {
				tt.setup(territoryRepo)
			}

			svc := territory.NewTerritoryService(&mockTxBeginner{}, territoryRepo, newMockModelRepo())
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
