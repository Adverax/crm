package engine

import (
	"testing"
)

func TestParseSimpleSelect(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple select",
			query:   "SELECT Name FROM Account",
			wantErr: false,
		},
		{
			name:    "multiple fields",
			query:   "SELECT Name, Email, Phone FROM Contact",
			wantErr: false,
		},
		{
			name:    "dot notation",
			query:   "SELECT Name, Account.Name FROM Contact",
			wantErr: false,
		},
		{
			name:    "deep dot notation",
			query:   "SELECT Name, Account.Owner.Name FROM Contact",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && g == nil {
				t.Error("Parse() returned nil grammar")
			}
		})
	}
}

func TestParseSelectRow(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantIsRow bool
		wantErr   bool
	}{
		{
			name:      "SELECT ROW with single field",
			query:     "SELECT ROW Id FROM Account WHERE Id = '123'",
			wantIsRow: true,
		},
		{
			name:      "SELECT ROW with multiple fields",
			query:     "SELECT ROW Id, Name, Email FROM Contact WHERE Id = '123'",
			wantIsRow: true,
		},
		{
			name:      "SELECT ROW case insensitive",
			query:     "select row Id, Name from Account where Id = '123'",
			wantIsRow: true,
		},
		{
			name:      "regular SELECT is not row",
			query:     "SELECT Id, Name FROM Account",
			wantIsRow: false,
		},
		{
			name:      "SELECT ROW with WHERE and LIMIT",
			query:     "SELECT ROW Id, Name FROM Account WHERE Id = '123' LIMIT 1",
			wantIsRow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if g == nil {
					t.Fatal("Parse() returned nil grammar")
				}
				if g.IsRow != tt.wantIsRow {
					t.Errorf("IsRow = %v, want %v", g.IsRow, tt.wantIsRow)
				}
			}
		})
	}
}

func TestIsRowQuery(t *testing.T) {
	tests := []struct {
		name string
		soql string
		want bool
	}{
		{
			name: "SELECT ROW query",
			soql: "SELECT ROW Id, Name FROM Account WHERE Id = :id",
			want: true,
		},
		{
			name: "regular SELECT query",
			soql: "SELECT Id, Name FROM Contact WHERE AccountId = :id",
			want: false,
		},
		{
			name: "invalid SOQL returns false",
			soql: "NOT VALID SQL",
			want: false,
		},
		{
			name: "empty string returns false",
			soql: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRowQuery(tt.soql)
			if got != tt.want {
				t.Errorf("IsRowQuery(%q) = %v, want %v", tt.soql, got, tt.want)
			}
		})
	}
}

func TestParseWhere(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple equality",
			query:   "SELECT Name FROM Account WHERE Name = 'Acme'",
			wantErr: false,
		},
		{
			name:    "not equal",
			query:   "SELECT Name FROM Account WHERE Name != 'Test'",
			wantErr: false,
		},
		{
			name:    "comparison operators",
			query:   "SELECT Name FROM Account WHERE Amount > 1000",
			wantErr: false,
		},
		{
			name:    "AND condition",
			query:   "SELECT Name FROM Account WHERE Name = 'Acme' AND Industry = 'Tech'",
			wantErr: false,
		},
		{
			name:    "OR condition",
			query:   "SELECT Name FROM Account WHERE Name = 'Acme' OR Name = 'Test'",
			wantErr: false,
		},
		{
			name:    "NOT condition",
			query:   "SELECT Name FROM Account WHERE NOT IsDeleted = TRUE",
			wantErr: false,
		},
		{
			name:    "IS NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NULL",
			wantErr: false,
		},
		{
			name:    "IS NOT NULL",
			query:   "SELECT Name FROM Contact WHERE Email IS NOT NULL",
			wantErr: false,
		},
		{
			name:    "IN clause",
			query:   "SELECT Name FROM Account WHERE Industry IN ('Tech', 'Finance', 'Healthcare')",
			wantErr: false,
		},
		{
			name:    "NOT IN clause",
			query:   "SELECT Name FROM Account WHERE Industry NOT IN ('Government')",
			wantErr: false,
		},
		{
			name:    "LIKE clause",
			query:   "SELECT Name FROM Account WHERE Name LIKE 'Acme%'",
			wantErr: false,
		},
		{
			name:    "parentheses",
			query:   "SELECT Name FROM Account WHERE (Industry = 'Tech' OR Industry = 'Finance') AND Revenue > 1000000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && g.Where == nil {
				t.Error("Parse() WHERE clause is nil")
			}
		})
	}
}

func TestParseDateLiterals(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "TODAY",
			query:   "SELECT Name FROM Account WHERE CreatedDate = TODAY",
			wantErr: false,
		},
		{
			name:    "YESTERDAY",
			query:   "SELECT Name FROM Account WHERE CreatedDate = YESTERDAY",
			wantErr: false,
		},
		{
			name:    "THIS_WEEK",
			query:   "SELECT Name FROM Account WHERE CreatedDate = THIS_WEEK",
			wantErr: false,
		},
		{
			name:    "LAST_MONTH",
			query:   "SELECT Name FROM Account WHERE CreatedDate = LAST_MONTH",
			wantErr: false,
		},
		{
			name:    "THIS_YEAR",
			query:   "SELECT Name FROM Account WHERE CreatedDate = THIS_YEAR",
			wantErr: false,
		},
		{
			name:    "LAST_N_DAYS",
			query:   "SELECT Name FROM Account WHERE CreatedDate = LAST_N_DAYS:30",
			wantErr: false,
		},
		{
			name:    "NEXT_N_MONTHS",
			query:   "SELECT Name FROM Opportunity WHERE CloseDate = NEXT_N_MONTHS:3",
			wantErr: false,
		},
		{
			name:    "ISO date",
			query:   "SELECT Name FROM Account WHERE CreatedDate > 2024-01-15",
			wantErr: false,
		},
		{
			name:    "ISO datetime",
			query:   "SELECT Name FROM Account WHERE CreatedDate > 2024-01-15T10:30:00Z",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple order by",
			query:   "SELECT Name FROM Account ORDER BY Name",
			wantErr: false,
		},
		{
			name:    "order by ASC",
			query:   "SELECT Name FROM Account ORDER BY Name ASC",
			wantErr: false,
		},
		{
			name:    "order by DESC",
			query:   "SELECT Name FROM Account ORDER BY Name DESC",
			wantErr: false,
		},
		{
			name:    "multiple order by",
			query:   "SELECT Name FROM Account ORDER BY Industry ASC, Name DESC",
			wantErr: false,
		},
		{
			name:    "NULLS FIRST",
			query:   "SELECT Name FROM Account ORDER BY Phone ASC NULLS FIRST",
			wantErr: false,
		},
		{
			name:    "NULLS LAST",
			query:   "SELECT Name FROM Account ORDER BY Phone DESC NULLS LAST",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(g.OrderBy) == 0 {
				t.Error("Parse() ORDER BY clause is empty")
			}
		})
	}
}

func TestParseLimitOffset(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  *int
		wantOffset *int
		wantErr    bool
	}{
		{
			name:      "LIMIT only",
			query:     "SELECT Name FROM Account LIMIT 10",
			wantLimit: intPtr(10),
			wantErr:   false,
		},
		{
			name:       "LIMIT and OFFSET",
			query:      "SELECT Name FROM Account LIMIT 10 OFFSET 20",
			wantLimit:  intPtr(10),
			wantOffset: intPtr(20),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantLimit != nil {
				if g.Limit == nil || *g.Limit != *tt.wantLimit {
					t.Errorf("Parse() LIMIT = %v, want %v", g.Limit, *tt.wantLimit)
				}
			}
			if tt.wantOffset != nil {
				if g.Offset == nil || *g.Offset != *tt.wantOffset {
					t.Errorf("Parse() OFFSET = %v, want %v", g.Offset, *tt.wantOffset)
				}
			}
		})
	}
}

func TestParseAggregate(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "COUNT",
			query:   "SELECT COUNT(Id) FROM Account",
			wantErr: false,
		},
		{
			name:    "SUM",
			query:   "SELECT SUM(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "AVG",
			query:   "SELECT AVG(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "MIN",
			query:   "SELECT MIN(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "MAX",
			query:   "SELECT MAX(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "multiple aggregates",
			query:   "SELECT COUNT(Id), SUM(Amount), AVG(Amount) FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "aggregate with alias",
			query:   "SELECT COUNT(Id) total FROM Account",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(g.Select) == 0 {
				t.Error("Parse() SELECT clause is empty")
			}
		})
	}
}

func TestParseGroupBy(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple group by",
			query:   "SELECT Industry, COUNT(Id) FROM Account GROUP BY Industry",
			wantErr: false,
		},
		{
			name:    "multiple group by",
			query:   "SELECT Industry, Type, COUNT(Id) FROM Account GROUP BY Industry, Type",
			wantErr: false,
		},
		{
			name:    "group by with having",
			query:   "SELECT Industry, COUNT(Id) cnt FROM Account GROUP BY Industry HAVING COUNT(Id) > 5",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(g.GroupBy) == 0 {
				t.Error("Parse() GROUP BY clause is empty")
			}
		})
	}
}

func TestParseRelationshipSubquery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple subquery",
			query:   "SELECT Name, (SELECT FirstName, LastName FROM Contacts) FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with where",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts WHERE Email IS NOT NULL) FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with order and limit",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts ORDER BY LastName LIMIT 5) FROM Account",
			wantErr: false,
		},
		{
			name:    "multiple subqueries",
			query:   "SELECT Name, (SELECT Subject FROM Tasks), (SELECT Subject FROM Events) FROM Account",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				hasSubquery := false
				for _, sel := range g.Select {
					if sel.Item != nil && sel.Item.Subquery != nil {
						hasSubquery = true
						break
					}
				}
				if !hasSubquery {
					t.Error("Parse() expected subquery in SELECT")
				}
			}
		})
	}
}

func TestParseStringLiterals(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple string",
			query:   "SELECT Name FROM Account WHERE Name = 'Acme Corp'",
			wantErr: false,
		},
		{
			name:    "escaped quote",
			query:   "SELECT Name FROM Account WHERE Name = 'O''Brien'",
			wantErr: false,
		},
		{
			name:    "empty string",
			query:   "SELECT Name FROM Account WHERE Name = ''",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseComplexQueries(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name: "complex query with all features",
			query: `SELECT
				Name,
				Industry,
				Account.Owner.Name,
				(SELECT FirstName, Email FROM Contacts WHERE Email IS NOT NULL ORDER BY LastName LIMIT 10)
			FROM Account
			WHERE Industry IN ('Tech', 'Finance')
				AND CreatedDate = LAST_N_DAYS:30
				AND Name LIKE 'Acme%'
			ORDER BY Name ASC NULLS LAST
			LIMIT 100
			OFFSET 20`,
			wantErr: false,
		},
		{
			name: "aggregate with group by and having",
			query: `SELECT
				Account.Industry,
				SUM(Amount) totalAmount,
				COUNT(Id) dealCount
			FROM Opportunity
			WHERE StageName = 'Closed Won'
				AND CloseDate = THIS_YEAR
			GROUP BY Account.Industry
			HAVING COUNT(Id) > 5
			ORDER BY SUM(Amount) DESC
			LIMIT 20`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "multiplication in SELECT",
			query:   "SELECT Amount * 0.1 FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "addition in SELECT",
			query:   "SELECT Price + Tax FROM LineItem",
			wantErr: false,
		},
		{
			name:    "subtraction in SELECT",
			query:   "SELECT Amount - Discount FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "division in SELECT",
			query:   "SELECT Total / 100 FROM Invoice",
			wantErr: false,
		},
		{
			name:    "modulo in SELECT",
			query:   "SELECT Quantity % 10 FROM LineItem",
			wantErr: false,
		},
		{
			name:    "arithmetic with alias",
			query:   "SELECT Amount * 0.1 taxAmount FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "arithmetic in WHERE",
			query:   "SELECT Name FROM Account WHERE Amount * 2 > 1000",
			wantErr: false,
		},
		{
			name:    "complex arithmetic in WHERE",
			query:   "SELECT Name FROM Account WHERE Price + Tax > 500",
			wantErr: false,
		},
		{
			name:    "parenthesized arithmetic",
			query:   "SELECT (Amount + Tax) * 0.9 FROM Opportunity",
			wantErr: false,
		},
		{
			name:    "multiple arithmetic operations",
			query:   "SELECT Amount * Rate + Fee FROM Transaction",
			wantErr: false,
		},
		{
			name:    "field arithmetic in WHERE",
			query:   "SELECT Name FROM Account WHERE Amount + Bonus > Budget - Expenses",
			wantErr: false,
		},
		{
			name:    "unary minus",
			query:   "SELECT -Amount FROM Account",
			wantErr: false,
		},
		{
			name:    "unary plus",
			query:   "SELECT +Amount FROM Account",
			wantErr: false,
		},
		{
			name:    "arithmetic with lookup",
			query:   "SELECT Account.Amount * 0.1 FROM Contact",
			wantErr: false,
		},
		{
			name:    "string concatenation",
			query:   "SELECT FirstName || ' ' || LastName FROM Contact",
			wantErr: false,
		},
		{
			name:    "string concatenation with field",
			query:   "SELECT Name || ' - ' || Industry FROM Account",
			wantErr: false,
		},
		{
			name:    "string concatenation in WHERE",
			query:   "SELECT Name FROM Contact WHERE FirstName || LastName = 'JohnDoe'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && g == nil {
				t.Error("Parse() returned nil grammar")
			}
		})
	}
}

func TestParseWhereSubquery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "simple IN subquery",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
		{
			name:    "NOT IN subquery",
			query:   "SELECT Name FROM Account WHERE Id NOT IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
		{
			name:    "subquery with WHERE",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact WHERE Email IS NOT NULL)",
			wantErr: false,
		},
		{
			name:    "subquery with LIMIT",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact LIMIT 100)",
			wantErr: false,
		},
		{
			name:    "combined with AND",
			query:   "SELECT Name FROM Account WHERE Type = 'Customer' AND Id IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
		{
			name:    "combined with OR",
			query:   "SELECT Name FROM Account WHERE Type = 'Prospect' OR Id IN (SELECT AccountId FROM Opportunity WHERE StageName = 'Closed Won')",
			wantErr: false,
		},
		{
			name:    "subquery with multiple WHERE conditions",
			query:   "SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact WHERE Email IS NOT NULL AND Status = 'Active')",
			wantErr: false,
		},
		{
			name:    "multiple fields in main query",
			query:   "SELECT Name, Industry, Phone FROM Account WHERE Id IN (SELECT AccountId FROM Contact)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if g == nil {
					t.Error("Parse() returned nil grammar")
					return
				}
				if g.Where == nil {
					t.Error("Parse() WHERE clause is nil")
					return
				}
				// Verify the subquery was parsed
				hasSubquery := findWhereSubqueryInExpression(g.Where)
				if !hasSubquery {
					t.Error("Parse() expected WHERE subquery in WHERE clause")
				}
			}
		})
	}
}

// findWhereSubqueryInExpression recursively checks if an expression contains a WHERE subquery.
func findWhereSubqueryInExpression(expr *Expression) bool {
	if expr == nil || expr.Or == nil {
		return false
	}
	for _, and := range expr.Or.And {
		for _, not := range and.Not {
			if not.Compare != nil {
				if findWhereSubqueryInInExpr(not.Compare.Left) {
					return true
				}
				if findWhereSubqueryInInExpr(not.Compare.Right) {
					return true
				}
			}
		}
	}
	return false
}

// findWhereSubqueryInInExpr checks if an InExpr contains a WHERE subquery.
func findWhereSubqueryInInExpr(in *InExpr) bool {
	if in == nil {
		return false
	}
	return in.Subquery != nil
}

func intPtr(i int) *int {
	return &i
}

func TestParseFunctions(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		// String functions
		{"COALESCE single arg", "SELECT COALESCE(Name) FROM Account", false},
		{"COALESCE multiple args", "SELECT COALESCE(Name, 'Unknown') FROM Account", false},
		{"COALESCE three args", "SELECT COALESCE(FirstName, LastName, 'N/A') FROM Contact", false},
		{"NULLIF", "SELECT NULLIF(Status, 'Inactive') FROM Account", false},
		{"CONCAT", "SELECT CONCAT(FirstName, LastName) FROM Contact", false},
		{"CONCAT multiple args", "SELECT CONCAT(FirstName, ' ', LastName) FROM Contact", false},
		{"UPPER", "SELECT UPPER(Name) FROM Account", false},
		{"LOWER", "SELECT LOWER(Email) FROM Contact", false},
		{"TRIM", "SELECT TRIM(Description) FROM Account", false},
		{"LENGTH", "SELECT LENGTH(Name) FROM Account", false},
		{"LEN alias", "SELECT LEN(Name) FROM Account", false},
		{"SUBSTRING two args", "SELECT SUBSTRING(Name, 1) FROM Account", false},
		{"SUBSTRING three args", "SELECT SUBSTRING(Name, 1, 10) FROM Account", false},
		{"SUBSTR alias", "SELECT SUBSTR(Name, 1, 5) FROM Account", false},

		// Math functions
		{"ABS", "SELECT ABS(Amount) FROM Opportunity", false},
		{"ROUND single arg", "SELECT ROUND(Price) FROM Product", false},
		{"ROUND with decimals", "SELECT ROUND(Price, 2) FROM Product", false},
		{"FLOOR", "SELECT FLOOR(Amount) FROM Opportunity", false},
		{"CEIL", "SELECT CEIL(Amount) FROM Opportunity", false},
		{"CEILING alias", "SELECT CEILING(Amount) FROM Opportunity", false},

		// Functions in WHERE clause
		{"function in WHERE", "SELECT Name FROM Account WHERE UPPER(Name) = 'TEST'", false},
		{"function in WHERE with LIKE", "SELECT Name FROM Account WHERE LOWER(Name) LIKE '%test%'", false},
		{"COALESCE in WHERE", "SELECT Name FROM Account WHERE COALESCE(Status, 'Unknown') = 'Active'", false},
		{"LENGTH in WHERE", "SELECT Name FROM Account WHERE LENGTH(Name) > 10", false},

		// Nested functions
		{"nested functions", "SELECT UPPER(TRIM(Name)) FROM Account", false},
		{"nested with COALESCE", "SELECT COALESCE(TRIM(Name), 'N/A') FROM Account", false},
		{"CONCAT with UPPER/LOWER", "SELECT CONCAT(UPPER(FirstName), ' ', LOWER(LastName)) FROM Contact", false},

		// Functions with expressions
		{"function with arithmetic", "SELECT ROUND(Price * 1.1, 2) FROM Product", false},
		{"ABS with subtraction", "SELECT ABS(Amount - Discount) FROM Opportunity", false},

		// Function as alias
		{"function with alias", "SELECT UPPER(Name) AS UpperName FROM Account", false},
		{"LENGTH with alias", "SELECT LENGTH(Description) AS DescLen FROM Account", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"empty query", ""},
		{"missing SELECT", "Name FROM Account"},
		{"missing FROM", "SELECT Name"},
		{"missing object name", "SELECT Name FROM"},
		{"missing field", "SELECT FROM Account"},
		{"incomplete WHERE", "SELECT Name FROM Account WHERE"},
		{"incomplete AND", "SELECT Name FROM Account WHERE Name = 'Test' AND"},
		{"incomplete OR", "SELECT Name FROM Account WHERE Name = 'Test' OR"},
		{"unmatched paren left", "SELECT Name FROM Account WHERE (Name = 'Test'"},
		{"unmatched paren right", "SELECT Name FROM Account WHERE Name = 'Test')"},
		{"invalid operator", "SELECT Name FROM Account WHERE Name <=> 'Test'"},
		{"incomplete IN clause", "SELECT Name FROM Account WHERE Name IN"},
		{"empty IN clause", "SELECT Name FROM Account WHERE Name IN ()"},
		{"incomplete LIKE", "SELECT Name FROM Account WHERE Name LIKE"},
		{"incomplete ORDER BY", "SELECT Name FROM Account ORDER BY"},
		{"incomplete GROUP BY", "SELECT Name FROM Account GROUP BY"},
		{"incomplete HAVING", "SELECT Name FROM Account GROUP BY Name HAVING"},
		{"invalid LIMIT", "SELECT Name FROM Account LIMIT abc"},
		{"invalid OFFSET", "SELECT Name FROM Account LIMIT 10 OFFSET abc"},
		{"incomplete subquery", "SELECT Name, (SELECT FROM Contacts) FROM Account"},
		{"dangling comma in SELECT", "SELECT Name, FROM Account"},
		{"dangling comma in IN", "SELECT Name FROM Account WHERE Name IN ('a', )"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if err == nil {
				t.Errorf("Parse(%q) should have failed", tt.query)
			}
		})
	}
}

func TestParseBooleanLiterals(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"TRUE uppercase", "SELECT Name FROM Account WHERE IsActive = TRUE", false},
		{"FALSE uppercase", "SELECT Name FROM Account WHERE IsActive = FALSE", false},
		{"true lowercase", "SELECT Name FROM Account WHERE IsActive = true", false},
		{"false lowercase", "SELECT Name FROM Account WHERE IsActive = false", false},
		{"True mixed", "SELECT Name FROM Account WHERE IsActive = True", false},
		{"False mixed", "SELECT Name FROM Account WHERE IsActive = False", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseNumericLiterals(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"integer", "SELECT Name FROM Account WHERE Amount = 100", false},
		{"decimal", "SELECT Name FROM Account WHERE Amount = 100.50", false},
		{"negative integer", "SELECT Name FROM Account WHERE Amount = -100", false},
		{"negative decimal", "SELECT Name FROM Account WHERE Amount = -100.50", false},
		{"zero", "SELECT Name FROM Account WHERE Amount = 0", false},
		{"large integer", "SELECT Name FROM Account WHERE Amount = 999999999", false},
		{"small decimal", "SELECT Name FROM Account WHERE Amount = 0.001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCaseSensitivity(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"lowercase keywords", "select Name from Account where Name = 'Test'", false},
		{"uppercase keywords", "SELECT Name FROM Account WHERE Name = 'Test'", false},
		{"mixed case keywords", "SeLeCt Name FrOm Account WhErE Name = 'Test'", false},
		{"lowercase ORDER BY", "SELECT Name FROM Account order by Name", false},
		{"lowercase GROUP BY", "SELECT Name FROM Account group by Name", false},
		{"lowercase LIMIT", "SELECT Name FROM Account limit 10", false},
		{"lowercase ASC", "SELECT Name FROM Account ORDER BY Name asc", false},
		{"lowercase DESC", "SELECT Name FROM Account ORDER BY Name desc", false},
		{"lowercase NULLS FIRST", "SELECT Name FROM Account ORDER BY Name nulls first", false},
		{"lowercase AND", "SELECT Name FROM Account WHERE A = 1 and B = 2", false},
		{"lowercase OR", "SELECT Name FROM Account WHERE A = 1 or B = 2", false},
		{"lowercase NOT", "SELECT Name FROM Account WHERE not A = 1", false},
		{"lowercase IS NULL", "SELECT Name FROM Account WHERE A is null", false},
		{"lowercase IS NOT NULL", "SELECT Name FROM Account WHERE A is not null", false},
		{"lowercase IN", "SELECT Name FROM Account WHERE A in ('x')", false},
		{"lowercase NOT IN", "SELECT Name FROM Account WHERE A not in ('x')", false},
		{"lowercase LIKE", "SELECT Name FROM Account WHERE A like 'x%'", false},
		{"lowercase aggregate", "SELECT count(Id), sum(Amount) FROM Account", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseNestedExpressions(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "deeply nested parentheses",
			query:   "SELECT Name FROM Account WHERE ((((A = 1))))",
			wantErr: false,
		},
		{
			name:    "complex nested AND/OR",
			query:   "SELECT Name FROM Account WHERE (A = 1 AND B = 2) OR (C = 3 AND D = 4)",
			wantErr: false,
		},
		{
			name:    "triple nested",
			query:   "SELECT Name FROM Account WHERE ((A = 1 OR B = 2) AND (C = 3 OR D = 4)) OR E = 5",
			wantErr: false,
		},
		{
			name:    "NOT with nested",
			query:   "SELECT Name FROM Account WHERE NOT (A = 1 AND B = 2)",
			wantErr: false,
		},
		{
			name:    "nested NOT",
			query:   "SELECT Name FROM Account WHERE NOT (NOT A = 1)",
			wantErr: false,
		},
		{
			name:    "complex with all operators",
			query:   "SELECT Name FROM Account WHERE (A > 1 AND B < 2) OR (C >= 3 AND D <= 4) AND NOT (E = 5)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAliases(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"simple alias", "SELECT Name n FROM Account", false},
		{"alias with AS", "SELECT Name AS n FROM Account", false},
		{"multiple aliases", "SELECT Name n, Industry ind FROM Account", false},
		{"aggregate alias", "SELECT COUNT(Id) total FROM Account", false},
		{"aggregate alias with AS", "SELECT COUNT(Id) AS total FROM Account", false},
		{"expression alias", "SELECT Amount * 0.1 tax FROM Opportunity", false},
		{"expression alias with AS", "SELECT Amount * 0.1 AS tax FROM Opportunity", false},
		{"function alias", "SELECT UPPER(Name) upperName FROM Account", false},
		{"function alias with AS", "SELECT UPPER(Name) AS upperName FROM Account", false},
		{"lookup alias", "SELECT Account.Name accName FROM Contact", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && g != nil && len(g.Select) > 0 {
				// Verify that at least one select item has an alias
				hasAlias := false
				for _, sel := range g.Select {
					if sel.Alias != nil && *sel.Alias != "" {
						hasAlias = true
						break
					}
				}
				if !hasAlias {
					t.Error("expected alias in select")
				}
			}
		})
	}
}

func TestParseAllDateLiterals(t *testing.T) {
	staticLiterals := []string{
		"TODAY", "YESTERDAY", "TOMORROW",
		"THIS_WEEK", "LAST_WEEK", "NEXT_WEEK",
		"THIS_MONTH", "LAST_MONTH", "NEXT_MONTH",
		"THIS_QUARTER", "LAST_QUARTER", "NEXT_QUARTER",
		"THIS_YEAR", "LAST_YEAR", "NEXT_YEAR",
		"THIS_FISCAL_QUARTER", "LAST_FISCAL_QUARTER", "NEXT_FISCAL_QUARTER",
		"THIS_FISCAL_YEAR", "LAST_FISCAL_YEAR", "NEXT_FISCAL_YEAR",
	}

	for _, lit := range staticLiterals {
		t.Run(lit, func(t *testing.T) {
			query := "SELECT Name FROM Account WHERE CreatedDate = " + lit
			_, err := Parse(query)
			if err != nil {
				t.Errorf("Parse() error for %s: %v", lit, err)
			}
		})
	}

	dynamicLiterals := []struct {
		prefix string
		n      int
	}{
		{"LAST_N_DAYS", 30},
		{"NEXT_N_DAYS", 7},
		{"LAST_N_WEEKS", 4},
		{"NEXT_N_WEEKS", 2},
		{"LAST_N_MONTHS", 3},
		{"NEXT_N_MONTHS", 6},
		{"LAST_N_QUARTERS", 2},
		{"NEXT_N_QUARTERS", 1},
		{"LAST_N_YEARS", 5},
		{"NEXT_N_YEARS", 1},
		// Fiscal dynamic periods
		{"LAST_N_FISCAL_QUARTERS", 2},
		{"NEXT_N_FISCAL_QUARTERS", 3},
		{"LAST_N_FISCAL_YEARS", 2},
		{"NEXT_N_FISCAL_YEARS", 1},
	}

	for _, lit := range dynamicLiterals {
		name := lit.prefix + ":N"
		t.Run(name, func(t *testing.T) {
			query := "SELECT Name FROM Account WHERE CreatedDate = " + lit.prefix + ":10"
			_, err := Parse(query)
			if err != nil {
				t.Errorf("Parse() error for %s: %v", name, err)
			}
		})
	}
}

func TestParseSpecialCharactersInStrings(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"single quote escape", "SELECT Name FROM Account WHERE Name = 'O''Brien'", false},
		{"double single quote", "SELECT Name FROM Account WHERE Name = 'It''s a test'", false},
		{"unicode characters", "SELECT Name FROM Account WHERE Name = 'Компания'", false},
		{"emoji", "SELECT Name FROM Account WHERE Name = 'Test 🎉'", false},
		{"special SQL chars", "SELECT Name FROM Account WHERE Name = 'Test % _ chars'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseMultipleSubqueries(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		count   int
	}{
		{
			name:    "single subquery",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts) FROM Account",
			wantErr: false,
			count:   1,
		},
		{
			name:    "two subqueries",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts), (SELECT Subject FROM Tasks) FROM Account",
			wantErr: false,
			count:   2,
		},
		{
			name:    "three subqueries",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts), (SELECT Subject FROM Tasks), (SELECT Name FROM Opportunities) FROM Account",
			wantErr: false,
			count:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				subqueryCount := 0
				for _, sel := range g.Select {
					if sel.Item != nil && sel.Item.Subquery != nil {
						subqueryCount++
					}
				}
				if subqueryCount != tt.count {
					t.Errorf("subquery count = %d, want %d", subqueryCount, tt.count)
				}
			}
		})
	}
}

func TestParseComplexSubqueries(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "subquery with multiple fields",
			query:   "SELECT Name, (SELECT FirstName, LastName, Email, Phone FROM Contacts) FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with complex WHERE",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts WHERE Email IS NOT NULL AND Status = 'Active') FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with ORDER BY and LIMIT",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts ORDER BY LastName DESC LIMIT 5) FROM Account",
			wantErr: false,
		},
		{
			name:    "subquery with NULLS handling",
			query:   "SELECT Name, (SELECT FirstName FROM Contacts ORDER BY Email NULLS LAST) FROM Account",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseComparisonOperators(t *testing.T) {
	operators := []struct {
		op   string
		name string
	}{
		{"=", "equals"},
		{"!=", "not equals"},
		{"<>", "not equals alt"},
		{">", "greater than"},
		{"<", "less than"},
		{">=", "greater or equal"},
		{"<=", "less or equal"},
	}

	for _, op := range operators {
		t.Run(op.name, func(t *testing.T) {
			query := "SELECT Name FROM Account WHERE Amount " + op.op + " 100"
			_, err := Parse(query)
			if err != nil {
				t.Errorf("Parse() error for %s: %v", op.op, err)
			}
		})
	}
}

func TestParseModuloOperator(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"modulo in SELECT", "SELECT Amount % 10 FROM Account", false},
		{"modulo in WHERE", "SELECT Name FROM Account WHERE Amount % 2 = 0", false},
		{"modulo with parentheses", "SELECT (Amount + 5) % 10 FROM Account", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseWhitespaceHandling(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"minimal whitespace", "SELECT Name FROM Account", false},
		{"extra spaces", "SELECT   Name   FROM   Account", false},
		{"tabs", "SELECT\tName\tFROM\tAccount", false},
		{"newlines", "SELECT\nName\nFROM\nAccount", false},
		{"mixed whitespace", "SELECT \t\n Name \t\n FROM \t\n Account", false},
		{
			name: "formatted query",
			query: `
				SELECT
					Name,
					Industry
				FROM
					Account
				WHERE
					Name = 'Test'
				ORDER BY
					Name
			`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseWithSecurityEnforced(t *testing.T) {
	tests := []struct {
		name                 string
		query                string
		wantErr              bool
		withSecurityEnforced bool
	}{
		{
			name:                 "simple WITH SECURITY_ENFORCED",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with WHERE",
			query:                "SELECT Name FROM Account WHERE Industry = 'Tech' WITH SECURITY_ENFORCED",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with ORDER BY",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED ORDER BY Name",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with GROUP BY",
			query:                "SELECT Industry, COUNT(Id) FROM Account WITH SECURITY_ENFORCED GROUP BY Industry",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with LIMIT",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED LIMIT 10",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED with all clauses",
			query:                "SELECT Name FROM Account WHERE Active = TRUE WITH SECURITY_ENFORCED ORDER BY Name LIMIT 10",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "WITH SECURITY_ENFORCED and FOR UPDATE",
			query:                "SELECT Name FROM Account WITH SECURITY_ENFORCED FOR UPDATE",
			wantErr:              false,
			withSecurityEnforced: true,
		},
		{
			name:                 "without WITH SECURITY_ENFORCED",
			query:                "SELECT Name FROM Account",
			wantErr:              false,
			withSecurityEnforced: false,
		},
		{
			name:                 "WITH SECURITY_ENFORCED case insensitive",
			query:                "SELECT Name FROM Account with security_enforced",
			wantErr:              false,
			withSecurityEnforced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ast.WithSecurityEnforced != tt.withSecurityEnforced {
				t.Errorf("WithSecurityEnforced = %v, want %v", ast.WithSecurityEnforced, tt.withSecurityEnforced)
			}
		})
	}
}

func TestParseForUpdate(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantErr   bool
		forUpdate bool
	}{
		{
			name:      "simple FOR UPDATE",
			query:     "SELECT Name FROM Account FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with WHERE",
			query:     "SELECT Name FROM Account WHERE Industry = 'Tech' FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with ORDER BY",
			query:     "SELECT Name FROM Account ORDER BY Name FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with LIMIT",
			query:     "SELECT Name FROM Account LIMIT 10 FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with OFFSET",
			query:     "SELECT Name FROM Account LIMIT 10 OFFSET 5 FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "FOR UPDATE with all clauses",
			query:     "SELECT Name FROM Account WHERE Active = TRUE ORDER BY Name LIMIT 10 FOR UPDATE",
			wantErr:   false,
			forUpdate: true,
		},
		{
			name:      "without FOR UPDATE",
			query:     "SELECT Name FROM Account",
			wantErr:   false,
			forUpdate: false,
		},
		{
			name:      "FOR UPDATE case insensitive",
			query:     "SELECT Name FROM Account for update",
			wantErr:   false,
			forUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ast.ForUpdate != tt.forUpdate {
				t.Errorf("ForUpdate = %v, want %v", ast.ForUpdate, tt.forUpdate)
			}
		})
	}
}

func TestParseTypeof(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		wantErr        bool
		fieldName      string
		whenClausesCnt int
		hasElse        bool
	}{
		{
			name: "simple TYPEOF with single WHEN",
			query: `SELECT
				TYPEOF What
					WHEN Account THEN Name, Industry
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "What",
			whenClausesCnt: 1,
			hasElse:        false,
		},
		{
			name: "TYPEOF with multiple WHEN clauses",
			query: `SELECT
				TYPEOF What
					WHEN Account THEN Name, Industry
					WHEN Opportunity THEN Name, StageName, Amount
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "What",
			whenClausesCnt: 2,
			hasElse:        false,
		},
		{
			name: "TYPEOF with ELSE clause",
			query: `SELECT
				TYPEOF What
					WHEN Account THEN Name, Industry
					WHEN Opportunity THEN Name, StageName
					ELSE Name
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "What",
			whenClausesCnt: 2,
			hasElse:        true,
		},
		{
			name: "TYPEOF with ELSE multiple fields",
			query: `SELECT
				TYPEOF What
					WHEN Account THEN Name
					ELSE Name, Id
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "What",
			whenClausesCnt: 1,
			hasElse:        true,
		},
		{
			name: "TYPEOF with other fields in SELECT",
			query: `SELECT
				Id, Subject,
				TYPEOF What
					WHEN Account THEN Name, Industry
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "What",
			whenClausesCnt: 1,
			hasElse:        false,
		},
		{
			name:           "TYPEOF case insensitive",
			query:          "SELECT typeof what when Account then Name end FROM Task",
			wantErr:        false,
			fieldName:      "what",
			whenClausesCnt: 1,
			hasElse:        false,
		},
		{
			name: "TYPEOF with single field in WHEN",
			query: `SELECT
				TYPEOF WhoId
					WHEN Contact THEN Email
					WHEN Lead THEN Email
				END
			FROM Task`,
			wantErr:        false,
			fieldName:      "WhoId",
			whenClausesCnt: 2,
			hasElse:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Find TYPEOF expression in SELECT
			var typeof *TypeofExpression
			for _, sel := range ast.Select {
				if sel.Item != nil && sel.Item.Typeof != nil {
					typeof = sel.Item.Typeof
					break
				}
			}

			if typeof == nil {
				t.Error("TYPEOF expression not found")
				return
			}

			if typeof.Field != tt.fieldName {
				t.Errorf("TYPEOF field = %v, want %v", typeof.Field, tt.fieldName)
			}

			if len(typeof.WhenClauses) != tt.whenClausesCnt {
				t.Errorf("WHEN clauses count = %v, want %v", len(typeof.WhenClauses), tt.whenClausesCnt)
			}

			hasElse := len(typeof.ElseFields) > 0
			if hasElse != tt.hasElse {
				t.Errorf("has ELSE = %v, want %v", hasElse, tt.hasElse)
			}
		})
	}
}

func TestParseTypeofWhenClauses(t *testing.T) {
	query := `SELECT
		TYPEOF What
			WHEN Account THEN Name, Industry, AnnualRevenue
			WHEN Opportunity THEN Name, StageName
		END
	FROM Task`

	ast, err := Parse(query)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find TYPEOF expression
	var typeof *TypeofExpression
	for _, sel := range ast.Select {
		if sel.Item != nil && sel.Item.Typeof != nil {
			typeof = sel.Item.Typeof
			break
		}
	}

	if typeof == nil {
		t.Fatal("TYPEOF expression not found")
	}

	// Verify WHEN clauses
	if len(typeof.WhenClauses) != 2 {
		t.Fatalf("Expected 2 WHEN clauses, got %d", len(typeof.WhenClauses))
	}

	// Check first WHEN clause (Account)
	when1 := typeof.WhenClauses[0]
	if when1.ObjectType != "Account" {
		t.Errorf("WHEN[0] ObjectType = %v, want Account", when1.ObjectType)
	}
	if len(when1.Fields) != 3 {
		t.Errorf("WHEN[0] fields count = %v, want 3", len(when1.Fields))
	}
	expectedFields1 := []string{"Name", "Industry", "AnnualRevenue"}
	for i, expected := range expectedFields1 {
		if i < len(when1.Fields) && when1.Fields[i] != expected {
			t.Errorf("WHEN[0].Fields[%d] = %v, want %v", i, when1.Fields[i], expected)
		}
	}

	// Check second WHEN clause (Opportunity)
	when2 := typeof.WhenClauses[1]
	if when2.ObjectType != "Opportunity" {
		t.Errorf("WHEN[1] ObjectType = %v, want Opportunity", when2.ObjectType)
	}
	if len(when2.Fields) != 2 {
		t.Errorf("WHEN[1] fields count = %v, want 2", len(when2.Fields))
	}
	expectedFields2 := []string{"Name", "StageName"}
	for i, expected := range expectedFields2 {
		if i < len(when2.Fields) && when2.Fields[i] != expected {
			t.Errorf("WHEN[1].Fields[%d] = %v, want %v", i, when2.Fields[i], expected)
		}
	}
}
