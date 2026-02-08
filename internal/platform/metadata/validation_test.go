package metadata

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateObjectDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		obj     ObjectDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid custom object",
			obj: ObjectDefinition{
				APIName:     "Invoice__c",
				Label:       "Счёт",
				PluralLabel: "Счета",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: false,
		},
		{
			name: "valid standard object",
			obj: ObjectDefinition{
				APIName:     "Account",
				Label:       "Account",
				PluralLabel: "Accounts",
				ObjectType:  ObjectTypeStandard,
			},
			wantErr: false,
		},
		{
			name: "empty api_name",
			obj: ObjectDefinition{
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "api_name is required",
		},
		{
			name: "api_name too long",
			obj: ObjectDefinition{
				APIName:     string(make([]byte, 101)),
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "at most 100",
		},
		{
			name: "api_name starts with digit",
			obj: ObjectDefinition{
				APIName:     "1Invalid__c",
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "must start with a letter",
		},
		{
			name: "reserved word",
			obj: ObjectDefinition{
				APIName:     "select",
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeStandard,
			},
			wantErr: true,
			errMsg:  "reserved word",
		},
		{
			name: "custom without __c suffix",
			obj: ObjectDefinition{
				APIName:     "Invoice",
				Label:       "Счёт",
				PluralLabel: "Счета",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "must end with '__c'",
		},
		{
			name: "standard with __c suffix",
			obj: ObjectDefinition{
				APIName:     "Account__c",
				Label:       "Account",
				PluralLabel: "Accounts",
				ObjectType:  ObjectTypeStandard,
			},
			wantErr: true,
			errMsg:  "must not end with '__c'",
		},
		{
			name: "missing label",
			obj: ObjectDefinition{
				APIName:     "Test__c",
				PluralLabel: "Tests",
				ObjectType:  ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "label is required",
		},
		{
			name: "missing plural_label",
			obj: ObjectDefinition{
				APIName:    "Test__c",
				Label:      "Test",
				ObjectType: ObjectTypeCustom,
			},
			wantErr: true,
			errMsg:  "plural_label is required",
		},
		{
			name: "invalid object_type",
			obj: ObjectDefinition{
				APIName:     "Test",
				Label:       "Test",
				PluralLabel: "Tests",
				ObjectType:  "invalid",
			},
			wantErr: true,
			errMsg:  "object_type must be",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateObjectDefinition(&tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObjectDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); !contains(got, tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", got, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateFieldDefinition(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()
	refObjectID := uuid.New()

	subPlain := SubtypePlain
	subAssoc := SubtypeAssociation
	subComp := SubtypeComposition
	subPoly := SubtypePolymorphic

	maxLen := 100
	prec := 18
	scale := 2

	tests := []struct {
		name    string
		field   FieldDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text/plain field",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "first_name",
				Label:        "Имя",
				FieldType:    FieldTypeText,
				FieldSubtype: &subPlain,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			wantErr: false,
		},
		{
			name: "valid boolean field (no subtype)",
			field: FieldDefinition{
				ObjectID:  objectID,
				APIName:   "is_active",
				Label:     "Активен",
				FieldType: FieldTypeBoolean,
			},
			wantErr: false,
		},
		{
			name: "valid association field",
			field: FieldDefinition{
				ObjectID:           objectID,
				APIName:            "account_id",
				Label:              "Account",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subAssoc,
				ReferencedObjectID: &refObjectID,
				Config:             FieldConfig{RelationshipName: strPtr("Contacts")},
			},
			wantErr: false,
		},
		{
			name: "empty api_name",
			field: FieldDefinition{
				ObjectID:     objectID,
				Label:        "Test",
				FieldType:    FieldTypeText,
				FieldSubtype: &subPlain,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			wantErr: true,
			errMsg:  "api_name is required",
		},
		{
			name: "invalid type/subtype",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "bad_field",
				Label:        "Bad",
				FieldType:    FieldTypeText,
				FieldSubtype: &subAssoc,
			},
			wantErr: true,
			errMsg:  "invalid type/subtype",
		},
		{
			name: "missing required config (max_length for text/plain)",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "name",
				Label:        "Name",
				FieldType:    FieldTypeText,
				FieldSubtype: &subPlain,
				Config:       FieldConfig{},
			},
			wantErr: true,
			errMsg:  "config.max_length is required",
		},
		{
			name: "text/plain max_length too large",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "name",
				Label:        "Name",
				FieldType:    FieldTypeText,
				FieldSubtype: &subPlain,
				Config:       FieldConfig{MaxLength: intPtr(300)},
			},
			wantErr: true,
			errMsg:  "max_length must be between 1 and 255",
		},
		{
			name: "number/decimal missing precision",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "amount",
				Label:        "Amount",
				FieldType:    FieldTypeNumber,
				FieldSubtype: func() *FieldSubtype { s := SubtypeDecimal; return &s }(),
				Config:       FieldConfig{Scale: &scale},
			},
			wantErr: true,
			errMsg:  "config.precision is required",
		},
		{
			name: "association without referenced_object_id",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "account_id",
				Label:        "Account",
				FieldType:    FieldTypeReference,
				FieldSubtype: &subAssoc,
				Config:       FieldConfig{RelationshipName: strPtr("Contacts")},
			},
			wantErr: true,
			errMsg:  "referenced_object_id is required",
		},
		{
			name: "composition without referenced_object_id",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "deal_id",
				Label:        "Deal",
				FieldType:    FieldTypeReference,
				FieldSubtype: &subComp,
				Config:       FieldConfig{RelationshipName: strPtr("LineItems"), Precision: &prec},
			},
			wantErr: true,
			errMsg:  "referenced_object_id is required",
		},
		{
			name: "polymorphic with referenced_object_id",
			field: FieldDefinition{
				ObjectID:           objectID,
				APIName:            "what_id",
				Label:              "Related To",
				FieldType:          FieldTypeReference,
				FieldSubtype:       &subPoly,
				ReferencedObjectID: &refObjectID,
				Config:             FieldConfig{RelationshipName: strPtr("Activities")},
			},
			wantErr: true,
			errMsg:  "must be null for polymorphic",
		},
		{
			name: "reserved word as field api_name",
			field: FieldDefinition{
				ObjectID:     objectID,
				APIName:      "select",
				Label:        "Select",
				FieldType:    FieldTypeText,
				FieldSubtype: &subPlain,
				Config:       FieldConfig{MaxLength: &maxLen},
			},
			wantErr: true,
			errMsg:  "reserved word",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateFieldDefinition(&tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); !contains(got, tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", got, tt.errMsg)
				}
			}
		})
	}
}

func TestGenerateTableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		apiName string
		want    string
	}{
		{name: "custom object", apiName: "Invoice__c", want: "obj_invoice"},
		{name: "standard object", apiName: "Account", want: "obj_account"},
		{name: "mixed case", apiName: "CustomThing__c", want: "obj_customthing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GenerateTableName(tt.apiName); got != tt.want {
				t.Errorf("GenerateTableName(%q) = %q, want %q", tt.apiName, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func strPtr(s string) *string { return &s }
func intPtr(n int) *int       { return &n }
