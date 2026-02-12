package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileInsert(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	t.Run("single row insert", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech')")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Equal(t, OperationInsert, compiled.Operation)
		assert.Equal(t, "Account", compiled.Object)
		assert.Equal(t, "public.accounts", compiled.Table)
		assert.Contains(t, compiled.SQL, "INSERT INTO public.accounts")
		assert.Contains(t, compiled.SQL, "RETURNING id")
		assert.Len(t, compiled.Params, 2)
		assert.Equal(t, "Acme", compiled.Params[0])
		assert.Equal(t, "Tech", compiled.Params[1])
	})

	t.Run("multi-row insert", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech'), ('Globex', 'Finance')")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Len(t, compiled.Params, 4)
		assert.Contains(t, compiled.SQL, "$1")
		assert.Contains(t, compiled.SQL, "$4")
	})

	t.Run("insert with null", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Description) VALUES ('Acme', NULL)")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		// NULL is compiled as literal, not as a parameter
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "Acme", compiled.Params[0])
		assert.Contains(t, compiled.SQL, "NULL")
	})

	t.Run("insert with different types", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Contact (FirstName, Age, IsActive) VALUES ('John', 30, TRUE)")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Equal(t, "John", compiled.Params[0])
		assert.Equal(t, 30, compiled.Params[1])
		assert.Equal(t, true, compiled.Params[2])
	})
}

func TestCompileUpdate(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	t.Run("update with where", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET Status = 'Active' WHERE Email = 'test@example.com'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Equal(t, OperationUpdate, compiled.Operation)
		assert.Contains(t, compiled.SQL, "UPDATE public.contacts SET")
		assert.Contains(t, compiled.SQL, "status = $1")
		assert.Contains(t, compiled.SQL, "WHERE")
		assert.Contains(t, compiled.SQL, "RETURNING id")
		assert.Len(t, compiled.Params, 2)
		assert.Equal(t, "Active", compiled.Params[0])
		assert.Equal(t, "test@example.com", compiled.Params[1])
	})

	t.Run("update without where", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET Status = 'Active'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.NotContains(t, compiled.SQL, "WHERE")
		assert.Len(t, compiled.Params, 1)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET FirstName = 'John', LastName = 'Doe'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "first_name = $1")
		assert.Contains(t, compiled.SQL, "last_name = $2")
		assert.Len(t, compiled.Params, 2)
	})
}

func TestCompileDelete(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, &NoLimits)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	t.Run("delete with where", func(t *testing.T) {
		ast, err := Parse("DELETE FROM Task WHERE Status = 'Completed'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Equal(t, OperationDelete, compiled.Operation)
		assert.Contains(t, compiled.SQL, "DELETE FROM public.tasks")
		assert.Contains(t, compiled.SQL, "WHERE")
		assert.Contains(t, compiled.SQL, "RETURNING id")
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "Completed", compiled.Params[0])
	})

	t.Run("delete with complex where", func(t *testing.T) {
		ast, err := Parse("DELETE FROM Task WHERE Status = 'Completed' AND Priority = 'Low'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "AND")
		assert.Len(t, compiled.Params, 2)
	})

	t.Run("delete with or condition", func(t *testing.T) {
		ast, err := Parse("DELETE FROM Task WHERE Status = 'Completed' OR Status = 'Cancelled'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "OR")
	})
}

func TestCompileUpsert(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	t.Run("basic upsert", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (ExternalId, Name, Industry) VALUES ('ext-001', 'Acme', 'Tech') ON ExternalId")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Equal(t, OperationUpsert, compiled.Operation)
		assert.Contains(t, compiled.SQL, "INSERT INTO public.accounts")
		assert.Contains(t, compiled.SQL, "ON CONFLICT (external_id)")
		assert.Contains(t, compiled.SQL, "DO UPDATE SET")
		assert.Contains(t, compiled.SQL, "name = EXCLUDED.name")
		assert.Contains(t, compiled.SQL, "industry = EXCLUDED.industry")
		// External ID should not be in DO UPDATE SET (except for no-op case)
		assert.NotContains(t, compiled.SQL, "external_id = EXCLUDED.external_id")
		assert.Contains(t, compiled.SQL, "RETURNING id")
		assert.Len(t, compiled.Params, 3)
	})

	t.Run("multi-row upsert", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (ExternalId, Name) VALUES ('ext-001', 'Acme'), ('ext-002', 'Globex') ON ExternalId")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Len(t, compiled.Params, 4)
		assert.Contains(t, compiled.SQL, "$1")
		assert.Contains(t, compiled.SQL, "$4")
	})
}

func TestCompileWhereExpressions(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, &NoLimits)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name         string
		input        string
		wantSQLPart  string
		wantParamCnt int
	}{
		{
			name:         "equality",
			input:        "DELETE FROM Task WHERE Status = 'Active'",
			wantSQLPart:  "status = $1",
			wantParamCnt: 1,
		},
		{
			name:         "not equal",
			input:        "DELETE FROM Task WHERE Status != 'Active'",
			wantSQLPart:  "status != $1",
			wantParamCnt: 1,
		},
		{
			name:         "greater than",
			input:        "DELETE FROM Contact WHERE Age > 30",
			wantSQLPart:  "age > $1",
			wantParamCnt: 1,
		},
		{
			name:         "is null",
			input:        "DELETE FROM Task WHERE Subject IS NULL",
			wantSQLPart:  "subject IS NULL",
			wantParamCnt: 0,
		},
		{
			name:         "is not null",
			input:        "DELETE FROM Task WHERE Subject IS NOT NULL",
			wantSQLPart:  "subject IS NOT NULL",
			wantParamCnt: 0,
		},
		{
			name:         "in clause",
			input:        "DELETE FROM Task WHERE Status IN ('A', 'B')",
			wantSQLPart:  "status IN ($1, $2)",
			wantParamCnt: 2,
		},
		{
			name:         "not in clause",
			input:        "DELETE FROM Task WHERE Status NOT IN ('X', 'Y')",
			wantSQLPart:  "status NOT IN ($1, $2)",
			wantParamCnt: 2,
		},
		{
			name:         "like",
			input:        "DELETE FROM Task WHERE Subject LIKE 'Test%'",
			wantSQLPart:  "subject LIKE $1",
			wantParamCnt: 1,
		},
		{
			name:         "not like",
			input:        "DELETE FROM Task WHERE Subject NOT LIKE '%test%'",
			wantSQLPart:  "subject NOT LIKE $1",
			wantParamCnt: 1,
		},
		{
			name:         "and",
			input:        "DELETE FROM Task WHERE Status = 'A' AND Priority = 'B'",
			wantSQLPart:  "AND",
			wantParamCnt: 2,
		},
		{
			name:         "or",
			input:        "DELETE FROM Task WHERE Status = 'A' OR Status = 'B'",
			wantSQLPart:  "OR",
			wantParamCnt: 2,
		},
		{
			name:         "not",
			input:        "DELETE FROM Task WHERE NOT Status = 'Active'",
			wantSQLPart:  "NOT",
			wantParamCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			require.NoError(t, err)

			validated, err := validator.Validate(ctx, ast)
			require.NoError(t, err)

			compiled, err := compiler.Compile(validated)
			require.NoError(t, err)

			assert.Contains(t, compiled.SQL, tt.wantSQLPart)
			assert.Len(t, compiled.Params, tt.wantParamCnt)
		})
	}
}

func TestEngineIntegration(t *testing.T) {
	metadata := newTestMetadata()
	engine := NewEngine(
		WithMetadata(metadata),
		WithLimits(&DefaultLimits),
	)

	t.Run("prepare insert", func(t *testing.T) {
		compiled, err := engine.Prepare(context.Background(), "INSERT INTO Account (Name) VALUES ('Test')")
		require.NoError(t, err)
		assert.Equal(t, OperationInsert, compiled.Operation)
		assert.NotEmpty(t, compiled.SQL)
	})

	t.Run("prepare update", func(t *testing.T) {
		compiled, err := engine.Prepare(context.Background(), "UPDATE Contact SET Status = 'Active' WHERE Email = 'test@test.com'")
		require.NoError(t, err)
		assert.Equal(t, OperationUpdate, compiled.Operation)
	})

	t.Run("prepare delete", func(t *testing.T) {
		compiled, err := engine.Prepare(context.Background(), "DELETE FROM Task WHERE Status = 'Completed'")
		require.NoError(t, err)
		assert.Equal(t, OperationDelete, compiled.Operation)
	})

	t.Run("prepare upsert", func(t *testing.T) {
		compiled, err := engine.Prepare(context.Background(), "UPSERT Account (ExternalId, Name) VALUES ('ext-1', 'Test') ON ExternalId")
		require.NoError(t, err)
		assert.Equal(t, OperationUpsert, compiled.Operation)
	})

	t.Run("parse error", func(t *testing.T) {
		_, err := engine.Prepare(context.Background(), "INVALID SYNTAX")
		require.Error(t, err)
		assert.True(t, IsParseError(err))
	})

	t.Run("validation error", func(t *testing.T) {
		_, err := engine.Prepare(context.Background(), "INSERT INTO Unknown (Name) VALUES ('Test')")
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
	})
}

func TestCompileFunctions(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	t.Run("insert with UPPER function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (UPPER('acme'))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "UPPER($1)")
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "acme", compiled.Params[0])
	})

	t.Run("insert with LOWER function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (LOWER('ACME'))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "LOWER($1)")
		assert.Equal(t, "ACME", compiled.Params[0])
	})

	t.Run("insert with TRIM function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (TRIM('  Acme  '))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "TRIM($1)")
		assert.Equal(t, "  Acme  ", compiled.Params[0])
	})

	t.Run("insert with COALESCE function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (COALESCE(NULL, 'Default'))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "COALESCE(NULL, $1)")
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "Default", compiled.Params[0])
	})

	t.Run("insert with CONCAT function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (CONCAT('Hello', ' ', 'World'))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "CONCAT($1, $2, $3)")
		assert.Len(t, compiled.Params, 3)
		assert.Equal(t, "Hello", compiled.Params[0])
		assert.Equal(t, " ", compiled.Params[1])
		assert.Equal(t, "World", compiled.Params[2])
	})

	t.Run("insert with nested functions", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (UPPER(TRIM('  acme  ')))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "UPPER(TRIM($1))")
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "  acme  ", compiled.Params[0])
	})

	t.Run("update with function", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET FirstName = UPPER('john')")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "first_name = UPPER($1)")
		assert.Len(t, compiled.Params, 1)
		assert.Equal(t, "john", compiled.Params[0])
	})

	t.Run("insert with LENGTH function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Contact (FirstName, Age) VALUES ('John', LENGTH('Hello'))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "LENGTH($2)")
		assert.Equal(t, "John", compiled.Params[0])
		assert.Equal(t, "Hello", compiled.Params[1])
	})

	t.Run("insert with SUBSTRING function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (SUBSTRING('Hello World', 1, 5))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "SUBSTRING($1, $2, $3)")
		assert.Len(t, compiled.Params, 3)
		assert.Equal(t, "Hello World", compiled.Params[0])
		assert.Equal(t, 1, compiled.Params[1])
		assert.Equal(t, 5, compiled.Params[2])
	})

	t.Run("insert with ABS function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Contact (FirstName, Age) VALUES ('John', ABS(25))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "ABS($2)")
		assert.Equal(t, "John", compiled.Params[0])
		assert.Equal(t, 25, compiled.Params[1])
	})

	t.Run("update with ROUND function", func(t *testing.T) {
		// Use UPDATE to bypass type checking since we're testing SQL generation
		ast, err := Parse("UPDATE Account SET Name = CONCAT('Value: ', ROUND(25))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "ROUND($2)")
		assert.Contains(t, compiled.SQL, "CONCAT($1, ROUND($2))")
	})

	t.Run("insert with FLOOR function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Contact (FirstName, Age) VALUES ('John', FLOOR(25))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "FLOOR($2)")
	})

	t.Run("insert with CEIL function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Contact (FirstName, Age) VALUES ('John', CEIL(25))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "CEIL($2)")
	})

	t.Run("insert with NULLIF function", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES (NULLIF('', ''))")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "NULLIF($1, $2)")
		assert.Len(t, compiled.Params, 2)
	})

	t.Run("upsert with function", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (ExternalId, Name) VALUES ('ext-1', UPPER('test')) ON ExternalId")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)

		compiled, err := compiler.Compile(validated)
		require.NoError(t, err)

		assert.Contains(t, compiled.SQL, "UPPER($2)")
		assert.Len(t, compiled.Params, 2)
		assert.Equal(t, "ext-1", compiled.Params[0])
		assert.Equal(t, "test", compiled.Params[1])
	})
}
