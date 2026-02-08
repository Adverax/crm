package ddl

import (
	"testing"
)

func TestMapFieldToColumn(t *testing.T) {
	t.Parallel()

	intPtr := func(n int) *int { return &n }
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name         string
		field        FieldInfo
		wantCount    int
		wantDataType string
		wantNotNull  bool
	}{
		{
			name: "text/plain VARCHAR(100)",
			field: FieldInfo{
				APIName:      "first_name",
				FieldType:    "text",
				FieldSubtype: "plain",
				MaxLength:    intPtr(100),
			},
			wantCount:    1,
			wantDataType: "VARCHAR(100)",
		},
		{
			name: "text/area TEXT",
			field: FieldInfo{
				APIName:      "notes",
				FieldType:    "text",
				FieldSubtype: "area",
			},
			wantCount:    1,
			wantDataType: "TEXT",
		},
		{
			name: "text/email VARCHAR(255)",
			field: FieldInfo{
				APIName:      "email",
				FieldType:    "text",
				FieldSubtype: "email",
			},
			wantCount:    1,
			wantDataType: "VARCHAR(255)",
		},
		{
			name: "text/phone VARCHAR(40)",
			field: FieldInfo{
				APIName:      "phone",
				FieldType:    "text",
				FieldSubtype: "phone",
			},
			wantCount:    1,
			wantDataType: "VARCHAR(40)",
		},
		{
			name: "text/url VARCHAR(2048)",
			field: FieldInfo{
				APIName:      "website",
				FieldType:    "text",
				FieldSubtype: "url",
			},
			wantCount:    1,
			wantDataType: "VARCHAR(2048)",
		},
		{
			name: "number/integer NUMERIC(18,0)",
			field: FieldInfo{
				APIName:      "count",
				FieldType:    "number",
				FieldSubtype: "integer",
				Precision:    intPtr(18),
			},
			wantCount:    1,
			wantDataType: "NUMERIC(18,0)",
		},
		{
			name: "number/decimal NUMERIC(10,3)",
			field: FieldInfo{
				APIName:      "amount",
				FieldType:    "number",
				FieldSubtype: "decimal",
				Precision:    intPtr(10),
				Scale:        intPtr(3),
			},
			wantCount:    1,
			wantDataType: "NUMERIC(10,3)",
		},
		{
			name: "number/currency NUMERIC(18,2)",
			field: FieldInfo{
				APIName:      "price",
				FieldType:    "number",
				FieldSubtype: "currency",
			},
			wantCount:    1,
			wantDataType: "NUMERIC(18,2)",
		},
		{
			name: "number/percent NUMERIC(5,2)",
			field: FieldInfo{
				APIName:      "discount",
				FieldType:    "number",
				FieldSubtype: "percent",
			},
			wantCount:    1,
			wantDataType: "NUMERIC(5,2)",
		},
		{
			name: "number/auto_number IDENTITY",
			field: FieldInfo{
				APIName:      "seq",
				FieldType:    "number",
				FieldSubtype: "auto_number",
			},
			wantCount:    1,
			wantDataType: "INTEGER GENERATED ALWAYS AS IDENTITY",
		},
		{
			name: "boolean BOOLEAN",
			field: FieldInfo{
				APIName:   "is_paid",
				FieldType: "boolean",
			},
			wantCount:    1,
			wantDataType: "BOOLEAN",
		},
		{
			name: "datetime/date DATE",
			field: FieldInfo{
				APIName:      "birth_date",
				FieldType:    "datetime",
				FieldSubtype: "date",
			},
			wantCount:    1,
			wantDataType: "DATE",
		},
		{
			name: "datetime/datetime TIMESTAMPTZ",
			field: FieldInfo{
				APIName:      "start_datetime",
				FieldType:    "datetime",
				FieldSubtype: "datetime",
			},
			wantCount:    1,
			wantDataType: "TIMESTAMPTZ",
		},
		{
			name: "datetime/time TIME",
			field: FieldInfo{
				APIName:      "start_time",
				FieldType:    "datetime",
				FieldSubtype: "time",
			},
			wantCount:    1,
			wantDataType: "TIME",
		},
		{
			name: "picklist/single VARCHAR(255)",
			field: FieldInfo{
				APIName:      "status",
				FieldType:    "picklist",
				FieldSubtype: "single",
			},
			wantCount:    1,
			wantDataType: "VARCHAR(255)",
		},
		{
			name: "picklist/multi TEXT[]",
			field: FieldInfo{
				APIName:      "tags",
				FieldType:    "picklist",
				FieldSubtype: "multi",
			},
			wantCount:    1,
			wantDataType: "TEXT[]",
		},
		{
			name: "reference/association UUID",
			field: FieldInfo{
				APIName:      "account_id",
				FieldType:    "reference",
				FieldSubtype: "association",
			},
			wantCount:    1,
			wantDataType: "UUID",
			wantNotNull:  false,
		},
		{
			name: "reference/composition UUID NOT NULL",
			field: FieldInfo{
				APIName:      "deal_id",
				FieldType:    "reference",
				FieldSubtype: "composition",
			},
			wantCount:    1,
			wantDataType: "UUID",
			wantNotNull:  true,
		},
		{
			name: "reference/polymorphic two columns",
			field: FieldInfo{
				APIName:      "what",
				FieldType:    "reference",
				FieldSubtype: "polymorphic",
			},
			wantCount: 2,
		},
		{
			name: "association with restrict",
			field: FieldInfo{
				APIName:      "parent_id",
				FieldType:    "reference",
				FieldSubtype: "association",
				OnDelete:     strPtr("restrict"),
			},
			wantCount:    1,
			wantDataType: "UUID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cols, err := MapFieldToColumn(tt.field)
			if err != nil {
				t.Fatalf("MapFieldToColumn() error = %v", err)
			}
			if len(cols) != tt.wantCount {
				t.Fatalf("MapFieldToColumn() returned %d columns, want %d", len(cols), tt.wantCount)
			}
			if tt.wantCount == 1 {
				if cols[0].DataType != tt.wantDataType {
					t.Errorf("DataType = %q, want %q", cols[0].DataType, tt.wantDataType)
				}
				if cols[0].NotNull != tt.wantNotNull {
					t.Errorf("NotNull = %v, want %v", cols[0].NotNull, tt.wantNotNull)
				}
			}
		})
	}
}
