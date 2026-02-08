package metadata

import (
	"testing"

	"github.com/google/uuid"
)

func TestCompositionChecker_ValidateNewComposition(t *testing.T) {
	t.Parallel()

	objA := uuid.New()
	objB := uuid.New()
	objC := uuid.New()
	objD := uuid.New()

	subComp := SubtypeComposition
	subAssoc := SubtypeAssociation

	tests := []struct {
		name           string
		existingFields []FieldDefinition
		newField       FieldDefinition
		allObjects     map[uuid.UUID]ObjectDefinition
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid simple composition A->B",
			existingFields: nil,
			newField: FieldDefinition{
				ObjectID:           objB,
				APIName:            "a_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subComp,
				ReferencedObjectID: &objA,
			},
			wantErr: false,
		},
		{
			name: "valid chain A->B->C (depth 2)",
			existingFields: []FieldDefinition{
				{ObjectID: objB, FieldType: FieldTypeReference, FieldSubtype: &subComp, ReferencedObjectID: &objA},
			},
			newField: FieldDefinition{
				ObjectID:           objC,
				APIName:            "b_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subComp,
				ReferencedObjectID: &objB,
			},
			wantErr: false,
		},
		{
			name: "rejects chain A->B->C->D (depth 3 exceeds max 2)",
			existingFields: []FieldDefinition{
				{ObjectID: objB, FieldType: FieldTypeReference, FieldSubtype: &subComp, ReferencedObjectID: &objA},
				{ObjectID: objC, FieldType: FieldTypeReference, FieldSubtype: &subComp, ReferencedObjectID: &objB},
			},
			newField: FieldDefinition{
				ObjectID:           objD,
				APIName:            "c_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subComp,
				ReferencedObjectID: &objC,
			},
			wantErr: true,
			errMsg:  "composition chain depth would exceed",
		},
		{
			name:           "rejects self-reference composition",
			existingFields: nil,
			newField: FieldDefinition{
				ObjectID:           objA,
				APIName:            "parent_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subComp,
				ReferencedObjectID: &objA,
			},
			wantErr: true,
			errMsg:  "self-reference is not allowed",
		},
		{
			name: "rejects cycle B->A when A->B exists (caught by depth or cycle)",
			existingFields: []FieldDefinition{
				{ObjectID: objB, FieldType: FieldTypeReference, FieldSubtype: &subComp, ReferencedObjectID: &objA},
			},
			newField: FieldDefinition{
				ObjectID:           objA,
				APIName:            "b_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subComp,
				ReferencedObjectID: &objB,
			},
			wantErr: true,
			errMsg:  "composition",
		},
		{
			name:           "skip non-composition fields",
			existingFields: nil,
			newField: FieldDefinition{
				ObjectID:           objB,
				APIName:            "a_id",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subAssoc,
				ReferencedObjectID: &objA,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checker := NewCompositionChecker(tt.existingFields)
			err := checker.ValidateNewComposition(tt.newField, tt.allObjects)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNewComposition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); !contains(got, tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", got, tt.errMsg)
				}
			}
		})
	}
}
