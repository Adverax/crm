package procedure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpressionResolver_ResolveString(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	resolver := NewExpressionResolver(cache)

	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "literal string passes through",
			expr: "hello",
			vars: map[string]any{},
			want: "hello",
		},
		{
			name: "resolves $.input.name",
			expr: "$.input.name",
			vars: map[string]any{"input": map[string]any{"name": "John"}},
			want: "John",
		},
		{
			name:    "error on invalid expression",
			expr:    "$.nonexistent_var",
			vars:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			execCtx := &ExecutionContext{
				Vars:     tt.vars,
				Deadline: time.Now().Add(MaxExecutionTimeout),
			}

			result, err := resolver.ResolveString(tt.expr, execCtx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExpressionResolver_ResolveMap(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	resolver := NewExpressionResolver(cache)

	execCtx := &ExecutionContext{
		Vars: map[string]any{
			"input": map[string]any{"email": "test@example.com"},
		},
		Deadline: time.Now().Add(MaxExecutionTimeout),
	}

	data := map[string]string{
		"greeting": "hello",
		"email":    "$.input.email",
	}

	result, err := resolver.ResolveMap(data, execCtx)
	require.NoError(t, err)
	assert.Equal(t, "hello", result["greeting"])
	assert.Equal(t, "test@example.com", result["email"])
}

func TestExpressionResolver_ResolveBool(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	resolver := NewExpressionResolver(cache)

	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    bool
		wantErr bool
	}{
		{
			name: "empty expression returns true",
			expr: "",
			vars: map[string]any{},
			want: true,
		},
		{
			name: "evaluates true expression",
			expr: "input.active == true",
			vars: map[string]any{"input": map[string]any{"active": true}},
			want: true,
		},
		{
			name: "evaluates false expression",
			expr: "input.active == true",
			vars: map[string]any{"input": map[string]any{"active": false}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			execCtx := &ExecutionContext{
				Vars:     tt.vars,
				Deadline: time.Now().Add(MaxExecutionTimeout),
			}

			result, err := resolver.ResolveBool(tt.expr, execCtx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
