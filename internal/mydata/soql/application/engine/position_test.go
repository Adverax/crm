package engine

import (
	"context"
	"errors"
	"testing"
)

// testMetadataProvider is a simple in-memory metadata provider for testing.
type testMetadataProvider struct {
	objects map[string]*ObjectMeta
}

func (p *testMetadataProvider) GetObject(ctx context.Context, name string) (*ObjectMeta, error) {
	obj, ok := p.objects[name]
	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (p *testMetadataProvider) ListObjects(ctx context.Context) ([]string, error) {
	names := make([]string, 0, len(p.objects))
	for name := range p.objects {
		names = append(names, name)
	}
	return names, nil
}

// TestParseErrorPositions tests that parse errors include correct positions.
func TestParseErrorPositions(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name         string
		query        string
		wantLine     int
		wantColRange [2]int // min, max column (sometimes parser position varies slightly)
	}{
		{
			name:         "missing field after SELECT",
			query:        "SELECT FROM Account",
			wantLine:     1,
			wantColRange: [2]int{7, 12}, // around "FROM"
		},
		{
			name:         "multiline error in WHERE",
			query:        "SELECT Name\nFROM Account\nWHERE = 'test'",
			wantLine:     3,
			wantColRange: [2]int{1, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Parse(tt.query)
			if err == nil {
				t.Fatal("expected parse error")
			}

			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("expected *ParseError, got %T: %v", err, err)
			}

			if pe.Pos.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", pe.Pos.Line, tt.wantLine)
			}

			if pe.Pos.Column < tt.wantColRange[0] || pe.Pos.Column > tt.wantColRange[1] {
				t.Errorf("Column = %d, want in range [%d, %d]", pe.Pos.Column, tt.wantColRange[0], tt.wantColRange[1])
			}
		})
	}
}

// TestValidationErrorPositions tests that validation errors include correct positions.
func TestValidationErrorPositions(t *testing.T) {
	metadata := &testMetadataProvider{
		objects: map[string]*ObjectMeta{
			"Account": {
				Name:      "Account",
				TableName: "accounts",
				Fields: map[string]*FieldMeta{
					"Id":   {Name: "Id", Column: "id", Type: FieldTypeString, Filterable: true, Sortable: true, Groupable: true},
					"Name": {Name: "Name", Column: "name", Type: FieldTypeString, Filterable: true, Sortable: true, Groupable: true},
				},
			},
		},
	}

	engine := NewEngine(
		WithMetadata(metadata),
	)

	tests := []struct {
		name         string
		query        string
		wantLine     int
		wantColRange [2]int
		wantCode     ValidationErrorCode
	}{
		{
			name:         "unknown field in SELECT",
			query:        "SELECT UnknownField FROM Account",
			wantLine:     1,
			wantColRange: [2]int{7, 12}, // "UnknownField" starts at column 8
			wantCode:     ErrCodeUnknownField,
		},
		{
			name:         "unknown field second position",
			query:        "SELECT Name, BadField FROM Account",
			wantLine:     1,
			wantColRange: [2]int{13, 17}, // "BadField" starts around column 14
			wantCode:     ErrCodeUnknownField,
		},
		{
			name:         "unknown field in WHERE",
			query:        "SELECT Name FROM Account WHERE BadField = 'test'",
			wantLine:     1,
			wantColRange: [2]int{31, 40}, // "BadField" in WHERE
			wantCode:     ErrCodeUnknownField,
		},
		{
			name:         "multiline unknown field in WHERE",
			query:        "SELECT Name\nFROM Account\nWHERE BadField = 'x'",
			wantLine:     3,
			wantColRange: [2]int{6, 14}, // "BadField" on line 3
			wantCode:     ErrCodeUnknownField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Prepare(context.Background(), tt.query)
			if err == nil {
				t.Fatal("expected validation error")
			}

			// Use errors.As to handle wrapped errors
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *ValidationError, got %T: %v", err, err)
			}

			if ve.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", ve.Code, tt.wantCode)
			}

			if ve.Pos.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", ve.Pos.Line, tt.wantLine)
			}

			if ve.Pos.Column < tt.wantColRange[0] || ve.Pos.Column > tt.wantColRange[1] {
				t.Errorf("Column = %d, want in range [%d, %d]", ve.Pos.Column, tt.wantColRange[0], tt.wantColRange[1])
			}
		})
	}
}
