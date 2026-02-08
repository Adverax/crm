package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

func newObjectService(
	objRepo *mockObjectRepo,
	fieldRepo *mockFieldRepo,
	ddlExec *mockDDLExec,
	cache *mockCacheInvalidator,
) ObjectService {
	return NewObjectService(
		&mockTxBeginner{},
		objRepo,
		fieldRepo,
		ddlExec,
		cache,
	)
}

func TestObjectServiceCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateObjectInput
		setup      func(*mockObjectRepo, *mockDDLExec, *mockCacheInvalidator)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "creates custom object successfully",
			input: CreateObjectInput{
				APIName:               "Invoice__c",
				Label:                 "Invoice",
				PluralLabel:           "Invoices",
				ObjectType:            ObjectTypeCustom,
				IsCustomFieldsAllowed: true,
				IsDeleteableObject:    true,
				IsCreateable:          true,
				IsUpdateable:          true,
				IsDeleteable:          true,
				IsQueryable:           true,
			},
		},
		{
			name: "creates standard object successfully",
			input: CreateObjectInput{
				APIName:      "Account",
				Label:        "Account",
				PluralLabel:  "Accounts",
				ObjectType:   ObjectTypeStandard,
				IsCreateable: true,
				IsQueryable:  true,
			},
		},
		{
			name: "rejects empty api_name",
			input: CreateObjectInput{
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr:    true,
			wantErrMsg: "api_name is required",
		},
		{
			name: "rejects empty label",
			input: CreateObjectInput{
				APIName:     "Test__c",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr:    true,
			wantErrMsg: "label is required",
		},
		{
			name: "rejects custom object without __c suffix",
			input: CreateObjectInput{
				APIName:     "Invoice",
				Label:       "Invoice",
				PluralLabel: "Invoices",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr:    true,
			wantErrMsg: "__c",
		},
		{
			name: "rejects duplicate api_name",
			input: CreateObjectInput{
				APIName:     "Account",
				Label:       "Account",
				PluralLabel: "Accounts",
				ObjectType:  ObjectTypeStandard,
			},
			setup: func(repo *mockObjectRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				repo.addObject(&ObjectDefinition{
					ID:      uuid.New(),
					APIName: "Account",
				})
			},
			wantErr:    true,
			wantErrMsg: "already exists",
		},
		{
			name: "rejects reserved word as api_name",
			input: CreateObjectInput{
				APIName:     "select",
				Label:       "Select",
				PluralLabel: "Selects",
				ObjectType:  ObjectTypeStandard,
			},
			wantErr:    true,
			wantErrMsg: "reserved word",
		},
		{
			name: "returns error when DDL fails",
			input: CreateObjectInput{
				APIName:     "Broken__c",
				Label:       "Broken",
				PluralLabel: "Brokens",
				ObjectType:  ObjectTypeCustom,
			},
			setup: func(_ *mockObjectRepo, ddl *mockDDLExec, _ *mockCacheInvalidator) {
				ddl.execErr = errors.New("pg: connection refused")
			},
			wantErr:    true,
			wantErrMsg: "DDL CREATE TABLE",
		},
		{
			name: "returns error when cache invalidation fails",
			input: CreateObjectInput{
				APIName:     "Cached__c",
				Label:       "Cached",
				PluralLabel: "Cacheds",
				ObjectType:  ObjectTypeCustom,
			},
			setup: func(_ *mockObjectRepo, _ *mockDDLExec, cache *mockCacheInvalidator) {
				cache.err = errors.New("cache failure")
			},
			wantErr:    true,
			wantErrMsg: "cache invalidate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			fieldRepo := newMockFieldRepo()
			ddlExec := newMockDDLExec()
			cache := newMockCache()

			if tt.setup != nil {
				tt.setup(objRepo, ddlExec, cache)
			}

			svc := newObjectService(objRepo, fieldRepo, ddlExec, cache)
			result, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if tt.wantErrMsg != "" && !containsStr(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErrMsg)
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
				t.Errorf("APIName = %q, want %q", result.APIName, tt.input.APIName)
			}
			if result.TableName == "" {
				t.Error("expected non-empty TableName")
			}
			if len(ddlExec.executed) == 0 {
				t.Error("expected DDL to be executed")
			}
			if cache.invalidated != 1 {
				t.Errorf("cache.invalidated = %d, want 1", cache.invalidated)
			}
		})
	}
}

func TestObjectServiceGetByID(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		setup      func(*mockObjectRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns object when exists",
			id:   existingID,
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:      existingID,
					APIName: "Account",
					Label:   "Account",
				})
			},
		},
		{
			name:       "returns NotFound when object does not exist",
			id:         uuid.New(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			if tt.setup != nil {
				tt.setup(objRepo)
			}

			svc := newObjectService(objRepo, newMockFieldRepo(), newMockDDLExec(), newMockCache())
			result, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var appErr *apperror.AppError
				if errors.As(err, &appErr) && appErr.HTTPStatus != 404 {
					t.Errorf("expected 404 status, got %d", appErr.HTTPStatus)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.id {
				t.Errorf("ID = %v, want %v", result.ID, tt.id)
			}
		})
	}
}

func TestObjectServiceList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filter    ObjectFilter
		setup     func(*mockObjectRepo)
		wantCount int
		wantTotal int64
	}{
		{
			name:   "returns empty list",
			filter: ObjectFilter{Page: 1, PerPage: 20},
		},
		{
			name:   "returns objects with pagination",
			filter: ObjectFilter{Page: 1, PerPage: 10},
			setup: func(repo *mockObjectRepo) {
				for i := 0; i < 3; i++ {
					repo.addObject(&ObjectDefinition{
						ID:      uuid.New(),
						APIName: "Obj" + string(rune('A'+i)),
					})
				}
			},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:   "uses default PerPage when zero",
			filter: ObjectFilter{Page: 0, PerPage: 0},
		},
		{
			name:   "caps PerPage at 100",
			filter: ObjectFilter{Page: 1, PerPage: 999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			if tt.setup != nil {
				tt.setup(objRepo)
			}

			svc := newObjectService(objRepo, newMockFieldRepo(), newMockDDLExec(), newMockCache())
			objects, total, err := svc.List(context.Background(), tt.filter)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(objects) != tt.wantCount {
				t.Errorf("got %d objects, want %d", len(objects), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestObjectServiceUpdate(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	platformManagedID := uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		input      UpdateObjectInput
		setup      func(*mockObjectRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "updates object successfully",
			id:   existingID,
			input: UpdateObjectInput{
				Label:       "Updated Label",
				PluralLabel: "Updated Labels",
				Description: "Updated",
			},
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:      existingID,
					APIName: "Account",
					Label:   "Account",
				})
			},
		},
		{
			name: "returns NotFound when object does not exist",
			id:   uuid.New(),
			input: UpdateObjectInput{
				Label:       "Test",
				PluralLabel: "Tests",
			},
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects update of platform-managed object",
			id:   platformManagedID,
			input: UpdateObjectInput{
				Label:       "Cannot Update",
				PluralLabel: "Cannot Updates",
			},
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:                platformManagedID,
					APIName:           "System",
					IsPlatformManaged: true,
				})
			},
			wantErr:    true,
			wantErrMsg: "platform-managed",
		},
		{
			name: "rejects empty label",
			id:   existingID,
			input: UpdateObjectInput{
				PluralLabel: "Tests",
			},
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:      existingID,
					APIName: "Account",
				})
			},
			wantErr:    true,
			wantErrMsg: "label is required",
		},
		{
			name: "rejects empty plural_label",
			id:   existingID,
			input: UpdateObjectInput{
				Label: "Test",
			},
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:      existingID,
					APIName: "Account",
				})
			},
			wantErr:    true,
			wantErrMsg: "plural_label is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			if tt.setup != nil {
				tt.setup(objRepo)
			}

			svc := newObjectService(objRepo, newMockFieldRepo(), newMockDDLExec(), newMockCache())
			result, err := svc.Update(context.Background(), tt.id, tt.input)

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
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Label != tt.input.Label {
				t.Errorf("Label = %q, want %q", result.Label, tt.input.Label)
			}
		})
	}
}

func TestObjectServiceDelete(t *testing.T) {
	t.Parallel()

	deletableID := uuid.New()
	nonDeletableID := uuid.New()
	platformManagedID := uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		setup      func(*mockObjectRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "deletes object successfully",
			id:   deletableID,
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:                 deletableID,
					APIName:            "Temp__c",
					TableName:          "obj_temp",
					IsDeleteableObject: true,
				})
			},
		},
		{
			name:       "returns NotFound when object does not exist",
			id:         uuid.New(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects deletion of non-deleteable object",
			id:   nonDeletableID,
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:                 nonDeletableID,
					APIName:            "Core",
					IsDeleteableObject: false,
				})
			},
			wantErr:    true,
			wantErrMsg: "cannot be deleted",
		},
		{
			name: "rejects deletion of platform-managed object",
			id:   platformManagedID,
			setup: func(repo *mockObjectRepo) {
				repo.addObject(&ObjectDefinition{
					ID:                 platformManagedID,
					APIName:            "System",
					IsDeleteableObject: true,
					IsPlatformManaged:  true,
				})
			},
			wantErr:    true,
			wantErrMsg: "platform-managed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			if tt.setup != nil {
				tt.setup(objRepo)
			}

			ddlExec := newMockDDLExec()
			cache := newMockCache()
			svc := newObjectService(objRepo, newMockFieldRepo(), ddlExec, cache)
			err := svc.Delete(context.Background(), tt.id)

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
			if len(ddlExec.executed) == 0 {
				t.Error("expected DDL to be executed")
			}
			if cache.invalidated != 1 {
				t.Errorf("cache.invalidated = %d, want 1", cache.invalidated)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
