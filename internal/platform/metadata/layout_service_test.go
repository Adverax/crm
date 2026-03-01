package metadata

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock LayoutRepository ---

type mockLayoutRepository struct {
	layouts    map[uuid.UUID]*Layout
	byOVAndKey map[string]*Layout // key: ovID|ff|mode
}

func newMockLayoutRepo() *mockLayoutRepository {
	return &mockLayoutRepository{
		layouts:    make(map[uuid.UUID]*Layout),
		byOVAndKey: make(map[string]*Layout),
	}
}

func layoutKey(ovID uuid.UUID, ff, mode string) string {
	return ovID.String() + "|" + ff + "|" + mode
}

func (m *mockLayoutRepository) Create(_ context.Context, input CreateLayoutInput) (*Layout, error) {
	key := layoutKey(input.ObjectViewID, input.FormFactor, input.Mode)
	if _, exists := m.byOVAndKey[key]; exists {
		return nil, &layoutDuplicateError{key: key}
	}

	now := time.Now()
	l := &Layout{
		ID:           uuid.New(),
		ObjectViewID: input.ObjectViewID,
		FormFactor:   input.FormFactor,
		Mode:         input.Mode,
		Config:       input.Config,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.layouts[l.ID] = l
	m.byOVAndKey[key] = l
	return l, nil
}

func (m *mockLayoutRepository) GetByID(_ context.Context, id uuid.UUID) (*Layout, error) {
	return m.layouts[id], nil
}

func (m *mockLayoutRepository) ListByObjectViewID(_ context.Context, ovID uuid.UUID) ([]Layout, error) {
	var result []Layout
	for _, l := range m.layouts {
		if l.ObjectViewID == ovID {
			result = append(result, *l)
		}
	}
	return result, nil
}

func (m *mockLayoutRepository) ListAll(_ context.Context) ([]Layout, error) {
	result := make([]Layout, 0, len(m.layouts))
	for _, l := range m.layouts {
		result = append(result, *l)
	}
	return result, nil
}

func (m *mockLayoutRepository) Update(_ context.Context, id uuid.UUID, input UpdateLayoutInput) (*Layout, error) {
	l := m.layouts[id]
	if l == nil {
		return nil, nil
	}
	l.Config = input.Config
	l.UpdatedAt = time.Now()
	return l, nil
}

func (m *mockLayoutRepository) Delete(_ context.Context, id uuid.UUID) error {
	l := m.layouts[id]
	if l != nil {
		key := layoutKey(l.ObjectViewID, l.FormFactor, l.Mode)
		delete(m.byOVAndKey, key)
		delete(m.layouts, id)
	}
	return nil
}

type layoutDuplicateError struct {
	key string
}

func (e *layoutDuplicateError) Error() string {
	return "duplicate layout: " + e.key
}

// --- Mock CacheLoader for Layout tests ---

type mockLayoutCacheLoader struct {
	objectViews []ObjectView
	layouts     []Layout
}

func (m *mockLayoutCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return m.objectViews, nil
}
func (m *mockLayoutCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllAutomationRules(_ context.Context) ([]AutomationRule, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) LoadAllLayouts(_ context.Context) ([]Layout, error) {
	return m.layouts, nil
}
func (m *mockLayoutCacheLoader) LoadAllSharedLayouts(_ context.Context) ([]SharedLayout, error) {
	return nil, nil
}
func (m *mockLayoutCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

// --- Test helpers ---

func setupLayoutServiceTest(t *testing.T) (LayoutService, *mockLayoutRepository) {
	t.Helper()

	repo := newMockLayoutRepo()
	loader := &mockLayoutCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	// LayoutService needs pool for OV existence check; in unit tests we skip that
	// by using nil pool (the mock repo handles creation without OV check).
	// For proper unit testing, we test validation separately.
	svc := NewLayoutService(nil, repo, cache)
	return svc, repo
}

func validCreateLayoutInput() CreateLayoutInput {
	return CreateLayoutInput{
		ObjectViewID: uuid.New(),
		FormFactor:   "desktop",
		Mode:         "read",
		Config: LayoutConfig{
			Root: &LayoutComponent{
				Type:    "grid",
				Columns: 2,
			},
		},
	}
}

// --- Tests ---

func TestLayoutService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateLayoutInput
		wantErr bool
		errMsg  string
	}{
		{
			name:  "creates successfully with valid input",
			input: validCreateLayoutInput(),
		},
		{
			name: "creates with tablet form factor",
			input: CreateLayoutInput{
				ObjectViewID: uuid.New(),
				FormFactor:   "tablet",
				Mode:         "view",
			},
		},
		{
			name: "creates with mobile form factor",
			input: CreateLayoutInput{
				ObjectViewID: uuid.New(),
				FormFactor:   "mobile",
				Mode:         "read",
			},
		},
		{
			name: "rejects invalid form_factor",
			input: CreateLayoutInput{
				ObjectViewID: uuid.New(),
				FormFactor:   "laptop",
				Mode:         "read",
			},
			wantErr: true,
			errMsg:  "form_factor must be one of",
		},
		{
			name: "rejects invalid mode",
			input: CreateLayoutInput{
				ObjectViewID: uuid.New(),
				FormFactor:   "desktop",
				Mode:         "preview",
			},
			wantErr: true,
			errMsg:  "mode must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupLayoutServiceTest(t)

			layout, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, layout)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, layout)
			assert.NotEqual(t, uuid.Nil, layout.ID)
			assert.Equal(t, tt.input.FormFactor, layout.FormFactor)
			assert.Equal(t, tt.input.Mode, layout.Mode)
		})
	}
}

func TestLayoutService_Create_Duplicate(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)
	ctx := context.Background()

	input := validCreateLayoutInput()
	_, err := svc.Create(ctx, input)
	require.NoError(t, err)

	// Same OV + form_factor + mode = duplicate
	_, err = svc.Create(ctx, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestLayoutService_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc LayoutService) uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing layout",
			setup: func(svc LayoutService) uuid.UUID {
				l, _ := svc.Create(context.Background(), validCreateLayoutInput())
				return l.ID
			},
		},
		{
			name: "returns not found for nonexistent id",
			setup: func(_ LayoutService) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupLayoutServiceTest(t)
			id := tt.setup(svc)

			layout, err := svc.GetByID(context.Background(), id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, layout)
			assert.Equal(t, id, layout.ID)
		})
	}
}

func TestLayoutService_ListByObjectViewID(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)
	ctx := context.Background()

	ovID := uuid.New()
	otherOVID := uuid.New()

	_, err := svc.Create(ctx, CreateLayoutInput{
		ObjectViewID: ovID, FormFactor: "desktop", Mode: "read",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateLayoutInput{
		ObjectViewID: ovID, FormFactor: "desktop", Mode: "view",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateLayoutInput{
		ObjectViewID: otherOVID, FormFactor: "desktop", Mode: "read",
	})
	require.NoError(t, err)

	layouts, err := svc.ListByObjectViewID(ctx, ovID)
	require.NoError(t, err)
	assert.Len(t, layouts, 2)

	layouts, err = svc.ListByObjectViewID(ctx, otherOVID)
	require.NoError(t, err)
	assert.Len(t, layouts, 1)
}

func TestLayoutService_ListAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(svc LayoutService)
		wantCount int
	}{
		{
			name:      "returns empty list when no layouts exist",
			setup:     func(_ LayoutService) {},
			wantCount: 0,
		},
		{
			name: "returns all layouts",
			setup: func(svc LayoutService) {
				ctx := context.Background()
				ovID := uuid.New()
				_, _ = svc.Create(ctx, CreateLayoutInput{
					ObjectViewID: ovID, FormFactor: "desktop", Mode: "read",
				})
				_, _ = svc.Create(ctx, CreateLayoutInput{
					ObjectViewID: ovID, FormFactor: "desktop", Mode: "view",
				})
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupLayoutServiceTest(t)
			tt.setup(svc)

			layouts, err := svc.ListAll(context.Background())
			require.NoError(t, err)
			assert.Len(t, layouts, tt.wantCount)
		})
	}
}

func TestLayoutService_Update(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, validCreateLayoutInput())
	require.NoError(t, err)

	newConfig := LayoutConfig{
		Root: &LayoutComponent{
			Type:    "grid",
			Columns: 3,
		},
	}

	updated, err := svc.Update(ctx, created.ID, UpdateLayoutInput{Config: newConfig})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, 3, updated.Config.Root.Columns)
}

func TestLayoutService_Update_NotFound(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)

	_, err := svc.Update(context.Background(), uuid.New(), UpdateLayoutInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLayoutService_Delete(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)
	ctx := context.Background()

	layout, err := svc.Create(ctx, validCreateLayoutInput())
	require.NoError(t, err)

	err = svc.Delete(ctx, layout.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, layout.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLayoutService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	svc, _ := setupLayoutServiceTest(t)

	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLayoutService_ErrorWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operation  func(svc LayoutService) error
		wantPrefix string
	}{
		{
			name: "Create wraps with method name",
			operation: func(svc LayoutService) error {
				_, err := svc.Create(context.Background(), CreateLayoutInput{
					FormFactor: "invalid",
					Mode:       "read",
				})
				return err
			},
			wantPrefix: "layoutService.Create:",
		},
		{
			name: "GetByID wraps with method name",
			operation: func(svc LayoutService) error {
				_, err := svc.GetByID(context.Background(), uuid.New())
				return err
			},
			wantPrefix: "layoutService.GetByID:",
		},
		{
			name: "Update wraps with method name",
			operation: func(svc LayoutService) error {
				_, err := svc.Update(context.Background(), uuid.New(), UpdateLayoutInput{})
				return err
			},
			wantPrefix: "layoutService.Update:",
		},
		{
			name: "Delete wraps with method name",
			operation: func(svc LayoutService) error {
				return svc.Delete(context.Background(), uuid.New())
			},
			wantPrefix: "layoutService.Delete:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := setupLayoutServiceTest(t)

			err := tt.operation(svc)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantPrefix)
		})
	}
}

// Unused variable check
var _ = strings.Repeat
