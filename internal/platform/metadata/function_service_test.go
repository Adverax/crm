package metadata

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFunctionRepository implements FunctionRepository for testing.
type mockFunctionRepository struct {
	functions map[uuid.UUID]*Function
	byName    map[string]*Function
}

func newMockFunctionRepo() *mockFunctionRepository {
	return &mockFunctionRepository{
		functions: make(map[uuid.UUID]*Function),
		byName:    make(map[string]*Function),
	}
}

func (m *mockFunctionRepository) Create(_ context.Context, input CreateFunctionInput) (*Function, error) {
	fn := &Function{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Params:      input.Params,
		ReturnType:  input.ReturnType,
		Body:        input.Body,
	}
	m.functions[fn.ID] = fn
	m.byName[fn.Name] = fn
	return fn, nil
}

func (m *mockFunctionRepository) GetByID(_ context.Context, id uuid.UUID) (*Function, error) {
	fn := m.functions[id]
	return fn, nil
}

func (m *mockFunctionRepository) GetByName(_ context.Context, name string) (*Function, error) {
	fn := m.byName[name]
	return fn, nil
}

func (m *mockFunctionRepository) ListAll(_ context.Context) ([]Function, error) {
	var result []Function
	for _, fn := range m.functions {
		result = append(result, *fn)
	}
	return result, nil
}

func (m *mockFunctionRepository) Update(_ context.Context, id uuid.UUID, input UpdateFunctionInput) (*Function, error) {
	fn := m.functions[id]
	if fn == nil {
		return nil, nil
	}
	fn.Description = input.Description
	fn.Params = input.Params
	fn.ReturnType = input.ReturnType
	fn.Body = input.Body
	return fn, nil
}

func (m *mockFunctionRepository) Delete(_ context.Context, id uuid.UUID) error {
	fn := m.functions[id]
	if fn != nil {
		delete(m.byName, fn.Name)
		delete(m.functions, id)
	}
	return nil
}

func (m *mockFunctionRepository) Count(_ context.Context) (int, error) {
	return len(m.functions), nil
}

// mockFnCacheLoader implements CacheLoader for function service tests.
type mockFnCacheLoader struct {
	functions []Function
}

func (m *mockFnCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) {
	return m.functions, nil
}
func (m *mockFnCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) LoadAllAutomationRules(_ context.Context) ([]AutomationRule, error) {
	return nil, nil
}
func (m *mockFnCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func setupFnServiceTest(t *testing.T) (FunctionService, *mockFunctionRepository, *mockFnCacheLoader) {
	t.Helper()
	repo := newMockFunctionRepo()
	loader := &mockFnCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewFunctionService(nil, repo, cache, nil)
	return svc, repo, loader
}

func TestFunctionService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateFunctionInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "creates successfully",
			input: CreateFunctionInput{
				Name:       "double",
				Body:       "x * 2",
				ReturnType: "number",
				Params:     []FunctionParam{{Name: "x", Type: "number"}},
			},
		},
		{
			name: "defaults return_type to any",
			input: CreateFunctionInput{
				Name: "identity",
				Body: "x",
			},
		},
		{
			name: "invalid name uppercase",
			input: CreateFunctionInput{
				Name: "Double",
				Body: "x * 2",
			},
			wantErr: true,
			errMsg:  "name must match",
		},
		{
			name: "invalid name starts with number",
			input: CreateFunctionInput{
				Name: "2double",
				Body: "x * 2",
			},
			wantErr: true,
			errMsg:  "name must match",
		},
		{
			name: "empty body",
			input: CreateFunctionInput{
				Name: "empty",
				Body: "",
			},
			wantErr: true,
			errMsg:  "body is required",
		},
		{
			name: "invalid return type",
			input: CreateFunctionInput{
				Name:       "bad",
				Body:       "x",
				ReturnType: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid return_type",
		},
		{
			name: "duplicate param names",
			input: CreateFunctionInput{
				Name: "dup",
				Body: "x + x",
				Params: []FunctionParam{
					{Name: "x", Type: "number"},
					{Name: "x", Type: "string"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate parameter",
		},
		{
			name: "too many params",
			input: CreateFunctionInput{
				Name: "many",
				Body: "x",
				Params: []FunctionParam{
					{Name: "a", Type: "any"}, {Name: "b", Type: "any"},
					{Name: "c", Type: "any"}, {Name: "d", Type: "any"},
					{Name: "e", Type: "any"}, {Name: "f", Type: "any"},
					{Name: "g", Type: "any"}, {Name: "h", Type: "any"},
					{Name: "i", Type: "any"}, {Name: "j", Type: "any"},
					{Name: "k", Type: "any"},
				},
			},
			wantErr: true,
			errMsg:  "at most 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := setupFnServiceTest(t)
			fn, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, fn.ID)
			assert.Equal(t, tt.input.Name, fn.Name)
		})
	}
}

func TestFunctionService_Create_DuplicateName(t *testing.T) {
	svc, _, _ := setupFnServiceTest(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, CreateFunctionInput{
		Name: "double", Body: "x * 2",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateFunctionInput{
		Name: "double", Body: "x * 3",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFunctionService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(svc FunctionService) uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing function",
			setup: func(svc FunctionService) uuid.UUID {
				fn, _ := svc.Create(context.Background(), CreateFunctionInput{
					Name: "test", Body: "x",
				})
				return fn.ID
			},
		},
		{
			name: "returns not found for nonexistent",
			setup: func(_ FunctionService) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := setupFnServiceTest(t)
			id := tt.setup(svc)

			fn, err := svc.GetByID(context.Background(), id)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, id, fn.ID)
		})
	}
}

func TestFunctionService_Update(t *testing.T) {
	tests := []struct {
		name    string
		input   UpdateFunctionInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "updates successfully",
			input: UpdateFunctionInput{
				Body:       "x * 3",
				ReturnType: "number",
			},
		},
		{
			name: "rejects empty body",
			input: UpdateFunctionInput{
				Body: "",
			},
			wantErr: true,
			errMsg:  "body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := setupFnServiceTest(t)
			created, err := svc.Create(context.Background(), CreateFunctionInput{
				Name: "test", Body: "x * 2",
			})
			require.NoError(t, err)

			fn, err := svc.Update(context.Background(), created.ID, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input.Body, fn.Body)
		})
	}
}

func TestFunctionService_Delete(t *testing.T) {
	svc, _, _ := setupFnServiceTest(t)
	ctx := context.Background()

	fn, err := svc.Create(ctx, CreateFunctionInput{
		Name: "to_delete", Body: "x",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, fn.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, fn.ID)
	assert.Error(t, err)
}

func TestFunctionService_Delete_NotFound(t *testing.T) {
	svc, _, _ := setupFnServiceTest(t)
	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFunctionService_ListAll(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc FunctionService)
		wantCount int
	}{
		{
			name:      "returns empty list when no functions",
			setup:     func(_ FunctionService) {},
			wantCount: 0,
		},
		{
			name: "returns all functions",
			setup: func(svc FunctionService) {
				ctx := context.Background()
				_, _ = svc.Create(ctx, CreateFunctionInput{Name: "fn_a", Body: "x"})
				_, _ = svc.Create(ctx, CreateFunctionInput{Name: "fn_b", Body: "y"})
				_, _ = svc.Create(ctx, CreateFunctionInput{Name: "fn_c", Body: "z"})
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := setupFnServiceTest(t)
			tt.setup(svc)

			functions, err := svc.ListAll(context.Background())
			require.NoError(t, err)
			assert.Len(t, functions, tt.wantCount)
		})
	}
}

func TestFunctionService_Update_NotFound(t *testing.T) {
	svc, _, _ := setupFnServiceTest(t)
	_, err := svc.Update(context.Background(), uuid.New(), UpdateFunctionInput{
		Body:       "x * 2",
		ReturnType: "number",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFunctionService_Update_InvalidReturnType(t *testing.T) {
	svc, _, _ := setupFnServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateFunctionInput{Name: "test", Body: "x"})
	require.NoError(t, err)

	_, err = svc.Update(ctx, created.ID, UpdateFunctionInput{
		Body:       "x * 2",
		ReturnType: "invalid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid return_type")
}

func TestFunctionService_Create_CycleDetection(t *testing.T) {
	fnA := Function{ID: uuid.New(), Name: "fn_a", Body: "fn.fn_b(x)"}

	repo := newMockFunctionRepo()
	repo.functions[fnA.ID] = &fnA
	repo.byName[fnA.Name] = &fnA

	loader := &mockFnCacheLoader{functions: []Function{fnA}}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewFunctionService(nil, repo, cache, nil)

	// Try to create fn_b that calls fn_a → cycle (fn_a calls fn_b, fn_b calls fn_a)
	_, err := svc.Create(context.Background(), CreateFunctionInput{Name: "fn_b", Body: "fn.fn_a(x)"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cyclic dependency")
}

func TestFunctionService_Create_NestingDepthExceeded(t *testing.T) {
	fnA := Function{ID: uuid.New(), Name: "fn_a", Body: "x"}
	fnB := Function{ID: uuid.New(), Name: "fn_b", Body: "fn.fn_a(x)"}
	fnC := Function{ID: uuid.New(), Name: "fn_c", Body: "fn.fn_b(x)"}

	repo := newMockFunctionRepo()
	for _, fn := range []*Function{&fnA, &fnB, &fnC} {
		repo.functions[fn.ID] = fn
		repo.byName[fn.Name] = fn
	}

	loader := &mockFnCacheLoader{functions: []Function{fnA, fnB, fnC}}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewFunctionService(nil, repo, cache, nil)

	// fn_d → fn_c → fn_b → fn_a (depth 4 > maxNestingDepth=3)
	_, err := svc.Create(context.Background(), CreateFunctionInput{Name: "fn_d", Body: "fn.fn_c(x)"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nesting depth")
}

func TestFunctionService_Update_CycleDetection(t *testing.T) {
	fnA := Function{ID: uuid.New(), Name: "fn_a", Body: "x * 2"}
	fnB := Function{ID: uuid.New(), Name: "fn_b", Body: "fn.fn_a(x)"}

	repo := newMockFunctionRepo()
	for _, fn := range []*Function{&fnA, &fnB} {
		repo.functions[fn.ID] = fn
		repo.byName[fn.Name] = fn
	}

	loader := &mockFnCacheLoader{functions: []Function{fnA, fnB}}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewFunctionService(nil, repo, cache, nil)

	// Update fn_a to call fn_b → cycle (fn_a → fn_b → fn_a)
	_, err := svc.Update(context.Background(), fnA.ID, UpdateFunctionInput{
		Body:       "fn.fn_b(x)",
		ReturnType: "any",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cyclic dependency")
}

func TestFunctionService_Create_WithOnChangeCallback(t *testing.T) {
	repo := newMockFunctionRepo()
	loader := &mockFnCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	callbackCalled := false
	onChange := func(_ context.Context) error {
		callbackCalled = true
		return nil
	}

	svc := NewFunctionService(nil, repo, cache, onChange)
	_, err := svc.Create(context.Background(), CreateFunctionInput{
		Name: "test", Body: "x",
	})
	require.NoError(t, err)
	assert.True(t, callbackCalled, "onChange callback should be called after create")
}

func TestFunctionService_Update_WithOnChangeCallback(t *testing.T) {
	repo := newMockFunctionRepo()
	loader := &mockFnCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	callbackCount := 0
	onChange := func(_ context.Context) error {
		callbackCount++
		return nil
	}

	svc := NewFunctionService(nil, repo, cache, onChange)
	ctx := context.Background()

	fn, err := svc.Create(ctx, CreateFunctionInput{Name: "test", Body: "x"})
	require.NoError(t, err)
	assert.Equal(t, 1, callbackCount)

	_, err = svc.Update(ctx, fn.ID, UpdateFunctionInput{Body: "x * 2", ReturnType: "any"})
	require.NoError(t, err)
	assert.Equal(t, 2, callbackCount)
}

func TestFunctionService_Delete_WithOnChangeCallback(t *testing.T) {
	repo := newMockFunctionRepo()
	loader := &mockFnCacheLoader{}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	callbackCount := 0
	onChange := func(_ context.Context) error {
		callbackCount++
		return nil
	}

	svc := NewFunctionService(nil, repo, cache, onChange)
	ctx := context.Background()

	fn, err := svc.Create(ctx, CreateFunctionInput{Name: "test", Body: "x"})
	require.NoError(t, err)

	err = svc.Delete(ctx, fn.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, callbackCount, "onChange should be called on create and delete")
}

func TestValidateFunctionInput(t *testing.T) {
	tests := []struct {
		name       string
		fnName     string
		body       string
		returnType string
		params     []FunctionParam
		wantErr    bool
	}{
		{
			name:       "valid input",
			fnName:     "my_func",
			body:       "x + 1",
			returnType: "number",
			params:     []FunctionParam{{Name: "x", Type: "number"}},
		},
		{
			name:       "empty name invalid",
			fnName:     "",
			body:       "x",
			returnType: "any",
			wantErr:    true,
		},
		{
			name:       "uppercase name invalid",
			fnName:     "MyFunc",
			body:       "x",
			returnType: "any",
			wantErr:    true,
		},
		{
			name:       "empty param name invalid",
			fnName:     "test",
			body:       "x",
			returnType: "any",
			params:     []FunctionParam{{Name: "", Type: "any"}},
			wantErr:    true,
		},
		{
			name:       "name too long",
			fnName:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			body:       "x",
			returnType: "any",
			wantErr:    true,
		},
		{
			name:       "invalid param type",
			fnName:     "test",
			body:       "x",
			returnType: "any",
			params:     []FunctionParam{{Name: "x", Type: "invalid_type"}},
			wantErr:    true,
		},
		{
			name:       "invalid param name format",
			fnName:     "test",
			body:       "x",
			returnType: "any",
			params:     []FunctionParam{{Name: "BadName", Type: "any"}},
			wantErr:    true,
		},
		{
			name:       "empty return type treated as valid",
			fnName:     "test",
			body:       "x",
			returnType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFunctionInput(tt.fnName, tt.body, tt.returnType, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
