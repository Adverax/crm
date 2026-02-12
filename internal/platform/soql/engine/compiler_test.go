package engine

import (
	"context"
	"strings"
	"testing"
)

func TestCompileSimpleSelect(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		query      string
		wantSQL    string
		wantParams int
	}{
		{
			name:       "single field",
			query:      "SELECT Name FROM Account",
			wantSQL:    `SELECT t0."name" AS`,
			wantParams: 0,
		},
		{
			name:       "multiple fields",
			query:      "SELECT Name, Industry FROM Account",
			wantSQL:    `SELECT t0."name" AS`,
			wantParams: 0,
		},
		{
			name:       "with where string",
			query:      "SELECT Name FROM Account WHERE Industry = 'Technology'",
			wantSQL:    `WHERE t0."industry" = 'Technology'`,
			wantParams: 0,
		},
		{
			name:       "with where number",
			query:      "SELECT Name FROM Account WHERE AnnualRevenue > 1000000",
			wantSQL:    `WHERE t0."annual_revenue" > 1000000`,
			wantParams: 0,
		},
		{
			name:       "with order by",
			query:      "SELECT Name FROM Account ORDER BY Name ASC",
			wantSQL:    `ORDER BY t0."name"`,
			wantParams: 0,
		},
		{
			name:       "with limit",
			query:      "SELECT Name FROM Account LIMIT 10",
			wantSQL:    "LIMIT 10",
			wantParams: 0,
		},
		{
			name:       "with limit and offset (offset ignored for keyset)",
			query:      "SELECT Name FROM Account LIMIT 10 OFFSET 20",
			wantSQL:    "LIMIT 10",
			wantParams: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}

			if len(compiled.Params) != tt.wantParams {
				t.Errorf("Params count = %d, want %d", len(compiled.Params), tt.wantParams)
			}
		})
	}
}

func TestCompileLookups(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		query     string
		wantSQL   string
		wantJoins int
	}{
		{
			name:      "simple lookup",
			query:     "SELECT Name, Account.Name FROM Contact",
			wantSQL:   `LEFT JOIN "accounts" AS t1 ON t0."account_id" = t1."id"`,
			wantJoins: 1,
		},
		{
			name:      "nested lookup",
			query:     "SELECT Name, Account.Owner.Name FROM Contact",
			wantSQL:   `LEFT JOIN "users" AS t2 ON t1."owner_id" = t2."id"`,
			wantJoins: 2,
		},
		{
			name:      "lookup in where",
			query:     "SELECT Name FROM Contact WHERE Account.Industry = 'Tech'",
			wantSQL:   `WHERE t1."industry"`,
			wantJoins: 1,
		},
		{
			name:      "lookup in order by",
			query:     "SELECT Name FROM Contact ORDER BY Account.Name",
			wantSQL:   `ORDER BY t1."name"`,
			wantJoins: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}

			joinCount := strings.Count(compiled.SQL, "LEFT JOIN")
			if joinCount < tt.wantJoins {
				t.Errorf("JOIN count = %d, want at least %d\nSQL: %s", joinCount, tt.wantJoins, compiled.SQL)
			}
		})
	}
}

func TestCompileAggregates(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "COUNT",
			query:   "SELECT COUNT(Id) FROM Account",
			wantSQL: `COUNT(t0."id")`,
		},
		{
			name:    "SUM",
			query:   "SELECT SUM(Amount) FROM Opportunity",
			wantSQL: `SUM(t0."amount")`,
		},
		{
			name:    "AVG",
			query:   "SELECT AVG(Amount) FROM Opportunity",
			wantSQL: `AVG(t0."amount")`,
		},
		{
			name:    "MIN",
			query:   "SELECT MIN(Amount) FROM Opportunity",
			wantSQL: `MIN(t0."amount")`,
		},
		{
			name:    "MAX",
			query:   "SELECT MAX(Amount) FROM Opportunity",
			wantSQL: `MAX(t0."amount")`,
		},
		{
			name:    "multiple aggregates",
			query:   "SELECT COUNT(Id), SUM(Amount), AVG(Amount) FROM Opportunity",
			wantSQL: `COUNT(t0."id")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileGroupBy(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "simple group by",
			query:   "SELECT StageName, COUNT(Id) FROM Opportunity GROUP BY StageName",
			wantSQL: `GROUP BY t0."stage_name"`,
		},
		{
			name:    "group by with having",
			query:   "SELECT StageName, COUNT(Id) FROM Opportunity GROUP BY StageName HAVING COUNT(Id) > 5",
			wantSQL: `HAVING COUNT(t0."id") > 5`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileSubqueries(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "simple subquery",
			query:   "SELECT Name, (SELECT FirstName, LastName FROM Contacts) FROM Account",
			wantSQL: "JSON_AGG",
		},
		{
			name:    "subquery with limit",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts LIMIT 5) FROM Account",
			wantSQL: "LIMIT 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}

			// Check that relationships are in shape
			if len(compiled.Shape.Relationships) == 0 {
				t.Error("expected relationships in shape")
			}
		})
	}
}

func TestCompileDateLiterals(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name           string
		query          string
		wantDateParams int
	}{
		{
			name:           "TODAY",
			query:          "SELECT Name FROM Account WHERE CreatedDate = TODAY",
			wantDateParams: 1,
		},
		{
			name:           "LAST_N_DAYS",
			query:          "SELECT Name FROM Account WHERE CreatedDate > LAST_N_DAYS:30",
			wantDateParams: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if len(compiled.DateParams) != tt.wantDateParams {
				t.Errorf("DateParams count = %d, want %d", len(compiled.DateParams), tt.wantDateParams)
			}

			// Check that SQL has placeholder
			if tt.wantDateParams > 0 && !strings.Contains(compiled.SQL, "$1") {
				t.Errorf("SQL should contain $1 placeholder\nGot: %s", compiled.SQL)
			}
		})
	}
}

func TestCompileOperators(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "equals",
			query:   "SELECT Name FROM Account WHERE Industry = 'Tech'",
			wantSQL: "= 'Tech'",
		},
		{
			name:    "not equals",
			query:   "SELECT Name FROM Account WHERE Industry != 'Tech'",
			wantSQL: "!= 'Tech'",
		},
		{
			name:    "greater than",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue > 1000000",
			wantSQL: "> 1000000",
		},
		{
			name:    "less than",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue < 1000000",
			wantSQL: "< 1000000",
		},
		{
			name:    "greater or equal",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue >= 1000000",
			wantSQL: ">= 1000000",
		},
		{
			name:    "less or equal",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue <= 1000000",
			wantSQL: "<= 1000000",
		},
		{
			name:    "IS NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NULL",
			wantSQL: "IS NULL",
		},
		{
			name:    "IS NOT NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NOT NULL",
			wantSQL: "IS NOT NULL",
		},
		{
			name:    "IN clause",
			query:   "SELECT Name FROM Account WHERE Industry IN ('Tech', 'Finance')",
			wantSQL: "IN ('Tech', 'Finance')",
		},
		{
			name:    "NOT IN clause",
			query:   "SELECT Name FROM Account WHERE Industry NOT IN ('Tech', 'Finance')",
			wantSQL: "NOT IN ('Tech', 'Finance')",
		},
		{
			name:    "LIKE",
			query:   "SELECT Name FROM Account WHERE Name LIKE 'Acme%'",
			wantSQL: "LIKE 'Acme%'",
		},
		{
			name:    "AND",
			query:   "SELECT Name FROM Account WHERE Industry = 'Tech' AND AnnualRevenue > 1000000",
			wantSQL: "AND",
		},
		{
			name:    "OR",
			query:   "SELECT Name FROM Account WHERE Industry = 'Tech' OR Industry = 'Finance'",
			wantSQL: "OR",
		},
		{
			name:    "NOT",
			query:   "SELECT Name FROM Account WHERE NOT Industry = 'Tech'",
			wantSQL: "NOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileArithmeticExpressions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "multiplication in SELECT",
			query:   "SELECT Amount * 0.1 FROM Opportunity",
			wantSQL: `t0."amount" * 0.1`,
		},
		{
			name:    "addition in SELECT",
			query:   "SELECT Amount + 100 FROM Opportunity",
			wantSQL: `t0."amount" + 100`,
		},
		{
			name:    "subtraction in SELECT",
			query:   "SELECT Amount - 50 FROM Opportunity",
			wantSQL: `t0."amount" - 50`,
		},
		{
			name:    "division in SELECT",
			query:   "SELECT Amount / 100 FROM Opportunity",
			wantSQL: `t0."amount" / 100`,
		},
		{
			name:    "arithmetic in WHERE",
			query:   "SELECT Name FROM Opportunity WHERE Amount * 2 > 1000",
			wantSQL: `t0."amount" * 2 > 1000`,
		},
		{
			name:    "complex arithmetic",
			query:   "SELECT Amount * 0.1 + 50 FROM Opportunity",
			wantSQL: `t0."amount" * 0.1 + 50`,
		},
		{
			name:    "parenthesized expression",
			query:   "SELECT (Amount + 100) * 0.9 FROM Opportunity",
			wantSQL: `(t0."amount" + 100) * 0.9`,
		},
		{
			name:    "unary minus",
			query:   "SELECT -Amount FROM Opportunity",
			wantSQL: `-t0."amount"`,
		},
		{
			name:    "field to field arithmetic",
			query:   "SELECT Name FROM Account WHERE AnnualRevenue - AnnualRevenue > 0",
			wantSQL: `t0."annual_revenue" - t0."annual_revenue" > 0`,
		},
		{
			name:    "string concatenation",
			query:   "SELECT FirstName || ' ' || LastName FROM Contact",
			wantSQL: `t0."first_name" || ' ' || t0."last_name"`,
		},
		{
			name:    "string concatenation in WHERE",
			query:   "SELECT Name FROM Contact WHERE FirstName || LastName = 'JohnDoe'",
			wantSQL: `t0."first_name" || t0."last_name" = 'JohnDoe'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileFunctions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		// String functions
		{
			name:    "COALESCE",
			query:   "SELECT COALESCE(Name, 'Unknown') FROM Account",
			wantSQL: `COALESCE(t0."name", 'Unknown')`,
		},
		{
			name:    "COALESCE three args",
			query:   "SELECT COALESCE(FirstName, LastName, 'N/A') FROM Contact",
			wantSQL: `COALESCE(t0."first_name", t0."last_name", 'N/A')`,
		},
		{
			name:    "NULLIF",
			query:   "SELECT NULLIF(Name, 'Inactive') FROM Account",
			wantSQL: `NULLIF(t0."name", 'Inactive')`,
		},
		{
			name:    "CONCAT",
			query:   "SELECT CONCAT(FirstName, LastName) FROM Contact",
			wantSQL: `CONCAT(t0."first_name", t0."last_name")`,
		},
		{
			name:    "CONCAT multiple args",
			query:   "SELECT CONCAT(FirstName, ' ', LastName) FROM Contact",
			wantSQL: `CONCAT(t0."first_name", ' ', t0."last_name")`,
		},
		{
			name:    "UPPER",
			query:   "SELECT UPPER(Name) FROM Account",
			wantSQL: `UPPER(t0."name")`,
		},
		{
			name:    "LOWER",
			query:   "SELECT LOWER(Email) FROM Contact",
			wantSQL: `LOWER(t0."email")`,
		},
		{
			name:    "TRIM",
			query:   "SELECT TRIM(Name) FROM Account",
			wantSQL: `TRIM(t0."name")`,
		},
		{
			name:    "LENGTH",
			query:   "SELECT LENGTH(Name) FROM Account",
			wantSQL: `LENGTH(t0."name")`,
		},
		{
			name:    "SUBSTRING two args",
			query:   "SELECT SUBSTRING(Name, 1) FROM Account",
			wantSQL: `SUBSTRING(t0."name", 1)`,
		},
		{
			name:    "SUBSTRING three args",
			query:   "SELECT SUBSTRING(Name, 1, 10) FROM Account",
			wantSQL: `SUBSTRING(t0."name", 1, 10)`,
		},
		// Math functions
		{
			name:    "ABS",
			query:   "SELECT ABS(AnnualRevenue) FROM Account",
			wantSQL: `ABS(t0."annual_revenue")`,
		},
		{
			name:    "ROUND single arg",
			query:   "SELECT ROUND(Amount) FROM Opportunity",
			wantSQL: `ROUND(t0."amount")`,
		},
		{
			name:    "ROUND with decimals",
			query:   "SELECT ROUND(Amount, 2) FROM Opportunity",
			wantSQL: `ROUND(t0."amount", 2)`,
		},
		{
			name:    "FLOOR",
			query:   "SELECT FLOOR(Amount) FROM Opportunity",
			wantSQL: `FLOOR(t0."amount")`,
		},
		{
			name:    "CEIL",
			query:   "SELECT CEIL(Amount) FROM Opportunity",
			wantSQL: `CEIL(t0."amount")`,
		},
		// Functions in WHERE
		{
			name:    "function in WHERE",
			query:   "SELECT Name FROM Account WHERE UPPER(Name) = 'TEST'",
			wantSQL: `UPPER(t0."name") = 'TEST'`,
		},
		{
			name:    "LENGTH in WHERE",
			query:   "SELECT Name FROM Account WHERE LENGTH(Name) > 10",
			wantSQL: `LENGTH(t0."name") > 10`,
		},
		// Nested functions
		{
			name:    "nested functions",
			query:   "SELECT UPPER(TRIM(Name)) FROM Account",
			wantSQL: `UPPER(TRIM(t0."name"))`,
		},
		// Functions with arithmetic
		{
			name:    "function with arithmetic",
			query:   "SELECT ROUND(Amount * 1.1, 2) FROM Opportunity",
			wantSQL: `ROUND(t0."amount" * 1.1, 2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileOrderByDirection(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "ASC",
			query:   "SELECT Name FROM Account ORDER BY Name ASC",
			wantSQL: `ORDER BY t0."name"`,
		},
		{
			name:    "DESC",
			query:   "SELECT Name FROM Account ORDER BY Name DESC",
			wantSQL: `ORDER BY t0."name" DESC`,
		},
		{
			name:    "NULLS FIRST",
			query:   "SELECT Name FROM Account ORDER BY Name NULLS FIRST",
			wantSQL: "NULLS FIRST",
		},
		{
			name:    "NULLS LAST",
			query:   "SELECT Name FROM Account ORDER BY Name NULLS LAST",
			wantSQL: "NULLS LAST",
		},
		{
			name:    "DESC NULLS FIRST",
			query:   "SELECT Name FROM Account ORDER BY Name DESC NULLS FIRST",
			wantSQL: "DESC NULLS FIRST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileResultShape(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := "SELECT Name, Industry FROM Account"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Check shape
	if compiled.Shape.Object != "Account" {
		t.Errorf("Shape.Object = %s, want Account", compiled.Shape.Object)
	}

	if compiled.Shape.Table != `"accounts"` {
		t.Errorf("Shape.Table = %s, want \"accounts\"", compiled.Shape.Table)
	}

	// 2 user fields + 1 keyset field (id for pagination)
	if len(compiled.Shape.Fields) < 2 {
		t.Errorf("Shape.Fields count = %d, want at least 2", len(compiled.Shape.Fields))
	}

	// First 2 fields should be user-selected
	if compiled.Shape.Fields[0].Name != "Name" {
		t.Errorf("Shape.Fields[0].Name = %s, want Name", compiled.Shape.Fields[0].Name)
	}
	if compiled.Shape.Fields[1].Name != "Industry" {
		t.Errorf("Shape.Fields[1].Name = %s, want Industry", compiled.Shape.Fields[1].Name)
	}
}

func TestCompileComplexQuery(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := `SELECT
		Name,
		Account.Name,
		Account.Owner.Name
	FROM Contact
	WHERE Account.Industry = 'Technology'
		AND Email IS NOT NULL
	ORDER BY LastName ASC, FirstName DESC
	LIMIT 100 OFFSET 10`

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Verify key parts of the SQL
	// Note: OFFSET is ignored with keyset pagination
	checks := []string{
		"SELECT",
		`FROM "contacts" AS t0`,
		`LEFT JOIN "accounts" AS t1`,
		`LEFT JOIN "users" AS t2`,
		"WHERE",
		"IS NOT NULL",
		"ORDER BY",
		"LIMIT 100",
	}

	for _, check := range checks {
		if !strings.Contains(compiled.SQL, check) {
			t.Errorf("SQL missing expected part: %s\nGot: %s", check, compiled.SQL)
		}
	}
}

func TestCompileBooleanLiterals(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	// Add a boolean field to Account for this test
	accountMeta, _ := metadata.GetObject(ctx, "Account")
	if accountMeta != nil {
		accountMeta.Fields["IsActive"] = &FieldMeta{
			Name:       "IsActive",
			Column:     "is_active",
			Type:       FieldTypeBoolean,
			Filterable: true,
		}
	}

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "TRUE literal",
			query:   "SELECT Name FROM Account WHERE IsActive = TRUE",
			wantSQL: "= true",
		},
		{
			name:    "FALSE literal",
			query:   "SELECT Name FROM Account WHERE IsActive = FALSE",
			wantSQL: "= false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(strings.ToLower(compiled.SQL), tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileNullHandling(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "IS NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NULL",
			wantSQL: "IS NULL",
		},
		{
			name:    "IS NOT NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NOT NULL",
			wantSQL: "IS NOT NULL",
		},
		{
			name:    "multiple null checks",
			query:   "SELECT Name FROM Contact WHERE Email IS NOT NULL AND Phone IS NULL",
			wantSQL: "IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileComplexWhere(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		wantPart string
	}{
		{
			name:     "nested AND/OR",
			query:    "SELECT Name FROM Account WHERE (Industry = 'Tech' OR Industry = 'Finance') AND AnnualRevenue > 1000000",
			wantPart: "AND",
		},
		{
			name:     "multiple OR",
			query:    "SELECT Name FROM Account WHERE Industry = 'Tech' OR Industry = 'Finance' OR Industry = 'Healthcare'",
			wantPart: "OR",
		},
		{
			name:     "NOT condition",
			query:    "SELECT Name FROM Account WHERE NOT Industry = 'Tech'",
			wantPart: "NOT",
		},
		{
			name:     "complex parentheses",
			query:    "SELECT Name FROM Account WHERE ((Industry = 'Tech') AND (AnnualRevenue > 1000000))",
			wantPart: "AND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantPart) {
				t.Errorf("SQL does not contain expected part\nGot: %s\nWant: %s", compiled.SQL, tt.wantPart)
			}
		})
	}
}

func TestCompileAliases(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		query     string
		wantAlias string
	}{
		{
			name:      "field alias",
			query:     "SELECT Name AS n FROM Account",
			wantAlias: "AS n",
		},
		{
			name:      "aggregate alias",
			query:     "SELECT COUNT(Id) AS total FROM Account",
			wantAlias: "AS total",
		},
		{
			name:      "expression alias",
			query:     "SELECT AnnualRevenue * 0.1 AS tax FROM Account",
			wantAlias: "AS tax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantAlias) {
				t.Errorf("SQL does not contain expected alias\nGot: %s\nWant: %s", compiled.SQL, tt.wantAlias)
			}
		})
	}
}

func TestCompileNestedLookups(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := "SELECT Name, Account.Name, Account.Owner.Name, Account.Owner.Manager.Name FROM Contact"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Should have 3 LEFT JOINs: Account, Owner, Manager
	joinCount := strings.Count(compiled.SQL, "LEFT JOIN")
	if joinCount < 3 {
		t.Errorf("expected at least 3 LEFT JOINs, got %d\nSQL: %s", joinCount, compiled.SQL)
	}

	// Verify table aliases are used
	if !strings.Contains(compiled.SQL, "AS t1") {
		t.Error("expected table alias t1")
	}
	if !strings.Contains(compiled.SQL, "AS t2") {
		t.Error("expected table alias t2")
	}
	if !strings.Contains(compiled.SQL, "AS t3") {
		t.Error("expected table alias t3")
	}
}

func TestCompileShapeFields(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := "SELECT Name, Industry, AnnualRevenue FROM Account"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Check shape includes all requested fields
	fieldNames := make(map[string]bool)
	for _, f := range compiled.Shape.Fields {
		fieldNames[f.Name] = true
	}

	expectedFields := []string{"Name", "Industry", "AnnualRevenue"}
	for _, name := range expectedFields {
		if !fieldNames[name] {
			t.Errorf("shape missing field: %s", name)
		}
	}
}

func TestCompileSubqueryShape(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := "SELECT Name, (SELECT FirstName, LastName FROM Contacts) FROM Account"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Check shape includes relationship
	if len(compiled.Shape.Relationships) == 0 {
		t.Error("shape should include relationships")
	}

	// Check relationship shape - find Contacts in the slice
	var foundContacts bool
	for _, rel := range compiled.Shape.Relationships {
		if rel.Name == "Contacts" {
			foundContacts = true
			if rel.Shape != nil && rel.Shape.Object != "Contact" {
				t.Errorf("relationship Object = %s, want Contact", rel.Shape.Object)
			}
			break
		}
	}
	if !foundContacts {
		t.Error("shape should include Contacts relationship")
	}
}

func TestCompileWhereSubquery(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "IN subquery",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact)",
			wantSQL: "IN (SELECT",
		},
		{
			name:    "NOT IN subquery",
			query:   "SELECT Name FROM Account WHERE Id NOT IN (SELECT AccountId FROM Contact)",
			wantSQL: "NOT IN (SELECT",
		},
		{
			name:    "subquery with WHERE",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact WHERE Email IS NOT NULL)",
			wantSQL: "IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompileAllDateLiterals(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	staticLiterals := []string{
		"TODAY", "YESTERDAY", "TOMORROW",
		"THIS_WEEK", "LAST_WEEK", "NEXT_WEEK",
		"THIS_MONTH", "LAST_MONTH", "NEXT_MONTH",
	}

	for _, lit := range staticLiterals {
		t.Run(lit, func(t *testing.T) {
			query := "SELECT Name FROM Account WHERE CreatedDate = " + lit

			ast, err := Parse(query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			// Should have a parameter placeholder
			if !strings.Contains(compiled.SQL, "$") {
				t.Errorf("SQL should contain parameter placeholder for date literal\nGot: %s", compiled.SQL)
			}

			// Should have DateParams
			if len(compiled.DateParams) == 0 {
				t.Error("expected DateParams for date literal")
			}
		})
	}
}

func TestCompileDynamicDateLiterals(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name  string
		query string
	}{
		{"LAST_N_DAYS", "SELECT Name FROM Account WHERE CreatedDate > LAST_N_DAYS:30"},
		{"NEXT_N_DAYS", "SELECT Name FROM Account WHERE CreatedDate < NEXT_N_DAYS:7"},
		{"LAST_N_MONTHS", "SELECT Name FROM Account WHERE CreatedDate >= LAST_N_MONTHS:3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			// Should have DateParams with N value
			if len(compiled.DateParams) == 0 {
				t.Error("expected DateParams for dynamic date literal")
			}
		})
	}
}

func TestCompileModuloOperator(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	query := "SELECT Name FROM Account WHERE AnnualRevenue % 1000 = 0"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// PostgreSQL uses % for modulo
	if !strings.Contains(compiled.SQL, "%") {
		t.Errorf("SQL should contain modulo operator\nGot: %s", compiled.SQL)
	}
}

func TestCompileNestedFunctions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantSQL string
	}{
		{
			name:    "UPPER(TRIM())",
			query:   "SELECT UPPER(TRIM(Name)) FROM Account",
			wantSQL: `UPPER(TRIM(t0."name"))`,
		},
		{
			name:    "COALESCE(TRIM(), literal)",
			query:   "SELECT COALESCE(TRIM(Name), 'N/A') FROM Account",
			wantSQL: `COALESCE(TRIM(t0."name"), 'N/A')`,
		},
		{
			name:    "ROUND(arithmetic)",
			query:   "SELECT ROUND(AnnualRevenue * 1.1, 2) FROM Account",
			wantSQL: `ROUND(t0."annual_revenue" * 1.1, 2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if !strings.Contains(compiled.SQL, tt.wantSQL) {
				t.Errorf("SQL does not contain expected string\nGot: %s\nWant to contain: %s", compiled.SQL, tt.wantSQL)
			}
		})
	}
}

func TestCompilerWithOptions(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	// Test with different compiler options if available
	compiler := NewCompiler(nil)

	query := "SELECT Name FROM Account LIMIT 10"

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	validated, err := validator.Validate(ctx, ast)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	compiled, err := compiler.Compile(validated)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Verify SQL is valid
	if compiled.SQL == "" {
		t.Error("compiled SQL is empty")
	}
	if compiled.Shape == nil {
		t.Error("compiled Shape is nil")
	}
}

func TestCompilePaginationInfo(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		wantSize int
	}{
		{
			name:     "with LIMIT",
			query:    "SELECT Name FROM Account LIMIT 50",
			wantSize: 50,
		},
		{
			name:     "with ORDER BY",
			query:    "SELECT Name FROM Account ORDER BY Name DESC LIMIT 25",
			wantSize: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			if compiled.Pagination == nil {
				t.Fatal("Pagination is nil")
			}

			if compiled.Pagination.PageSize != tt.wantSize {
				t.Errorf("PageSize = %d, want %d", compiled.Pagination.PageSize, tt.wantSize)
			}
		})
	}
}

func TestCompileForUpdate(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		query     string
		wantSQL   string
		forUpdate bool
	}{
		{
			name:      "simple FOR UPDATE",
			query:     "SELECT Name FROM Account FOR UPDATE",
			wantSQL:   "FOR UPDATE",
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with WHERE",
			query:     "SELECT Name FROM Account WHERE Industry = 'Tech' FOR UPDATE",
			wantSQL:   "FOR UPDATE",
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with ORDER BY and LIMIT",
			query:     "SELECT Name FROM Account ORDER BY Name LIMIT 10 FOR UPDATE",
			wantSQL:   "FOR UPDATE",
			forUpdate: true,
		},
		{
			name:      "without FOR UPDATE",
			query:     "SELECT Name FROM Account",
			wantSQL:   "",
			forUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			// Check ForUpdate flag
			if compiled.ForUpdate != tt.forUpdate {
				t.Errorf("ForUpdate = %v, want %v", compiled.ForUpdate, tt.forUpdate)
			}

			// Check SQL contains FOR UPDATE
			if tt.forUpdate {
				if !containsForUpdate(compiled.SQL) {
					t.Errorf("SQL should contain FOR UPDATE:\n%s", compiled.SQL)
				}
			} else {
				if containsForUpdate(compiled.SQL) {
					t.Errorf("SQL should not contain FOR UPDATE:\n%s", compiled.SQL)
				}
			}
		})
	}
}

func containsForUpdate(sql string) bool {
	return len(sql) >= 10 && (strings.Contains(strings.ToUpper(sql), "FOR UPDATE"))
}

func TestCompileWithSecurityEnforced(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name                 string
		query                string
		withSecurityEnforced bool
	}{
		{
			name:                 "simple WITH SECURITY_ENFORCED",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED",
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with WHERE",
			query:                "SELECT Name FROM Account WHERE Industry = 'Tech' WITH SECURITY_ENFORCED",
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with ORDER BY and LIMIT",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED ORDER BY Name LIMIT 10",
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED and FOR UPDATE",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED FOR UPDATE",
			withSecurityEnforced: true,
		},
		{
			name:                 "without WITH SECURITY_ENFORCED",
			query:                "SELECT Name FROM Account",
			withSecurityEnforced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			// Check WithSecurityEnforced flag
			if compiled.WithSecurityEnforced != tt.withSecurityEnforced {
				t.Errorf("WithSecurityEnforced = %v, want %v", compiled.WithSecurityEnforced, tt.withSecurityEnforced)
			}
		})
	}
}

func TestCompileTypeof(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name         string
		query        string
		wantErr      bool
		containsCase bool
		containsJSON bool
	}{
		{
			name: "simple TYPEOF with single WHEN",
			query: `SELECT
				TYPEOF WhatId
					WHEN Account THEN Name, Industry
				END
			FROM Task`,
			wantErr:      false,
			containsCase: true,
			containsJSON: true,
		},
		{
			name: "TYPEOF with multiple WHEN clauses",
			query: `SELECT
				TYPEOF WhatId
					WHEN Account THEN Name, Industry
					WHEN Opportunity THEN Name, StageName
				END
			FROM Task`,
			wantErr:      false,
			containsCase: true,
			containsJSON: true,
		},
		{
			name: "TYPEOF with other fields",
			query: `SELECT
				Id, Subject,
				TYPEOF WhatId
					WHEN Account THEN Name
				END
			FROM Task`,
			wantErr:      false,
			containsCase: true,
			containsJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			// Verify TYPEOF compiled to CASE expression
			if tt.containsCase && !strings.Contains(strings.ToUpper(compiled.SQL), "CASE") {
				t.Errorf("Expected SQL to contain CASE, got: %s", compiled.SQL)
			}

			// Verify TYPEOF uses JSON_BUILD_OBJECT
			if tt.containsJSON && !strings.Contains(strings.ToUpper(compiled.SQL), "JSON_BUILD_OBJECT") {
				t.Errorf("Expected SQL to contain JSON_BUILD_OBJECT, got: %s", compiled.SQL)
			}

			// Verify TypeofExpressions were tracked
			if len(validated.TypeofExpressions) == 0 {
				t.Error("Expected TypeofExpressions to be populated")
			}
		})
	}
}

func TestValidateTypeofErrors(t *testing.T) {
	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "TYPEOF with unknown polymorphic field",
			query: `SELECT
				TYPEOF UnknownField
					WHEN Account THEN Name
				END
			FROM Task`,
			wantErr: true,
			errMsg:  "UnknownField",
		},
		{
			name: "TYPEOF with unknown object type in WHEN",
			query: `SELECT
				TYPEOF WhatId
					WHEN UnknownObject THEN Name
				END
			FROM Task`,
			wantErr: true,
			errMsg:  "UnknownObject",
		},
		{
			name: "TYPEOF with unknown field in WHEN clause",
			query: `SELECT
				TYPEOF WhatId
					WHEN Account THEN Name, UnknownField
				END
			FROM Task`,
			wantErr: true,
			errMsg:  "UnknownField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			_, err = validator.Validate(ctx, ast)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error message should contain %q, got: %v", tt.errMsg, err)
			}
		})
	}
}
