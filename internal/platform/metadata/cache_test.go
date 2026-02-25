package metadata

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// mockCacheLoader is a test mock for CacheLoader.
type mockCacheLoader struct {
	objects    []ObjectDefinition
	fields     []FieldDefinition
	rels       []RelationshipInfo
	objectsErr error
	fieldsErr  error
	relsErr    error
	refreshErr error
	refreshed  int
}

func (m *mockCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	if m.objectsErr != nil {
		return nil, m.objectsErr
	}
	return m.objects, nil
}

func (m *mockCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	if m.fieldsErr != nil {
		return nil, m.fieldsErr
	}
	return m.fields, nil
}

func (m *mockCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	if m.relsErr != nil {
		return nil, m.relsErr
	}
	return m.rels, nil
}

func (m *mockCacheLoader) RefreshMaterializedView(_ context.Context) error {
	m.refreshed++
	return m.refreshErr
}

func (m *mockCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}

func (m *mockCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) {
	return nil, nil
}

func (m *mockCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return nil, nil
}

func (m *mockCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}

func TestMetadataCacheLoad(t *testing.T) {
	t.Parallel()

	objA := ObjectDefinition{ID: uuid.New(), APIName: "Account", TableName: "obj_account"}
	objB := ObjectDefinition{ID: uuid.New(), APIName: "Contact", TableName: "obj_contact"}

	subtypeAssoc := SubtypeAssociation
	fieldRef := FieldDefinition{
		ID:           uuid.New(),
		ObjectID:     objB.ID,
		APIName:      "account_id",
		FieldType:    FieldTypeReference,
		FieldSubtype: &subtypeAssoc,
	}

	rel := RelationshipInfo{
		FieldID:             fieldRef.ID,
		FieldAPIName:        "account_id",
		RelationshipName:    "Contacts",
		ChildObjectID:       objB.ID,
		ChildObjectAPIName:  "Contact",
		ParentObjectID:      objA.ID,
		ParentObjectAPIName: "Account",
		ReferenceSubtype:    SubtypeAssociation,
		OnDelete:            "set_null",
	}

	tests := []struct {
		name       string
		loader     *mockCacheLoader
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "loads successfully",
			loader: &mockCacheLoader{
				objects: []ObjectDefinition{objA, objB},
				fields:  []FieldDefinition{fieldRef},
				rels:    []RelationshipInfo{rel},
			},
		},
		{
			name: "returns error on objects load failure",
			loader: &mockCacheLoader{
				objectsErr: errors.New("db error"),
			},
			wantErr:    true,
			wantErrMsg: "objects",
		},
		{
			name: "returns error on fields load failure",
			loader: &mockCacheLoader{
				objects:   []ObjectDefinition{objA},
				fieldsErr: errors.New("db error"),
			},
			wantErr:    true,
			wantErrMsg: "fields",
		},
		{
			name: "returns error on relationships load failure",
			loader: &mockCacheLoader{
				objects: []ObjectDefinition{objA},
				fields:  []FieldDefinition{},
				relsErr: errors.New("db error"),
			},
			wantErr:    true,
			wantErrMsg: "relationships",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := NewMetadataCache(tt.loader)
			err := cache.Load(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrMsg != "" && !containsStr(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cache.IsLoaded() {
				t.Error("expected cache to be loaded")
			}
		})
	}
}

func TestMetadataCacheLookups(t *testing.T) {
	t.Parallel()

	objA := ObjectDefinition{ID: uuid.New(), APIName: "Account", TableName: "obj_account"}
	objB := ObjectDefinition{ID: uuid.New(), APIName: "Contact", TableName: "obj_contact"}

	subtypeAssoc := SubtypeAssociation
	field1 := FieldDefinition{ID: uuid.New(), ObjectID: objB.ID, APIName: "account_id", FieldType: FieldTypeReference, FieldSubtype: &subtypeAssoc}
	field2 := FieldDefinition{ID: uuid.New(), ObjectID: objB.ID, APIName: "first_name", FieldType: FieldTypeText}

	rel := RelationshipInfo{
		FieldID:             field1.ID,
		FieldAPIName:        "account_id",
		ChildObjectID:       objB.ID,
		ChildObjectAPIName:  "Contact",
		ParentObjectID:      objA.ID,
		ParentObjectAPIName: "Account",
		ReferenceSubtype:    SubtypeAssociation,
		OnDelete:            "set_null",
	}

	loader := &mockCacheLoader{
		objects: []ObjectDefinition{objA, objB},
		fields:  []FieldDefinition{field1, field2},
		rels:    []RelationshipInfo{rel},
	}

	cache := NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "GetObjectByID returns existing object",
			fn: func(t *testing.T) {
				obj, ok := cache.GetObjectByID(objA.ID)
				if !ok {
					t.Fatal("expected to find object")
				}
				if obj.APIName != "Account" {
					t.Errorf("APIName = %q, want Account", obj.APIName)
				}
			},
		},
		{
			name: "GetObjectByID returns false for nonexistent",
			fn: func(t *testing.T) {
				_, ok := cache.GetObjectByID(uuid.New())
				if ok {
					t.Error("expected not to find object")
				}
			},
		},
		{
			name: "GetObjectByAPIName returns existing object",
			fn: func(t *testing.T) {
				obj, ok := cache.GetObjectByAPIName("Contact")
				if !ok {
					t.Fatal("expected to find object")
				}
				if obj.ID != objB.ID {
					t.Errorf("ID = %v, want %v", obj.ID, objB.ID)
				}
			},
		},
		{
			name: "GetFieldByID returns existing field",
			fn: func(t *testing.T) {
				f, ok := cache.GetFieldByID(field1.ID)
				if !ok {
					t.Fatal("expected to find field")
				}
				if f.APIName != "account_id" {
					t.Errorf("APIName = %q, want account_id", f.APIName)
				}
			},
		},
		{
			name: "GetFieldsByObjectID returns fields for object",
			fn: func(t *testing.T) {
				fields := cache.GetFieldsByObjectID(objB.ID)
				if len(fields) != 2 {
					t.Errorf("got %d fields, want 2", len(fields))
				}
			},
		},
		{
			name: "GetFieldsByObjectID returns empty for unknown object",
			fn: func(t *testing.T) {
				fields := cache.GetFieldsByObjectID(uuid.New())
				if len(fields) != 0 {
					t.Errorf("got %d fields, want 0", len(fields))
				}
			},
		},
		{
			name: "GetForwardRelationships returns child relationships",
			fn: func(t *testing.T) {
				rels := cache.GetForwardRelationships(objB.ID)
				if len(rels) != 1 {
					t.Fatalf("got %d rels, want 1", len(rels))
				}
				if rels[0].ParentObjectAPIName != "Account" {
					t.Errorf("parent = %q, want Account", rels[0].ParentObjectAPIName)
				}
			},
		},
		{
			name: "GetReverseRelationships returns parent relationships",
			fn: func(t *testing.T) {
				rels := cache.GetReverseRelationships(objA.ID)
				if len(rels) != 1 {
					t.Fatalf("got %d rels, want 1", len(rels))
				}
				if rels[0].ChildObjectAPIName != "Contact" {
					t.Errorf("child = %q, want Contact", rels[0].ChildObjectAPIName)
				}
			},
		},
		{
			name: "GetReverseRelationships returns empty for leaf object",
			fn: func(t *testing.T) {
				rels := cache.GetReverseRelationships(objB.ID)
				if len(rels) != 0 {
					t.Errorf("got %d rels, want 0", len(rels))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fn(t)
		})
	}
}

func TestMetadataCacheInvalidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		loader     *mockCacheLoader
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "invalidates and reloads",
			loader: &mockCacheLoader{
				objects: []ObjectDefinition{{ID: uuid.New(), APIName: "Obj"}},
			},
		},
		{
			name: "returns error on refresh failure",
			loader: &mockCacheLoader{
				refreshErr: errors.New("pg: cannot refresh"),
			},
			wantErr:    true,
			wantErrMsg: "refresh MV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := NewMetadataCache(tt.loader)
			err := cache.Invalidate(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrMsg != "" && !containsStr(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.loader.refreshed != 1 {
				t.Errorf("refreshed = %d, want 1", tt.loader.refreshed)
			}
			if !cache.IsLoaded() {
				t.Error("expected cache to be loaded after invalidate")
			}
		})
	}
}

func TestMetadataCacheConcurrentAccess(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	loader := &mockCacheLoader{
		objects: []ObjectDefinition{{ID: objID, APIName: "Account"}},
		fields: []FieldDefinition{{
			ID:       uuid.New(),
			ObjectID: objID,
			APIName:  "name",
		}},
	}

	cache := NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.GetObjectByID(objID)
			cache.GetObjectByAPIName("Account")
			cache.GetFieldsByObjectID(objID)
			cache.GetForwardRelationships(objID)
			cache.GetReverseRelationships(objID)
		}()
	}
	wg.Wait()
}
