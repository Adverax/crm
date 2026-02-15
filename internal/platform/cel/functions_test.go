package cel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionRegistry_SimpleFunction(t *testing.T) {
	defs := []FunctionDef{
		{
			Name:       "double",
			Params:     []ParamDef{{Name: "x", Type: "number"}},
			ReturnType: "number",
			Body:       "x * 2",
		},
	}

	registry, err := NewFunctionRegistry(defs)
	require.NoError(t, err)

	env, err := StandardEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{"Amount": int64(50)},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	result, err := cache.EvaluateAny("fn.double(record.Amount)", vars)
	require.NoError(t, err)
	assert.Equal(t, int64(100), result)
}

func TestFunctionRegistry_MultiParam(t *testing.T) {
	defs := []FunctionDef{
		{
			Name: "add",
			Params: []ParamDef{
				{Name: "a", Type: "number"},
				{Name: "b", Type: "number"},
			},
			ReturnType: "number",
			Body:       "a + b",
		},
	}

	registry, err := NewFunctionRegistry(defs)
	require.NoError(t, err)

	env, err := StandardEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	result, err := cache.EvaluateAny("fn.add(10, 32)", vars)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestFunctionRegistry_StringFunction(t *testing.T) {
	defs := []FunctionDef{
		{
			Name:       "greet",
			Params:     []ParamDef{{Name: "name", Type: "string"}},
			ReturnType: "string",
			Body:       `"Hello, " + name + "!"`,
		},
	}

	registry, err := NewFunctionRegistry(defs)
	require.NoError(t, err)

	env, err := DefaultEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	result, err := cache.EvaluateAny(`fn.greet("World")`, vars)
	require.NoError(t, err)
	assert.Equal(t, "World", "World")
	assert.Equal(t, "Hello, World!", result)
}

func TestFunctionRegistry_BoolFunction(t *testing.T) {
	defs := []FunctionDef{
		{
			Name:       "is_positive",
			Params:     []ParamDef{{Name: "x", Type: "number"}},
			ReturnType: "boolean",
			Body:       "x > 0",
		},
	}

	registry, err := NewFunctionRegistry(defs)
	require.NoError(t, err)

	env, err := StandardEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{
			name: "positive number returns true",
			expr: "fn.is_positive(42)",
			want: true,
		},
		{
			name: "negative number returns false",
			expr: "fn.is_positive(-1)",
			want: false,
		},
		{
			name: "used in validation rule expression",
			expr: "fn.is_positive(record.Amount)",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars["record"] = map[string]any{"Amount": int64(-5)}
			result, err := cache.EvaluateBool(tt.expr, vars)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFunctionRegistry_InvalidBody(t *testing.T) {
	defs := []FunctionDef{
		{
			Name:       "bad",
			Params:     []ParamDef{{Name: "x", Type: "number"}},
			ReturnType: "number",
			Body:       "invalid !!!! syntax",
		},
	}

	_, err := NewFunctionRegistry(defs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad")
}

func TestFunctionRegistry_EmptyRegistry(t *testing.T) {
	registry, err := NewFunctionRegistry(nil)
	require.NoError(t, err)

	env, err := StandardEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	result, err := cache.EvaluateBool("true", vars)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestFunctionRegistry_WrongArgCount(t *testing.T) {
	defs := []FunctionDef{
		{
			Name: "add",
			Params: []ParamDef{
				{Name: "a", Type: "number"},
				{Name: "b", Type: "number"},
			},
			ReturnType: "number",
			Body:       "a + b",
		},
	}

	registry, err := NewFunctionRegistry(defs)
	require.NoError(t, err)

	env, err := StandardEnvWithFunctions(registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"record": map[string]any{},
		"old":    map[string]any{},
		"user":   map[string]any{},
		"now":    time.Now().UTC(),
	}

	// Calling with wrong number of args should fail at compile time
	_, err = cache.EvaluateAny("fn.add(1)", vars)
	assert.Error(t, err)
}

func TestFunctionBodyEnv(t *testing.T) {
	registry, err := NewFunctionRegistry(nil)
	require.NoError(t, err)

	params := []ParamDef{
		{Name: "x", Type: "number"},
		{Name: "y", Type: "string"},
	}

	env, err := FunctionBodyEnv(params, registry)
	require.NoError(t, err)

	cache := NewProgramCache(env)
	vars := map[string]any{
		"x": int64(10),
		"y": "hello",
	}

	result, err := cache.EvaluateAny("x * 2", vars)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result)
}
