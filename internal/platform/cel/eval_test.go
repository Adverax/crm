package cel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgramCache_EvaluateBool(t *testing.T) {
	env, err := StandardEnv()
	require.NoError(t, err)
	cache := NewProgramCache(env)

	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    bool
		wantErr bool
	}{
		{
			name: "simple true",
			expr: "true",
			vars: map[string]any{
				"record": map[string]any{},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: true,
		},
		{
			name: "simple false",
			expr: "false",
			vars: map[string]any{
				"record": map[string]any{},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: false,
		},
		{
			name: "record field access",
			expr: `record.Name == "Acme"`,
			vars: map[string]any{
				"record": map[string]any{"Name": "Acme"},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: true,
		},
		{
			name: "record field access false",
			expr: `record.Name == "Other"`,
			vars: map[string]any{
				"record": map[string]any{"Name": "Acme"},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: false,
		},
		{
			name: "user variable access",
			expr: `user.id == "user-123"`,
			vars: map[string]any{
				"record": map[string]any{},
				"old":    map[string]any{},
				"user":   map[string]any{"id": "user-123"},
				"now":    time.Now().UTC(),
			},
			want: true,
		},
		{
			name: "string size check",
			expr: `size(record.Name) > 0`,
			vars: map[string]any{
				"record": map[string]any{"Name": "Acme"},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: true,
		},
		{
			name: "string size check empty",
			expr: `size(record.Name) > 0`,
			vars: map[string]any{
				"record": map[string]any{"Name": ""},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: false,
		},
		{
			name: "comparison expression",
			expr: `record.Amount > 100`,
			vars: map[string]any{
				"record": map[string]any{"Amount": int64(200)},
				"old":    map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: true,
		},
		{
			name:    "compile error",
			expr:    `invalid syntax !!!`,
			vars:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cache.EvaluateBool(tt.expr, tt.vars)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgramCache_EvaluateAny(t *testing.T) {
	env, err := DefaultEnv()
	require.NoError(t, err)
	cache := NewProgramCache(env)

	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    any
		wantErr bool
	}{
		{
			name: "string result",
			expr: `"hello"`,
			vars: map[string]any{
				"record": map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: "hello",
		},
		{
			name: "user id",
			expr: `user.id`,
			vars: map[string]any{
				"record": map[string]any{},
				"user":   map[string]any{"id": "uid-001"},
				"now":    time.Now().UTC(),
			},
			want: "uid-001",
		},
		{
			name: "integer result",
			expr: `42`,
			vars: map[string]any{
				"record": map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: int64(42),
		},
		{
			name: "float result",
			expr: `3.14`,
			vars: map[string]any{
				"record": map[string]any{},
				"user":   map[string]any{},
				"now":    time.Now().UTC(),
			},
			want: 3.14,
		},
		{
			name:    "compile error returns error",
			expr:    `][`,
			vars:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cache.EvaluateAny(tt.expr, tt.vars)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgramCache_CacheHit(t *testing.T) {
	env, err := DefaultEnv()
	require.NoError(t, err)
	cache := NewProgramCache(env)

	vars := map[string]any{
		"record": map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	// First call compiles
	result1, err := cache.EvaluateAny(`"cached"`, vars)
	require.NoError(t, err)
	assert.Equal(t, "cached", result1)

	// Second call uses cache
	result2, err := cache.EvaluateAny(`"cached"`, vars)
	require.NoError(t, err)
	assert.Equal(t, "cached", result2)
}

func TestProgramCache_GetOrCompile_InvalidExpression(t *testing.T) {
	env, err := StandardEnv()
	require.NoError(t, err)
	cache := NewProgramCache(env)

	_, err = cache.GetOrCompile("this is not valid CEL !!!")
	require.Error(t, err)

	var compileErr *CompileError
	assert.ErrorAs(t, err, &compileErr)
	assert.Contains(t, compileErr.Expression, "this is not valid CEL")
}

func TestProgramCache_EvaluateBool_NonBoolResult(t *testing.T) {
	env, err := StandardEnv()
	require.NoError(t, err)
	cache := NewProgramCache(env)

	vars := map[string]any{
		"record": map[string]any{},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	_, err = cache.EvaluateBool(`"not a bool"`, vars)
	require.Error(t, err)

	var evalErr *EvalError
	assert.ErrorAs(t, err, &evalErr)
}
