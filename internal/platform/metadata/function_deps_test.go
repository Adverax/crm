package metadata

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFnReferences(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want []string
	}{
		{
			name: "no references",
			expr: "record.Amount > 100",
			want: nil,
		},
		{
			name: "single reference",
			expr: "fn.double(record.Amount) > 100",
			want: []string{"double"},
		},
		{
			name: "multiple references",
			expr: "fn.add(fn.double(x), fn.triple(y))",
			want: []string{"add", "double", "triple"},
		},
		{
			name: "duplicate references deduplicated",
			expr: "fn.calc(x) + fn.calc(y)",
			want: []string{"calc"},
		},
		{
			name: "reference with underscores",
			expr: "fn.my_long_func(x)",
			want: []string{"my_long_func"},
		},
		{
			name: "empty expression",
			expr: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractFnReferences(tt.expr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectCycles(t *testing.T) {
	tests := []struct {
		name      string
		functions []Function
		wantErr   bool
	}{
		{
			name:      "no functions",
			functions: nil,
			wantErr:   false,
		},
		{
			name: "no cycles",
			functions: []Function{
				{ID: uuid.New(), Name: "double", Body: "x * 2"},
				{ID: uuid.New(), Name: "quad", Body: "fn.double(fn.double(x))"},
			},
			wantErr: false,
		},
		{
			name: "direct cycle",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "fn.b(x)"},
				{ID: uuid.New(), Name: "b", Body: "fn.a(x)"},
			},
			wantErr: true,
		},
		{
			name: "indirect cycle",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "fn.b(x)"},
				{ID: uuid.New(), Name: "b", Body: "fn.c(x)"},
				{ID: uuid.New(), Name: "c", Body: "fn.a(x)"},
			},
			wantErr: true,
		},
		{
			name: "self reference",
			functions: []Function{
				{ID: uuid.New(), Name: "rec", Body: "fn.rec(x - 1)"},
			},
			wantErr: true,
		},
		{
			name: "reference to unknown function is ok",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "fn.unknown(x)"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DetectCycles(tt.functions)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDetectNestingDepth(t *testing.T) {
	tests := []struct {
		name      string
		functions []Function
		maxDepth  int
		wantErr   bool
	}{
		{
			name: "within limit",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "x * 2"},
				{ID: uuid.New(), Name: "b", Body: "fn.a(x)"},
				{ID: uuid.New(), Name: "c", Body: "fn.b(x)"},
			},
			maxDepth: 3,
			wantErr:  false,
		},
		{
			name: "exceeds limit",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "x * 2"},
				{ID: uuid.New(), Name: "b", Body: "fn.a(x)"},
				{ID: uuid.New(), Name: "c", Body: "fn.b(x)"},
				{ID: uuid.New(), Name: "d", Body: "fn.c(x)"},
			},
			maxDepth: 3,
			wantErr:  true,
		},
		{
			name: "single function depth 1",
			functions: []Function{
				{ID: uuid.New(), Name: "a", Body: "x * 2"},
			},
			maxDepth: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DetectNestingDepth(tt.functions, tt.maxDepth)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
