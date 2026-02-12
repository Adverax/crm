package engine

import "github.com/alecthomas/participle/v2/lexer"

// DMLStatement is the root AST node representing a DML statement.
// Only one of Insert, Update, Delete, or Upsert will be non-nil.
type DMLStatement struct {
	Pos    lexer.Position
	Insert *InsertStatement `  @@`
	Update *UpdateStatement `| @@`
	Delete *DeleteStatement `| @@`
	Upsert *UpsertStatement `| @@`
}

// GetOperation returns the operation type of this statement.
func (s *DMLStatement) GetOperation() Operation {
	switch {
	case s.Insert != nil:
		return OperationInsert
	case s.Update != nil:
		return OperationUpdate
	case s.Delete != nil:
		return OperationDelete
	case s.Upsert != nil:
		return OperationUpsert
	default:
		return OperationInsert
	}
}

// GetObject returns the object name this statement operates on.
func (s *DMLStatement) GetObject() string {
	switch {
	case s.Insert != nil:
		return s.Insert.Object
	case s.Update != nil:
		return s.Update.Object
	case s.Delete != nil:
		return s.Delete.Object
	case s.Upsert != nil:
		return s.Upsert.Object
	default:
		return ""
	}
}

// InsertStatement represents an INSERT statement.
// Example: INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech'), ('Globex', 'Finance')
type InsertStatement struct {
	Pos    lexer.Position
	Object string       `"INSERT" "INTO" @Ident`
	Fields []string     `"(" @Ident ("," @Ident)* ")"`
	Values []*ValueList `"VALUES" @@ ("," @@)*`
}

// ValueList represents a single row of values in INSERT/UPSERT.
// Example: ('Acme', 'Tech', 100) or (UPPER('test'), COALESCE(NULL, 'default'))
type ValueList struct {
	Pos    lexer.Position
	Values []*Expr `"(" @@ ("," @@)* ")"`
}

// UpdateStatement represents an UPDATE statement.
// Example: UPDATE Contact SET Status = 'Active', UpdatedAt = 2024-01-15 WHERE AccountId = 'acc-001'
type UpdateStatement struct {
	Pos         lexer.Position
	Object      string        `"UPDATE" @Ident`
	Assignments []*Assignment `"SET" @@ ("," @@)*`
	Where       *Expression   `("WHERE" @@)?`
}

// Assignment represents a field assignment in UPDATE.
// Example: Status = 'Active' or Name = UPPER('test')
type Assignment struct {
	Pos   lexer.Position
	Field string `@Ident "="`
	Value *Expr  `@@`
}

// DeleteStatement represents a DELETE statement.
// Example: DELETE FROM Task WHERE Status = 'Completed' AND CreatedDate < 2023-01-01
type DeleteStatement struct {
	Pos    lexer.Position
	Object string      `"DELETE" "FROM" @Ident`
	Where  *Expression `("WHERE" @@)?`
}

// UpsertStatement represents an UPSERT statement.
// Example: UPSERT Account (external_id, Name, Industry) VALUES ('ext-001', 'Acme', 'Tech') ON external_id
type UpsertStatement struct {
	Pos             lexer.Position
	Object          string       `"UPSERT" @Ident`
	Fields          []string     `"(" @Ident ("," @Ident)* ")"`
	Values          []*ValueList `"VALUES" @@ ("," @@)*`
	ExternalIdField string       `"ON" @Ident`
}

// =============================================================================
// WHERE clause expression AST (simplified from SOQL)
// =============================================================================

// Expression is the top-level expression node for WHERE clauses.
type Expression struct {
	Pos       lexer.Position
	Or        *OrExpr `@@`
	FieldType FieldType
}

// OrExpr represents OR expressions.
type OrExpr struct {
	And       []*AndExpr `@@ ("OR" @@)*`
	FieldType FieldType
}

// AndExpr represents AND expressions.
type AndExpr struct {
	Not       []*NotExpr `@@ ("AND" @@)*`
	FieldType FieldType
}

// NotExpr represents NOT expressions.
type NotExpr struct {
	Not       bool         `@"NOT"?`
	Compare   *CompareExpr `@@`
	FieldType FieldType
}

// CompareExpr represents comparison expressions.
type CompareExpr struct {
	Left      *InExpr   `@@`
	Operator  *Operator `(@("=" | "==" | "!=" | "<>" | ">" | "<" | ">=" | "<=")`
	Right     *InExpr   `@@)?`
	FieldType FieldType
}

// InExpr represents IN/NOT IN expressions.
type InExpr struct {
	Left      *LikeExpr `@@`
	Not       bool      `(@"NOT"?`
	In        bool      `@"IN"`
	Values    []*Value  `"(" @@ ("," @@)* ")")?`
	FieldType FieldType
}

// LikeExpr represents LIKE expressions.
type LikeExpr struct {
	Left      *IsExpr `@@`
	Not       bool    `(@"NOT"?`
	Like      bool    `@"LIKE"`
	Pattern   *Value  `@@)?`
	FieldType FieldType
}

// IsExpr represents IS NULL / IS NOT NULL expressions.
type IsExpr struct {
	Left      *Primary `@@`
	Is        bool     `(@"IS"`
	Not       bool     `@"NOT"?`
	Null      bool     `@"NULL")?`
	FieldType FieldType
}

// Primary represents primary expressions in WHERE clause.
type Primary struct {
	Subexpression *Expression `  "(" @@ ")"`
	Const         *Const      `| @@`
	Field         *Field      `| @@`
	FieldType     FieldType
}

// Value represents a value in IN clause or LIKE pattern.
type Value struct {
	Const     *Const `  @@`
	Field     *Field `| @@`
	FieldType FieldType
}

// Field represents a field reference in WHERE clause.
type Field struct {
	Pos       lexer.Position
	Name      string `@Ident`
	FieldType FieldType
}

// =============================================================================
// Value expressions for INSERT VALUES and UPDATE SET
// =============================================================================

// Expr represents a value expression in INSERT VALUES or UPDATE SET.
// Can be a function call, a constant, or a field reference.
type Expr struct {
	Pos       lexer.Position
	FuncCall  *FuncCall `  @@`
	Const     *Const    `| @@`
	Field     *Field    `| @@`
	FieldType FieldType
}

// Value returns the Go value of the expression (for simple constants only).
// For function calls, returns nil.
func (e *Expr) Value() any {
	if e.Const != nil {
		return e.Const.Value()
	}
	return nil
}

// IsConst returns true if this is a simple constant value.
func (e *Expr) IsConst() bool {
	return e.Const != nil
}

// IsFunc returns true if this is a function call.
func (e *Expr) IsFunc() bool {
	return e.FuncCall != nil
}

// IsField returns true if this is a field reference.
func (e *Expr) IsField() bool {
	return e.Field != nil
}

// FuncCall represents a function call.
// Example: UPPER('test'), COALESCE(Name, 'default'), ROUND(Amount, 2)
type FuncCall struct {
	Pos       lexer.Position
	Name      Function `@("COALESCE" | "NULLIF" | "CONCAT" | "UPPER" | "LOWER" | "TRIM" | "LENGTH" | "LEN" | "SUBSTRING" | "SUBSTR" | "ABS" | "ROUND" | "FLOOR" | "CEIL" | "CEILING")`
	Args      []*Expr  `"(" (@@ ("," @@)*)? ")"`
	FieldType FieldType
}
