package dml

import (
	"context"
	"testing"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCELDefaultResolver_ResolveDefaults(t *testing.T) {
	resolver, err := NewCELDefaultResolver(nil)
	require.NoError(t, err)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		object         *engine.ObjectMeta
		operation      engine.Operation
		providedFields []string
		wantFields     []string
		wantErr        bool
	}{
		{
			name: "static default on create",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					DefaultValue: strPtr("New"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     []string{"Status"},
		},
		{
			name: "skip provided field",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					DefaultValue: strPtr("New"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{"Status"},
			wantFields:     nil,
		},
		{
			name: "skip read-only field",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					ReadOnly: true, DefaultValue: strPtr("New"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     nil,
		},
		{
			name: "integer static default",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Priority", Column: "priority", Type: engine.FieldTypeInteger,
					DefaultValue: strPtr("5"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     []string{"Priority"},
		},
		{
			name: "boolean static default",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "IsActive", Column: "is_active", Type: engine.FieldTypeBoolean,
					DefaultValue: strPtr("true"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     []string{"IsActive"},
		},
		{
			name: "default_on=update only applies on update",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					DefaultValue: strPtr("Updated"), DefaultOn: strPtr("update"),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     nil,
		},
		{
			name: "default_on=update applies on update",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					DefaultValue: strPtr("Updated"), DefaultOn: strPtr("update"),
				}).
				Build(),
			operation:      engine.OperationUpdate,
			providedFields: []string{},
			wantFields:     []string{"Status"},
		},
		{
			name: "default_on=create,update applies on both",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Status", Column: "status", Type: engine.FieldTypeString,
					DefaultValue: strPtr("Active"), DefaultOn: strPtr("create,update"),
				}).
				Build(),
			operation:      engine.OperationUpdate,
			providedFields: []string{},
			wantFields:     []string{"Status"},
		},
		{
			name: "CEL expression default",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Note", Column: "note", Type: engine.FieldTypeString,
					DefaultExpr: strPtr(`"auto-generated"`),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantFields:     []string{"Note"},
		},
		{
			name: "invalid CEL expression returns error",
			object: engine.NewObjectMeta("Account", "obj_account").
				FieldFull(&engine.FieldMeta{
					Name: "Note", Column: "note", Type: engine.FieldTypeString,
					DefaultExpr: strPtr(`invalid!!`),
				}).
				Build(),
			operation:      engine.OperationInsert,
			providedFields: []string{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults, err := resolver.ResolveDefaults(context.Background(), tt.object, tt.operation, tt.providedFields)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.wantFields == nil {
				assert.Empty(t, defaults)
			} else {
				for _, fieldName := range tt.wantFields {
					assert.Contains(t, defaults, fieldName)
				}
			}
		})
	}
}

func TestConvertStaticDefault(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldType engine.FieldType
		want      any
	}{
		{name: "string", value: "hello", fieldType: engine.FieldTypeString, want: "hello"},
		{name: "integer", value: "42", fieldType: engine.FieldTypeInteger, want: 42},
		{name: "float", value: "3.14", fieldType: engine.FieldTypeFloat, want: 3.14},
		{name: "boolean true", value: "true", fieldType: engine.FieldTypeBoolean, want: true},
		{name: "boolean false", value: "false", fieldType: engine.FieldTypeBoolean, want: false},
		{name: "invalid integer falls back to string", value: "abc", fieldType: engine.FieldTypeInteger, want: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertStaticDefault(tt.value, tt.fieldType)
			assert.Equal(t, tt.want, got)
		})
	}
}
