package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock SharedLayoutRepository ---

type mockSharedLayoutRepository struct {
	layouts   map[uuid.UUID]*SharedLayout
	byAPIName map[string]*SharedLayout
	refCounts map[string]int // api_name → reference count
}

func newMockSharedLayoutRepo() *mockSharedLayoutRepository {
	return &mockSharedLayoutRepository{
		layouts:   make(map[uuid.UUID]*SharedLayout),
		byAPIName: make(map[string]*SharedLayout),
		refCounts: make(map[string]int),
	}
}

func (m *mockSharedLayoutRepository) Create(_ context.Context, input CreateSharedLayoutInput) (*SharedLayout, error) {
	if _, exists := m.byAPIName[input.APIName]; exists {
		return nil, fmt.Errorf("duplicate api_name: %s", input.APIName)
	}

	now := time.Now()
	config := input.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}
	sl := &SharedLayout{
		ID:        uuid.New(),
		APIName:   input.APIName,
		Type:      input.Type,
		Label:     input.Label,
		Config:    config,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.layouts[sl.ID] = sl
	m.byAPIName[input.APIName] = sl
	return sl, nil
}

func (m *mockSharedLayoutRepository) GetByID(_ context.Context, id uuid.UUID) (*SharedLayout, error) {
	return m.layouts[id], nil
}

func (m *mockSharedLayoutRepository) GetByAPIName(_ context.Context, apiName string) (*SharedLayout, error) {
	return m.byAPIName[apiName], nil
}

func (m *mockSharedLayoutRepository) ListAll(_ context.Context) ([]SharedLayout, error) {
	result := make([]SharedLayout, 0, len(m.layouts))
	for _, sl := range m.layouts {
		result = append(result, *sl)
	}
	return result, nil
}

func (m *mockSharedLayoutRepository) Update(_ context.Context, id uuid.UUID, input UpdateSharedLayoutInput) (*SharedLayout, error) {
	sl := m.layouts[id]
	if sl == nil {
		return nil, nil
	}
	sl.Label = input.Label
	config := input.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}
	sl.Config = config
	sl.UpdatedAt = time.Now()
	return sl, nil
}

func (m *mockSharedLayoutRepository) Delete(_ context.Context, id uuid.UUID) error {
	sl := m.layouts[id]
	if sl != nil {
		delete(m.byAPIName, sl.APIName)
		delete(m.layouts, id)
	}
	return nil
}

func (m *mockSharedLayoutRepository) CountReferences(_ context.Context, apiName string) (int, error) {
	return m.refCounts[apiName], nil
}

// --- Mock CacheLoader for SharedLayout tests ---

type mockSharedLayoutCacheLoader struct {
	sharedLayouts []SharedLayout
}

func (m *mockSharedLayoutCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllAutomationRules(_ context.Context) ([]AutomationRule, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllLayouts(_ context.Context) ([]Layout, error) {
	return nil, nil
}
func (m *mockSharedLayoutCacheLoader) LoadAllSharedLayouts(_ context.Context) ([]SharedLayout, error) {
	return m.sharedLayouts, nil
}
func (m *mockSharedLayoutCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

// --- Test helpers ---

func setupSharedLayoutServiceTest(t *testing.T) (SharedLayoutService, *mockSharedLayoutRepository) {
	t.Helper()

	repo := newMockSharedLayoutRepo()
	loader := &mockSharedLayoutCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewSharedLayoutService(nil, repo, cache)
	return svc, repo
}

func validCreateSharedLayoutInput() CreateSharedLayoutInput {
	return CreateSharedLayoutInput{
		APIName: "phone_field",
		Type:    "field",
		Label:   "Phone Field Layout",
		Config:  json.RawMessage(`{"col_span": 2}`),
	}
}

// --- Tests ---

func TestSharedLayoutService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateSharedLayoutInput
		wantErr bool
		errMsg  string
	}{
		{
			name:  "creates successfully with valid input",
			input: validCreateSharedLayoutInput(),
		},
		{
			name: "creates with section type",
			input: CreateSharedLayoutInput{
				APIName: "address_section",
				Type:    "section",
				Label:   "Address Section",
				Config:  json.RawMessage(`{"columns": 2}`),
			},
		},
		{
			name: "creates with list type",
			input: CreateSharedLayoutInput{
				APIName: "contacts_list",
				Type:    "list",
				Label:   "Contacts List",
			},
		},
		{
			name: "rejects invalid api_name format",
			input: CreateSharedLayoutInput{
				APIName: "Invalid-Name",
				Type:    "field",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name starting with number",
			input: CreateSharedLayoutInput{
				APIName: "1layout",
				Type:    "field",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name longer than 63 characters",
			input: CreateSharedLayoutInput{
				APIName: strings.Repeat("a", 64),
				Type:    "field",
			},
			wantErr: true,
			errMsg:  "api_name must be at most 63 characters",
		},
		{
			name: "rejects invalid type",
			input: CreateSharedLayoutInput{
				APIName: "my_layout",
				Type:    "widget",
			},
			wantErr: true,
			errMsg:  "type must be one of",
		},
		{
			name: "rejects label longer than 255 characters",
			input: CreateSharedLayoutInput{
				APIName: "my_layout",
				Type:    "field",
				Label:   strings.Repeat("x", 256),
			},
			wantErr: true,
			errMsg:  "label must be at most 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupSharedLayoutServiceTest(t)

			sl, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, sl)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, sl)
			assert.NotEqual(t, uuid.Nil, sl.ID)
			assert.Equal(t, tt.input.APIName, sl.APIName)
			assert.Equal(t, tt.input.Type, sl.Type)
		})
	}
}

func TestSharedLayoutService_Create_Duplicate(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, validCreateSharedLayoutInput())
	require.NoError(t, err)

	_, err = svc.Create(ctx, validCreateSharedLayoutInput())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestSharedLayoutService_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc SharedLayoutService) uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing shared layout",
			setup: func(svc SharedLayoutService) uuid.UUID {
				sl, _ := svc.Create(context.Background(), validCreateSharedLayoutInput())
				return sl.ID
			},
		},
		{
			name: "returns not found for nonexistent id",
			setup: func(_ SharedLayoutService) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupSharedLayoutServiceTest(t)
			id := tt.setup(svc)

			sl, err := svc.GetByID(context.Background(), id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, sl)
			assert.Equal(t, id, sl.ID)
		})
	}
}

func TestSharedLayoutService_GetByAPIName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc SharedLayoutService)
		apiName string
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing shared layout by api_name",
			setup: func(svc SharedLayoutService) {
				_, _ = svc.Create(context.Background(), validCreateSharedLayoutInput())
			},
			apiName: "phone_field",
		},
		{
			name:    "returns not found for nonexistent api_name",
			setup:   func(_ SharedLayoutService) {},
			apiName: "nonexistent",
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupSharedLayoutServiceTest(t)
			tt.setup(svc)

			sl, err := svc.GetByAPIName(context.Background(), tt.apiName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, sl)
			assert.Equal(t, tt.apiName, sl.APIName)
		})
	}
}

func TestSharedLayoutService_ListAll(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	layouts, err := svc.ListAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, layouts)

	_, err = svc.Create(ctx, CreateSharedLayoutInput{
		APIName: "layout_a", Type: "field",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateSharedLayoutInput{
		APIName: "layout_b", Type: "section",
	})
	require.NoError(t, err)

	layouts, err = svc.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, layouts, 2)
}

func TestSharedLayoutService_Update(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, validCreateSharedLayoutInput())
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, UpdateSharedLayoutInput{
		Label:  "Updated Label",
		Config: json.RawMessage(`{"col_span": 3}`),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Label", updated.Label)
}

func TestSharedLayoutService_Update_NotFound(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)

	_, err := svc.Update(context.Background(), uuid.New(), UpdateSharedLayoutInput{Label: "X"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedLayoutService_Update_RejectsLongLabel(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, validCreateSharedLayoutInput())
	require.NoError(t, err)

	_, err = svc.Update(ctx, created.ID, UpdateSharedLayoutInput{
		Label: strings.Repeat("x", 256),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label must be at most 255 characters")
}

func TestSharedLayoutService_Delete(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	sl, err := svc.Create(ctx, validCreateSharedLayoutInput())
	require.NoError(t, err)

	err = svc.Delete(ctx, sl.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, sl.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedLayoutService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	svc, _ := setupSharedLayoutServiceTest(t)

	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSharedLayoutService_Delete_ReferencedRestrict(t *testing.T) {
	t.Parallel()
	svc, repo := setupSharedLayoutServiceTest(t)
	ctx := context.Background()

	sl, err := svc.Create(ctx, validCreateSharedLayoutInput())
	require.NoError(t, err)

	// Simulate that this shared layout is referenced by 2 layouts
	repo.refCounts[sl.APIName] = 2

	err = svc.Delete(ctx, sl.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "referenced by 2 layout(s)")
}

func TestSharedLayoutService_ErrorWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operation  func(svc SharedLayoutService) error
		wantPrefix string
	}{
		{
			name: "Create wraps with method name",
			operation: func(svc SharedLayoutService) error {
				_, err := svc.Create(context.Background(), CreateSharedLayoutInput{
					APIName: "Invalid",
					Type:    "field",
				})
				return err
			},
			wantPrefix: "sharedLayoutService.Create:",
		},
		{
			name: "GetByID wraps with method name",
			operation: func(svc SharedLayoutService) error {
				_, err := svc.GetByID(context.Background(), uuid.New())
				return err
			},
			wantPrefix: "sharedLayoutService.GetByID:",
		},
		{
			name: "GetByAPIName wraps with method name",
			operation: func(svc SharedLayoutService) error {
				_, err := svc.GetByAPIName(context.Background(), "nonexistent")
				return err
			},
			wantPrefix: "sharedLayoutService.GetByAPIName:",
		},
		{
			name: "Update wraps with method name",
			operation: func(svc SharedLayoutService) error {
				_, err := svc.Update(context.Background(), uuid.New(), UpdateSharedLayoutInput{Label: "X"})
				return err
			},
			wantPrefix: "sharedLayoutService.Update:",
		},
		{
			name: "Delete wraps with method name",
			operation: func(svc SharedLayoutService) error {
				return svc.Delete(context.Background(), uuid.New())
			},
			wantPrefix: "sharedLayoutService.Delete:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupSharedLayoutServiceTest(t)

			err := tt.operation(svc)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantPrefix)
		})
	}
}
