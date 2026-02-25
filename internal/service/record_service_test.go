package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/soql"
)

// --- Mock QueryService ---

type mockQueryService struct {
	executeFunc func(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error)
}

func (m *mockQueryService) Execute(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error) {
	return m.executeFunc(ctx, query, params)
}

// --- Mock DMLService ---

type mockDMLService struct {
	executeFunc func(ctx context.Context, statement string) (*engine.Result, error)
}

func (m *mockDMLService) Execute(ctx context.Context, statement string) (*engine.Result, error) {
	return m.executeFunc(ctx, statement)
}

func (m *mockDMLService) Prepare(_ context.Context, _ string) (*engine.CompiledDML, error) {
	return nil, nil
}

func TestRecordService_List(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	cache := buildTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name      string
		object    string
		params    ListParams
		soqlFunc  func(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns records for valid object",
			object: "Account",
			params: ListParams{Page: 1, PerPage: 20},
			soqlFunc: func(_ context.Context, query string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				if query == "SELECT COUNT() FROM Account" {
					return &soql.QueryResult{TotalSize: 2, Done: true}, nil
				}
				return &soql.QueryResult{
					TotalSize: 2,
					Done:      true,
					Records: []map[string]any{
						{"Id": "id1", "Name": "Acme"},
						{"Id": "id2", "Name": "Globex"},
					},
				}, nil
			},
			wantCount: 2,
		},
		{
			name:   "returns error for unknown object",
			object: "Unknown",
			params: ListParams{Page: 1, PerPage: 20},
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:   "normalizes page and perPage",
			object: "Account",
			params: ListParams{Page: 0, PerPage: 0},
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return &soql.QueryResult{TotalSize: 0, Done: true, Records: nil}, nil
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewRecordService(cache, &mockQueryService{executeFunc: tt.soqlFunc}, nil)
			result, err := svc.List(context.Background(), tt.object, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Data) != tt.wantCount {
				t.Errorf("expected %d records, got %d", tt.wantCount, len(result.Data))
			}
		})
	}
}

func TestRecordService_GetByID(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	cache := buildTestCache(objID, "Account", "obj_account")
	validID := uuid.New().String()

	tests := []struct {
		name     string
		object   string
		recordID string
		soqlFunc func(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error)
		wantErr  bool
	}{
		{
			name:     "returns record when exists",
			object:   "Account",
			recordID: validID,
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return &soql.QueryResult{
					TotalSize: 1,
					Done:      true,
					Records:   []map[string]any{{"Id": validID, "Name": "Acme"}},
				}, nil
			},
		},
		{
			name:     "returns NotFound when record does not exist",
			object:   "Account",
			recordID: validID,
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return &soql.QueryResult{TotalSize: 0, Done: true, Records: nil}, nil
			},
			wantErr: true,
		},
		{
			name:     "returns error for invalid UUID",
			object:   "Account",
			recordID: "not-a-uuid",
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "returns error for unknown object",
			object:   "Unknown",
			recordID: validID,
			soqlFunc: func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewRecordService(cache, &mockQueryService{executeFunc: tt.soqlFunc}, nil)
			result, err := svc.GetByID(context.Background(), tt.object, tt.recordID)

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
				t.Fatal("expected non-nil result")
			}
		})
	}
}

func TestRecordService_Create(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	userID := uuid.New()
	cache := buildTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name    string
		object  string
		fields  map[string]any
		ctx     context.Context
		dmlFunc func(ctx context.Context, statement string) (*engine.Result, error)
		wantErr bool
		wantID  string
	}{
		{
			name:   "creates record successfully",
			object: "Account",
			fields: map[string]any{"Name": "Test"},
			ctx:    security.ContextWithUser(context.Background(), security.UserContext{UserID: userID}),
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return &engine.Result{
					RowsAffected: 1,
					InsertedIds:  []string{"new-id"},
				}, nil
			},
			wantID: "new-id",
		},
		{
			name:   "returns error without user context",
			object: "Account",
			fields: map[string]any{"Name": "Test"},
			ctx:    context.Background(),
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:   "returns error for non-createable object",
			object: "ReadOnly",
			fields: map[string]any{"Name": "Test"},
			ctx:    security.ContextWithUser(context.Background(), security.UserContext{UserID: userID}),
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCache := cache
			if tt.object == "ReadOnly" {
				testCache = buildTestCacheWithFlags(objID, "ReadOnly", "obj_readonly", false, true, true)
			}

			svc := NewRecordService(testCache, nil, &mockDMLService{executeFunc: tt.dmlFunc})
			result, err := svc.Create(tt.ctx, tt.object, tt.fields)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.wantID {
				t.Errorf("expected ID %s, got %s", tt.wantID, result.ID)
			}
		})
	}
}

func TestRecordService_Update(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	userID := uuid.New()
	recordID := uuid.New().String()
	cache := buildTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name     string
		object   string
		recordID string
		fields   map[string]any
		ctx      context.Context
		dmlFunc  func(ctx context.Context, statement string) (*engine.Result, error)
		wantErr  bool
	}{
		{
			name:     "updates record successfully",
			object:   "Account",
			recordID: recordID,
			fields:   map[string]any{"Name": "Updated"},
			ctx:      security.ContextWithUser(context.Background(), security.UserContext{UserID: userID}),
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return &engine.Result{RowsAffected: 1, UpdatedIds: []string{recordID}}, nil
			},
		},
		{
			name:     "returns error for invalid UUID",
			object:   "Account",
			recordID: "bad-uuid",
			fields:   map[string]any{"Name": "Updated"},
			ctx:      security.ContextWithUser(context.Background(), security.UserContext{UserID: userID}),
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewRecordService(cache, nil, &mockDMLService{executeFunc: tt.dmlFunc})
			err := svc.Update(tt.ctx, tt.object, tt.recordID, tt.fields)

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

func TestRecordService_Delete(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	recordID := uuid.New().String()
	cache := buildTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name     string
		object   string
		recordID string
		dmlFunc  func(ctx context.Context, statement string) (*engine.Result, error)
		wantErr  bool
	}{
		{
			name:     "deletes record successfully",
			object:   "Account",
			recordID: recordID,
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return &engine.Result{RowsAffected: 1, DeletedIds: []string{recordID}}, nil
			},
		},
		{
			name:     "returns error for invalid UUID",
			object:   "Account",
			recordID: "bad-uuid",
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "returns error for unknown object",
			object:   "Unknown",
			recordID: recordID,
			dmlFunc: func(_ context.Context, _ string) (*engine.Result, error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewRecordService(cache, nil, &mockDMLService{executeFunc: tt.dmlFunc})
			err := svc.Delete(context.Background(), tt.object, tt.recordID)

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

func TestFormatDMLValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: "NULL"},
		{name: "string", value: "hello", want: "'hello'"},
		{name: "string with quote", value: "it's", want: "'it''s'"},
		{name: "int", value: 42, want: "42"},
		{name: "float64 integer", value: float64(10), want: "10"},
		{name: "float64 decimal", value: 3.14, want: "3.14"},
		{name: "bool true", value: true, want: "TRUE"},
		{name: "bool false", value: false, want: "FALSE"},
		{name: "int64", value: int64(100), want: "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatDMLValue(tt.value)
			if got != tt.want {
				t.Errorf("formatDMLValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "valid UUID", id: uuid.New().String()},
		{name: "invalid UUID", id: "not-uuid", wantErr: true},
		{name: "empty string", id: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateUUID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUUID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

// --- Test Helpers ---

func buildTestCache(objID uuid.UUID, apiName, tableName string) *metadata.MetadataCache {
	return buildTestCacheWithFlags(objID, apiName, tableName, true, true, true)
}

func buildTestCacheWithFlags(objID uuid.UUID, apiName, tableName string, createable, updateable, deleteable bool) *metadata.MetadataCache {
	loader := &stubCacheLoader{
		objects: []metadata.ObjectDefinition{
			{
				ID:           objID,
				APIName:      apiName,
				TableName:    tableName,
				Label:        apiName,
				PluralLabel:  apiName + "s",
				IsCreateable: createable,
				IsUpdateable: updateable,
				IsDeleteable: deleteable,
				IsQueryable:  true,
			},
		},
		fields: []metadata.FieldDefinition{
			{
				ID:         uuid.New(),
				ObjectID:   objID,
				APIName:    "Name",
				Label:      "Name",
				FieldType:  metadata.FieldTypeText,
				IsRequired: true,
				SortOrder:  1,
			},
		},
	}
	cache := metadata.NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}
	return cache
}

type stubCacheLoader struct {
	objects []metadata.ObjectDefinition
	fields  []metadata.FieldDefinition
}

func (s *stubCacheLoader) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return s.objects, nil
}

func (s *stubCacheLoader) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return s.fields, nil
}

func (s *stubCacheLoader) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}

func (s *stubCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func (s *stubCacheLoader) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return nil, nil
}

func (s *stubCacheLoader) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}

func (s *stubCacheLoader) LoadAllObjectViews(_ context.Context) ([]metadata.ObjectView, error) {
	return nil, nil
}

func (s *stubCacheLoader) LoadAllProcedures(_ context.Context) ([]metadata.Procedure, error) {
	return nil, nil
}
