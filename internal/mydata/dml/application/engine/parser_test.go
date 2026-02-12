package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInsert(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantObj   string
		wantCols  []string
		wantRows  int
		wantErr   bool
	}{
		{
			name:     "simple insert",
			input:    "INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech')",
			wantObj:  "Account",
			wantCols: []string{"Name", "Industry"},
			wantRows: 1,
		},
		{
			name:     "multi-row insert",
			input:    "INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech'), ('Globex', 'Finance')",
			wantObj:  "Account",
			wantCols: []string{"Name", "Industry"},
			wantRows: 2,
		},
		{
			name:     "insert with integers",
			input:    "INSERT INTO Contact (FirstName, Age) VALUES ('John', 30)",
			wantObj:  "Contact",
			wantCols: []string{"FirstName", "Age"},
			wantRows: 1,
		},
		{
			name:     "insert with null",
			input:    "INSERT INTO Account (Name, Description) VALUES ('Acme', NULL)",
			wantObj:  "Account",
			wantCols: []string{"Name", "Description"},
			wantRows: 1,
		},
		{
			name:     "insert with boolean",
			input:    "INSERT INTO Contact (Name, IsActive) VALUES ('John', TRUE)",
			wantObj:  "Contact",
			wantCols: []string{"Name", "IsActive"},
			wantRows: 1,
		},
		{
			name:     "insert with date",
			input:    "INSERT INTO Task (Subject, DueDate) VALUES ('Call', 2024-01-15)",
			wantObj:  "Task",
			wantCols: []string{"Subject", "DueDate"},
			wantRows: 1,
		},
		{
			name:     "insert with datetime",
			input:    "INSERT INTO Event (Subject, StartTime) VALUES ('Meeting', 2024-01-15T10:30:00Z)",
			wantObj:  "Event",
			wantCols: []string{"Subject", "StartTime"},
			wantRows: 1,
		},
		{
			name:     "insert with float",
			input:    "INSERT INTO Product (Name, Price) VALUES ('Widget', 99.99)",
			wantObj:  "Product",
			wantCols: []string{"Name", "Price"},
			wantRows: 1,
		},
		{
			name:     "insert case insensitive",
			input:    "insert into Account (Name) values ('Test')",
			wantObj:  "Account",
			wantCols: []string{"Name"},
			wantRows: 1,
		},
		{
			name:    "invalid - missing values",
			input:   "INSERT INTO Account (Name)",
			wantErr: true,
		},
		{
			name:    "invalid - missing columns",
			input:   "INSERT INTO Account VALUES ('Test')",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ast)
			require.NotNil(t, ast.Insert)

			assert.Equal(t, tt.wantObj, ast.Insert.Object)
			assert.Equal(t, tt.wantCols, ast.Insert.Fields)
			assert.Len(t, ast.Insert.Values, tt.wantRows)
		})
	}
}

func TestParseUpdate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantObj        string
		wantAssigns    int
		wantHasWhere   bool
		wantErr        bool
	}{
		{
			name:         "simple update",
			input:        "UPDATE Contact SET Status = 'Active'",
			wantObj:      "Contact",
			wantAssigns:  1,
			wantHasWhere: false,
		},
		{
			name:         "update with where",
			input:        "UPDATE Contact SET Status = 'Active' WHERE AccountId = 'acc-001'",
			wantObj:      "Contact",
			wantAssigns:  1,
			wantHasWhere: true,
		},
		{
			name:         "update multiple fields",
			input:        "UPDATE Contact SET Status = 'Active', UpdatedAt = 2024-01-15",
			wantObj:      "Contact",
			wantAssigns:  2,
			wantHasWhere: false,
		},
		{
			name:         "update with complex where",
			input:        "UPDATE Contact SET Status = 'Active' WHERE AccountId = 'acc-001' AND Status = 'Pending'",
			wantObj:      "Contact",
			wantAssigns:  1,
			wantHasWhere: true,
		},
		{
			name:         "update with null",
			input:        "UPDATE Contact SET Description = NULL WHERE Id = 'con-001'",
			wantObj:      "Contact",
			wantAssigns:  1,
			wantHasWhere: true,
		},
		{
			name:         "update case insensitive",
			input:        "update Contact set Name = 'Test' where Id = '123'",
			wantObj:      "Contact",
			wantAssigns:  1,
			wantHasWhere: true,
		},
		{
			name:    "invalid - no set clause",
			input:   "UPDATE Contact WHERE Id = '123'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ast)
			require.NotNil(t, ast.Update)

			assert.Equal(t, tt.wantObj, ast.Update.Object)
			assert.Len(t, ast.Update.Assignments, tt.wantAssigns)
			if tt.wantHasWhere {
				assert.NotNil(t, ast.Update.Where)
			} else {
				assert.Nil(t, ast.Update.Where)
			}
		})
	}
}

func TestParseDelete(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantObj      string
		wantHasWhere bool
		wantErr      bool
	}{
		{
			name:         "simple delete",
			input:        "DELETE FROM Task",
			wantObj:      "Task",
			wantHasWhere: false,
		},
		{
			name:         "delete with where",
			input:        "DELETE FROM Task WHERE Status = 'Completed'",
			wantObj:      "Task",
			wantHasWhere: true,
		},
		{
			name:         "delete with complex where",
			input:        "DELETE FROM Task WHERE Status = 'Completed' AND CreatedDate < 2023-01-01",
			wantObj:      "Task",
			wantHasWhere: true,
		},
		{
			name:         "delete with or condition",
			input:        "DELETE FROM Task WHERE Status = 'Completed' OR Status = 'Cancelled'",
			wantObj:      "Task",
			wantHasWhere: true,
		},
		{
			name:         "delete case insensitive",
			input:        "delete from Task where Id = '123'",
			wantObj:      "Task",
			wantHasWhere: true,
		},
		{
			name:    "invalid - missing from",
			input:   "DELETE Task WHERE Id = '123'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ast)
			require.NotNil(t, ast.Delete)

			assert.Equal(t, tt.wantObj, ast.Delete.Object)
			if tt.wantHasWhere {
				assert.NotNil(t, ast.Delete.Where)
			} else {
				assert.Nil(t, ast.Delete.Where)
			}
		})
	}
}

func TestParseUpsert(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantObj    string
		wantCols   []string
		wantRows   int
		wantExtId  string
		wantErr    bool
	}{
		{
			name:      "simple upsert",
			input:     "UPSERT Account (external_id, Name, Industry) VALUES ('ext-001', 'Acme', 'Tech') ON external_id",
			wantObj:   "Account",
			wantCols:  []string{"external_id", "Name", "Industry"},
			wantRows:  1,
			wantExtId: "external_id",
		},
		{
			name:      "multi-row upsert",
			input:     "UPSERT Account (external_id, Name) VALUES ('ext-001', 'Acme'), ('ext-002', 'Globex') ON external_id",
			wantObj:   "Account",
			wantCols:  []string{"external_id", "Name"},
			wantRows:  2,
			wantExtId: "external_id",
		},
		{
			name:      "upsert case insensitive",
			input:     "upsert Account (ExternalId, Name) values ('ext-001', 'Test') on ExternalId",
			wantObj:   "Account",
			wantCols:  []string{"ExternalId", "Name"},
			wantRows:  1,
			wantExtId: "ExternalId",
		},
		{
			name:    "invalid - missing on clause",
			input:   "UPSERT Account (external_id, Name) VALUES ('ext-001', 'Acme')",
			wantErr: true,
		},
		{
			name:    "invalid - missing values",
			input:   "UPSERT Account (external_id, Name) ON external_id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ast)
			require.NotNil(t, ast.Upsert)

			assert.Equal(t, tt.wantObj, ast.Upsert.Object)
			assert.Equal(t, tt.wantCols, ast.Upsert.Fields)
			assert.Len(t, ast.Upsert.Values, tt.wantRows)
			assert.Equal(t, tt.wantExtId, ast.Upsert.ExternalIdField)
		})
	}
}

func TestParseWhereExpressions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "equality",
			input: "DELETE FROM Task WHERE Status = 'Active'",
		},
		{
			name:  "not equal",
			input: "DELETE FROM Task WHERE Status != 'Active'",
		},
		{
			name:  "not equal alternate",
			input: "DELETE FROM Task WHERE Status <> 'Active'",
		},
		{
			name:  "greater than",
			input: "DELETE FROM Task WHERE Amount > 100",
		},
		{
			name:  "less than",
			input: "DELETE FROM Task WHERE Amount < 100",
		},
		{
			name:  "greater or equal",
			input: "DELETE FROM Task WHERE Amount >= 100",
		},
		{
			name:  "less or equal",
			input: "DELETE FROM Task WHERE Amount <= 100",
		},
		{
			name:  "is null",
			input: "DELETE FROM Task WHERE Description IS NULL",
		},
		{
			name:  "is not null",
			input: "DELETE FROM Task WHERE Description IS NOT NULL",
		},
		{
			name:  "in clause",
			input: "DELETE FROM Task WHERE Status IN ('Active', 'Pending')",
		},
		{
			name:  "not in clause",
			input: "DELETE FROM Task WHERE Status NOT IN ('Completed', 'Cancelled')",
		},
		{
			name:  "like pattern",
			input: "DELETE FROM Task WHERE Name LIKE 'Test%'",
		},
		{
			name:  "not like pattern",
			input: "DELETE FROM Task WHERE Name NOT LIKE '%test%'",
		},
		{
			name:  "and condition",
			input: "DELETE FROM Task WHERE Status = 'Active' AND Priority = 'High'",
		},
		{
			name:  "or condition",
			input: "DELETE FROM Task WHERE Status = 'Completed' OR Status = 'Cancelled'",
		},
		{
			name:  "not condition",
			input: "DELETE FROM Task WHERE NOT Status = 'Active'",
		},
		{
			name:  "parentheses",
			input: "DELETE FROM Task WHERE (Status = 'Active' OR Status = 'Pending') AND Priority = 'High'",
		},
		{
			name:  "complex nested",
			input: "DELETE FROM Task WHERE Status = 'Active' AND (Priority = 'High' OR Priority = 'Critical')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ast)
			require.NotNil(t, ast.Delete)
			require.NotNil(t, ast.Delete.Where)
		})
	}
}

func TestConstTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, val *Const)
	}{
		{
			name:  "string value",
			input: "INSERT INTO T (F) VALUES ('hello')",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.String)
				assert.Equal(t, "hello", *val.String)
				assert.Equal(t, FieldTypeString, val.GetFieldType())
			},
		},
		{
			name:  "integer value",
			input: "INSERT INTO T (F) VALUES (42)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.Integer)
				assert.Equal(t, 42, *val.Integer)
				assert.Equal(t, FieldTypeInteger, val.GetFieldType())
			},
		},
		{
			name:  "float value",
			input: "INSERT INTO T (F) VALUES (3.14)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.Float)
				assert.Equal(t, 3.14, *val.Float)
				assert.Equal(t, FieldTypeFloat, val.GetFieldType())
			},
		},
		{
			name:  "boolean true",
			input: "INSERT INTO T (F) VALUES (TRUE)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.Boolean)
				assert.True(t, bool(*val.Boolean))
				assert.Equal(t, FieldTypeBoolean, val.GetFieldType())
			},
		},
		{
			name:  "boolean false",
			input: "INSERT INTO T (F) VALUES (FALSE)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.Boolean)
				assert.False(t, bool(*val.Boolean))
				assert.Equal(t, FieldTypeBoolean, val.GetFieldType())
			},
		},
		{
			name:  "null value",
			input: "INSERT INTO T (F) VALUES (NULL)",
			checkFn: func(t *testing.T, val *Const) {
				assert.True(t, val.Null)
				assert.Equal(t, FieldTypeNull, val.GetFieldType())
			},
		},
		{
			name:  "date value",
			input: "INSERT INTO T (F) VALUES (2024-01-15)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.Date)
				assert.Equal(t, 2024, val.Date.Year())
				assert.Equal(t, 1, int(val.Date.Month()))
				assert.Equal(t, 15, val.Date.Day())
				assert.Equal(t, FieldTypeDate, val.GetFieldType())
			},
		},
		{
			name:  "datetime value",
			input: "INSERT INTO T (F) VALUES (2024-01-15T10:30:00Z)",
			checkFn: func(t *testing.T, val *Const) {
				assert.NotNil(t, val.DateTime)
				assert.Equal(t, 2024, val.DateTime.Year())
				assert.Equal(t, 10, val.DateTime.Hour())
				assert.Equal(t, 30, val.DateTime.Minute())
				assert.Equal(t, FieldTypeDateTime, val.GetFieldType())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, ast.Insert)
			require.Len(t, ast.Insert.Values, 1)
			require.Len(t, ast.Insert.Values[0].Values, 1)

			expr := ast.Insert.Values[0].Values[0]
			require.True(t, expr.IsConst(), "expected constant expression")
			tt.checkFn(t, expr.Const)
		})
	}
}

func TestParseStringEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "INSERT INTO T (F) VALUES ('hello')",
			expected: "hello",
		},
		{
			name:     "escaped quote",
			input:    "INSERT INTO T (F) VALUES ('it''s')",
			expected: "it's",
		},
		{
			name:     "multiple escaped quotes",
			input:    "INSERT INTO T (F) VALUES ('he said ''hi''')",
			expected: "he said 'hi'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, ast.Insert)

			expr := ast.Insert.Values[0].Values[0]
			require.True(t, expr.IsConst())
			require.NotNil(t, expr.Const.String)
			assert.Equal(t, tt.expected, *expr.Const.String)
		})
	}
}

func TestDMLStatementMethods(t *testing.T) {
	t.Run("GetOperation returns correct type", func(t *testing.T) {
		insertAST, _ := Parse("INSERT INTO T (F) VALUES ('v')")
		assert.Equal(t, OperationInsert, insertAST.GetOperation())

		updateAST, _ := Parse("UPDATE T SET F = 'v'")
		assert.Equal(t, OperationUpdate, updateAST.GetOperation())

		deleteAST, _ := Parse("DELETE FROM T")
		assert.Equal(t, OperationDelete, deleteAST.GetOperation())

		upsertAST, _ := Parse("UPSERT T (F) VALUES ('v') ON F")
		assert.Equal(t, OperationUpsert, upsertAST.GetOperation())
	})

	t.Run("GetObject returns correct name", func(t *testing.T) {
		insertAST, _ := Parse("INSERT INTO Account (F) VALUES ('v')")
		assert.Equal(t, "Account", insertAST.GetObject())

		updateAST, _ := Parse("UPDATE Contact SET F = 'v'")
		assert.Equal(t, "Contact", updateAST.GetObject())

		deleteAST, _ := Parse("DELETE FROM Task")
		assert.Equal(t, "Task", deleteAST.GetObject())

		upsertAST, _ := Parse("UPSERT Lead (F) VALUES ('v') ON F")
		assert.Equal(t, "Lead", upsertAST.GetObject())
	})
}

func TestParseFunctionCalls(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		funcName   Function
		argCount   int
		wantErr    bool
	}{
		{
			name:     "UPPER function",
			input:    "INSERT INTO T (F) VALUES (UPPER('test'))",
			funcName: FuncUpper,
			argCount: 1,
		},
		{
			name:     "LOWER function",
			input:    "INSERT INTO T (F) VALUES (LOWER('TEST'))",
			funcName: FuncLower,
			argCount: 1,
		},
		{
			name:     "TRIM function",
			input:    "INSERT INTO T (F) VALUES (TRIM('  test  '))",
			funcName: FuncTrim,
			argCount: 1,
		},
		{
			name:     "COALESCE function",
			input:    "INSERT INTO T (F) VALUES (COALESCE(NULL, 'default'))",
			funcName: FuncCoalesce,
			argCount: 2,
		},
		{
			name:     "COALESCE with three args",
			input:    "INSERT INTO T (F) VALUES (COALESCE(NULL, NULL, 'default'))",
			funcName: FuncCoalesce,
			argCount: 3,
		},
		{
			name:     "CONCAT function",
			input:    "INSERT INTO T (F) VALUES (CONCAT('Hello', ' ', 'World'))",
			funcName: FuncConcat,
			argCount: 3,
		},
		{
			name:     "NULLIF function",
			input:    "INSERT INTO T (F) VALUES (NULLIF('test', 'test'))",
			funcName: FuncNullif,
			argCount: 2,
		},
		{
			name:     "LENGTH function",
			input:    "INSERT INTO T (F) VALUES (LENGTH('test'))",
			funcName: FuncLength,
			argCount: 1,
		},
		{
			name:     "LEN alias",
			input:    "INSERT INTO T (F) VALUES (LEN('test'))",
			funcName: FuncLength,
			argCount: 1,
		},
		{
			name:     "SUBSTRING with two args",
			input:    "INSERT INTO T (F) VALUES (SUBSTRING('hello', 1))",
			funcName: FuncSubstring,
			argCount: 2,
		},
		{
			name:     "SUBSTRING with three args",
			input:    "INSERT INTO T (F) VALUES (SUBSTRING('hello', 1, 3))",
			funcName: FuncSubstring,
			argCount: 3,
		},
		{
			name:     "SUBSTR alias",
			input:    "INSERT INTO T (F) VALUES (SUBSTR('hello', 2))",
			funcName: FuncSubstring,
			argCount: 2,
		},
		{
			name:     "ABS function",
			input:    "INSERT INTO T (F) VALUES (ABS(5))",
			funcName: FuncAbs,
			argCount: 1,
		},
		{
			name:     "ROUND with one arg",
			input:    "INSERT INTO T (F) VALUES (ROUND(3.14159))",
			funcName: FuncRound,
			argCount: 1,
		},
		{
			name:     "ROUND with two args",
			input:    "INSERT INTO T (F) VALUES (ROUND(3.14159, 2))",
			funcName: FuncRound,
			argCount: 2,
		},
		{
			name:     "FLOOR function",
			input:    "INSERT INTO T (F) VALUES (FLOOR(3.7))",
			funcName: FuncFloor,
			argCount: 1,
		},
		{
			name:     "CEIL function",
			input:    "INSERT INTO T (F) VALUES (CEIL(3.2))",
			funcName: FuncCeil,
			argCount: 1,
		},
		{
			name:     "CEILING alias",
			input:    "INSERT INTO T (F) VALUES (CEILING(3.2))",
			funcName: FuncCeil,
			argCount: 1,
		},
		{
			name:     "nested functions",
			input:    "INSERT INTO T (F) VALUES (UPPER(TRIM('  test  ')))",
			funcName: FuncUpper,
			argCount: 1,
		},
		{
			name:     "function in UPDATE SET",
			input:    "UPDATE T SET F = UPPER('value')",
			funcName: FuncUpper,
			argCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var expr *Expr
			if ast.Insert != nil {
				require.Len(t, ast.Insert.Values, 1)
				require.Len(t, ast.Insert.Values[0].Values, 1)
				expr = ast.Insert.Values[0].Values[0]
			} else if ast.Update != nil {
				require.Len(t, ast.Update.Assignments, 1)
				expr = ast.Update.Assignments[0].Value
			}

			require.NotNil(t, expr)
			require.True(t, expr.IsFunc(), "expected function expression")
			assert.Equal(t, tt.funcName, expr.FuncCall.Name)
			assert.Len(t, expr.FuncCall.Args, tt.argCount)
		})
	}
}

func TestParseNestedFunctions(t *testing.T) {
	t.Run("nested function call", func(t *testing.T) {
		ast, err := Parse("INSERT INTO T (F) VALUES (UPPER(TRIM(LOWER('TEST'))))")
		require.NoError(t, err)

		expr := ast.Insert.Values[0].Values[0]
		require.True(t, expr.IsFunc())

		// Outer: UPPER
		assert.Equal(t, FuncUpper, expr.FuncCall.Name)
		require.Len(t, expr.FuncCall.Args, 1)

		// Middle: TRIM
		trimArg := expr.FuncCall.Args[0]
		require.True(t, trimArg.IsFunc())
		assert.Equal(t, FuncTrim, trimArg.FuncCall.Name)
		require.Len(t, trimArg.FuncCall.Args, 1)

		// Inner: LOWER
		lowerArg := trimArg.FuncCall.Args[0]
		require.True(t, lowerArg.IsFunc())
		assert.Equal(t, FuncLower, lowerArg.FuncCall.Name)
		require.Len(t, lowerArg.FuncCall.Args, 1)

		// Innermost: constant 'TEST'
		constArg := lowerArg.FuncCall.Args[0]
		require.True(t, constArg.IsConst())
		assert.Equal(t, "TEST", *constArg.Const.String)
	})

	t.Run("function with multiple nested args", func(t *testing.T) {
		ast, err := Parse("INSERT INTO T (F) VALUES (COALESCE(TRIM(NULL), UPPER('default')))")
		require.NoError(t, err)

		expr := ast.Insert.Values[0].Values[0]
		require.True(t, expr.IsFunc())
		assert.Equal(t, FuncCoalesce, expr.FuncCall.Name)
		require.Len(t, expr.FuncCall.Args, 2)

		// First arg: TRIM(NULL)
		arg0 := expr.FuncCall.Args[0]
		require.True(t, arg0.IsFunc())
		assert.Equal(t, FuncTrim, arg0.FuncCall.Name)

		// Second arg: UPPER('default')
		arg1 := expr.FuncCall.Args[1]
		require.True(t, arg1.IsFunc())
		assert.Equal(t, FuncUpper, arg1.FuncCall.Name)
	})
}
