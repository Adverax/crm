package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

func newFieldService(
	objRepo *mockObjectRepo,
	fieldRepo *mockFieldRepo,
	polyRepo *mockPolymorphicRepo,
	ddlExec *mockDDLExec,
	cache *mockCacheInvalidator,
) FieldService {
	return NewFieldService(
		&mockTxBeginner{},
		objRepo,
		fieldRepo,
		polyRepo,
		ddlExec,
		cache,
	)
}

func TestFieldServiceCreate(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()
	refObjectID := uuid.New()
	subtypePlain := SubtypePlain
	subtypeAssoc := SubtypeAssociation
	subtypeComp := SubtypeComposition
	maxLen := 100
	relName := func(s string) *string { return &s }

	tests := []struct {
		name       string
		input      CreateFieldInput
		setup      func(*mockObjectRepo, *mockFieldRepo, *mockDDLExec, *mockCacheInvalidator)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "creates text field successfully",
			input: CreateFieldInput{
				ObjectID:     objectID,
				APIName:      "first_name__c",
				Label:        "First Name",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				IsCustom:     true,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Contact__c",
					TableName:             "obj_contact",
					IsCustomFieldsAllowed: true,
				})
			},
		},
		{
			name: "creates association reference field",
			input: CreateFieldInput{
				ObjectID:           objectID,
				APIName:            "account_id__c",
				Label:              "Account",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subtypeAssoc,
				ReferencedObjectID: &refObjectID,
				IsCustom:           true,
				Config:             FieldConfig{RelationshipName: relName("Account")},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Contact__c",
					TableName:             "obj_contact",
					IsCustomFieldsAllowed: true,
				})
				objRepo.addObject(&ObjectDefinition{
					ID:        refObjectID,
					APIName:   "Account",
					TableName: "obj_account",
				})
			},
		},
		{
			name: "rejects field on nonexistent object",
			input: CreateFieldInput{
				ObjectID:     uuid.New(),
				APIName:      "field__c",
				Label:        "Field",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				IsCustom:     true,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects custom field when not allowed",
			input: CreateFieldInput{
				ObjectID:     objectID,
				APIName:      "blocked__c",
				Label:        "Blocked",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				IsCustom:     true,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Locked",
					TableName:             "obj_locked",
					IsCustomFieldsAllowed: false,
				})
			},
			wantErr:    true,
			wantErrMsg: "custom fields are not allowed",
		},
		{
			name: "rejects duplicate field name",
			input: CreateFieldInput{
				ObjectID:     objectID,
				APIName:      "email__c",
				Label:        "Email",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				IsCustom:     true,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			setup: func(objRepo *mockObjectRepo, fieldRepo *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Contact__c",
					TableName:             "obj_contact",
					IsCustomFieldsAllowed: true,
				})
				fID := uuid.New()
				existing := &FieldDefinition{
					ID:       fID,
					ObjectID: objectID,
					APIName:  "email__c",
				}
				fieldRepo.fields[fID] = existing
				fieldRepo.byObjName[objectID.String()+"/email__c"] = existing
			},
			wantErr:    true,
			wantErrMsg: "already exists",
		},
		{
			name: "rejects reference to nonexistent object",
			input: CreateFieldInput{
				ObjectID:           objectID,
				APIName:            "ref__c",
				Label:              "Ref",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subtypeAssoc,
				ReferencedObjectID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				IsCustom:           true,
				Config:             FieldConfig{RelationshipName: relName("Ref")},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Contact__c",
					TableName:             "obj_contact",
					IsCustomFieldsAllowed: true,
				})
			},
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects empty api_name",
			input: CreateFieldInput{
				ObjectID:     objectID,
				Label:        "Test",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			wantErr:    true,
			wantErrMsg: "api_name is required",
		},
		{
			name: "returns error on DDL failure",
			input: CreateFieldInput{
				ObjectID:     objectID,
				APIName:      "broken__c",
				Label:        "Broken",
				FieldType:    FieldTypeText,
				FieldSubtype: &subtypePlain,
				IsCustom:     true,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, ddlExec *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Contact__c",
					TableName:             "obj_contact",
					IsCustomFieldsAllowed: true,
				})
				ddlExec.execErr = errors.New("pg: disk full")
			},
			wantErr:    true,
			wantErrMsg: "execute DDL",
		},
		{
			name: "validates composition depth",
			input: CreateFieldInput{
				ObjectID:           objectID,
				APIName:            "parent__c",
				Label:              "Parent",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subtypeComp,
				ReferencedObjectID: &objectID,
				IsCustom:           true,
				Config:             FieldConfig{RelationshipName: relName("Parent")},
			},
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo, _ *mockDDLExec, _ *mockCacheInvalidator) {
				objRepo.addObject(&ObjectDefinition{
					ID:                    objectID,
					APIName:               "Item__c",
					TableName:             "obj_item",
					IsCustomFieldsAllowed: true,
				})
			},
			wantErr:    true,
			wantErrMsg: "composition self-reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			fieldRepo := newMockFieldRepo()
			polyRepo := newMockPolymorphicRepo()
			ddlExec := newMockDDLExec()
			cache := newMockCache()

			if tt.setup != nil {
				tt.setup(objRepo, fieldRepo, ddlExec, cache)
			}

			svc := newFieldService(objRepo, fieldRepo, polyRepo, ddlExec, cache)
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
			if len(ddlExec.executed) == 0 {
				t.Error("expected DDL to be executed")
			}
			if cache.invalidated != 1 {
				t.Errorf("cache.invalidated = %d, want 1", cache.invalidated)
			}
		})
	}
}

func TestFieldServiceGetByID(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	objectID := uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		setup      func(*mockFieldRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns field when exists",
			id:   existingID,
			setup: func(repo *mockFieldRepo) {
				repo.fields[existingID] = &FieldDefinition{
					ID:       existingID,
					ObjectID: objectID,
					APIName:  "first_name",
					Label:    "First Name",
				}
			},
		},
		{
			name:       "returns NotFound when field does not exist",
			id:         uuid.New(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fieldRepo := newMockFieldRepo()
			if tt.setup != nil {
				tt.setup(fieldRepo)
			}

			svc := newFieldService(newMockObjectRepo(), fieldRepo, newMockPolymorphicRepo(), newMockDDLExec(), newMockCache())
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

func TestFieldServiceListByObjectID(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()
	otherObjectID := uuid.New()

	tests := []struct {
		name       string
		objectID   uuid.UUID
		setup      func(*mockObjectRepo, *mockFieldRepo)
		wantCount  int
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "returns fields for object",
			objectID: objectID,
			setup: func(objRepo *mockObjectRepo, fieldRepo *mockFieldRepo) {
				objRepo.addObject(&ObjectDefinition{ID: objectID, APIName: "Contact"})
				for i := 0; i < 3; i++ {
					id := uuid.New()
					fieldRepo.fields[id] = &FieldDefinition{
						ID:       id,
						ObjectID: objectID,
						APIName:  "field_" + string(rune('a'+i)),
					}
				}
				otherId := uuid.New()
				fieldRepo.fields[otherId] = &FieldDefinition{
					ID:       otherId,
					ObjectID: otherObjectID,
					APIName:  "other_field",
				}
			},
			wantCount: 3,
		},
		{
			name:     "returns empty list for object with no fields",
			objectID: objectID,
			setup: func(objRepo *mockObjectRepo, _ *mockFieldRepo) {
				objRepo.addObject(&ObjectDefinition{ID: objectID, APIName: "Empty"})
			},
			wantCount: 0,
		},
		{
			name:       "returns NotFound for nonexistent object",
			objectID:   uuid.New(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			fieldRepo := newMockFieldRepo()
			if tt.setup != nil {
				tt.setup(objRepo, fieldRepo)
			}

			svc := newFieldService(objRepo, fieldRepo, newMockPolymorphicRepo(), newMockDDLExec(), newMockCache())
			fields, err := svc.ListByObjectID(context.Background(), tt.objectID)

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
			if len(fields) != tt.wantCount {
				t.Errorf("got %d fields, want %d", len(fields), tt.wantCount)
			}
		})
	}
}

func TestFieldServiceUpdate(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	objectID := uuid.New()
	systemFieldID := uuid.New()
	platformFieldID := uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		input      UpdateFieldInput
		setup      func(*mockObjectRepo, *mockFieldRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "updates field successfully",
			id:   existingID,
			input: UpdateFieldInput{
				Label:       "Updated Name",
				Description: "Updated",
				IsRequired:  true,
			},
			setup: func(objRepo *mockObjectRepo, fieldRepo *mockFieldRepo) {
				objRepo.addObject(&ObjectDefinition{
					ID:        objectID,
					APIName:   "Contact",
					TableName: "obj_contact",
				})
				fieldRepo.fields[existingID] = &FieldDefinition{
					ID:        existingID,
					ObjectID:  objectID,
					APIName:   "first_name",
					Label:     "First Name",
					FieldType: FieldTypeText,
				}
			},
		},
		{
			name: "returns NotFound for nonexistent field",
			id:   uuid.New(),
			input: UpdateFieldInput{
				Label: "Test",
			},
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects update of system field",
			id:   systemFieldID,
			input: UpdateFieldInput{
				Label: "Cannot Update",
			},
			setup: func(_ *mockObjectRepo, fieldRepo *mockFieldRepo) {
				fieldRepo.fields[systemFieldID] = &FieldDefinition{
					ID:            systemFieldID,
					ObjectID:      objectID,
					APIName:       "id",
					IsSystemField: true,
				}
			},
			wantErr:    true,
			wantErrMsg: "system field",
		},
		{
			name: "rejects update of platform-managed field",
			id:   platformFieldID,
			input: UpdateFieldInput{
				Label: "Cannot Update",
			},
			setup: func(_ *mockObjectRepo, fieldRepo *mockFieldRepo) {
				fieldRepo.fields[platformFieldID] = &FieldDefinition{
					ID:                platformFieldID,
					ObjectID:          objectID,
					APIName:           "managed_field",
					IsPlatformManaged: true,
				}
			},
			wantErr:    true,
			wantErrMsg: "platform-managed",
		},
		{
			name: "rejects empty label",
			id:   existingID,
			input: UpdateFieldInput{
				Label: "",
			},
			setup: func(_ *mockObjectRepo, fieldRepo *mockFieldRepo) {
				fieldRepo.fields[existingID] = &FieldDefinition{
					ID:       existingID,
					ObjectID: objectID,
					APIName:  "first_name",
				}
			},
			wantErr:    true,
			wantErrMsg: "label is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			fieldRepo := newMockFieldRepo()
			if tt.setup != nil {
				tt.setup(objRepo, fieldRepo)
			}

			svc := newFieldService(objRepo, fieldRepo, newMockPolymorphicRepo(), newMockDDLExec(), newMockCache())
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

func TestFieldServiceDelete(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	objectID := uuid.New()
	systemFieldID := uuid.New()
	platformFieldID := uuid.New()
	subtypePlain := SubtypePlain

	tests := []struct {
		name       string
		id         uuid.UUID
		setup      func(*mockObjectRepo, *mockFieldRepo)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "deletes field successfully",
			id:   existingID,
			setup: func(objRepo *mockObjectRepo, fieldRepo *mockFieldRepo) {
				objRepo.addObject(&ObjectDefinition{
					ID:        objectID,
					APIName:   "Contact",
					TableName: "obj_contact",
				})
				fieldRepo.fields[existingID] = &FieldDefinition{
					ID:           existingID,
					ObjectID:     objectID,
					APIName:      "first_name",
					FieldType:    FieldTypeText,
					FieldSubtype: &subtypePlain,
				}
			},
		},
		{
			name:       "returns NotFound for nonexistent field",
			id:         uuid.New(),
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name: "rejects deletion of system field",
			id:   systemFieldID,
			setup: func(_ *mockObjectRepo, fieldRepo *mockFieldRepo) {
				fieldRepo.fields[systemFieldID] = &FieldDefinition{
					ID:            systemFieldID,
					ObjectID:      objectID,
					APIName:       "id",
					IsSystemField: true,
				}
			},
			wantErr:    true,
			wantErrMsg: "system field",
		},
		{
			name: "rejects deletion of platform-managed field",
			id:   platformFieldID,
			setup: func(_ *mockObjectRepo, fieldRepo *mockFieldRepo) {
				fieldRepo.fields[platformFieldID] = &FieldDefinition{
					ID:                platformFieldID,
					ObjectID:          objectID,
					APIName:           "managed",
					IsPlatformManaged: true,
				}
			},
			wantErr:    true,
			wantErrMsg: "platform-managed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objRepo := newMockObjectRepo()
			fieldRepo := newMockFieldRepo()
			if tt.setup != nil {
				tt.setup(objRepo, fieldRepo)
			}

			ddlExec := newMockDDLExec()
			cache := newMockCache()
			svc := newFieldService(objRepo, fieldRepo, newMockPolymorphicRepo(), ddlExec, cache)
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
