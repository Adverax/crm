package engine

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// Grammar is the root AST node representing a SOQL query
type Grammar struct {
	Pos                  lexer.Position
	IsRow                bool                `parser:"'SELECT' @'ROW'?"`
	Select               []*SelectExpression `parser:"@@ (',' @@)*"`
	From                 string              `parser:"'FROM' @Ident"`
	Where                *Expression         `parser:"('WHERE' @@)?"`
	WithSecurityEnforced bool                `parser:"@('WITH' 'SECURITY_ENFORCED')?"`
	GroupBy              []*GroupClause      `parser:"('GROUP' 'BY' @@ (',' @@)*)?"`
	Having               *Expression         `parser:"('HAVING' @@)?"`
	OrderBy              []*OrderClause      `parser:"('ORDER' 'BY' @@ (',' @@)*)?"`
	Limit                *int                `parser:"('LIMIT' @Integer)?"`
	Offset               *int                `parser:"('OFFSET' @Integer)?"`
	ForUpdate            bool                `parser:"@('FOR' 'UPDATE')?"`
}

// SelectExpression represents a single item in SELECT clause
type SelectExpression struct {
	Pos   lexer.Position
	Item  *SelectItem `parser:"@@"`
	Alias *string     `parser:"(('AS' @Ident) | (?! 'FROM' | ',' | 'WHERE' | 'GROUP' | 'ORDER' | 'LIMIT' | 'OFFSET' | 'HAVING') @Ident)?"`
}

// SelectItem represents the actual content of a SELECT item
type SelectItem struct {
	Pos       lexer.Position
	Typeof    *TypeofExpression     `parser:"  @@"`
	Aggregate *AggregateExpression  `parser:"| @@"`
	Subquery  *RelationshipSubquery `parser:"| @@"`
	Expr      *Expression           `parser:"| @@"`
}

// TypeofExpression represents a TYPEOF expression for polymorphic fields
// Example: TYPEOF What WHEN Account THEN Name, Industry WHEN Opportunity THEN Name, Amount ELSE Name END
type TypeofExpression struct {
	Pos         lexer.Position
	Field       string        `parser:"'TYPEOF' @Ident"`
	WhenClauses []*WhenClause `parser:"@@+"`
	ElseFields  []string      `parser:"('ELSE' @Ident (',' @Ident)*)?"`
	End         string        `parser:"'END'"`
}

// WhenClause represents a single WHEN clause in TYPEOF
// Example: WHEN Account THEN Name, Industry
type WhenClause struct {
	Pos        lexer.Position
	ObjectType string   `parser:"'WHEN' @Ident"`
	Fields     []string `parser:"'THEN' @Ident (',' @Ident)*"`
}

// AggregateExpression represents an aggregate function call
type AggregateExpression struct {
	Pos        lexer.Position
	Function   Aggregate   `parser:"@('COUNT' | 'COUNT_DISTINCT' | 'SUM' | 'AVG' | 'MIN' | 'MAX')"`
	OpenParen  string      `parser:"'('"`
	Expression *Expression `parser:"@@"`
	CloseParen string      `parser:"')'"`
	FieldType  FieldType   // inferred type
}

// RelationshipSubquery represents a Parent-to-Child subquery in SELECT
// Example: (SELECT FirstName, Email FROM Contacts)
type RelationshipSubquery struct {
	Pos        lexer.Position
	OpenParen  string              `parser:"'('"`
	Select     []*SelectExpression `parser:"'SELECT' @@ (',' @@)*"`
	From       string              `parser:"'FROM' @Ident"`
	Where      *Expression         `parser:"('WHERE' @@)?"`
	OrderBy    []*OrderClause      `parser:"('ORDER' 'BY' @@ (',' @@)*)?"`
	Limit      *int                `parser:"('LIMIT' @Integer)?"`
	CloseParen string              `parser:"')'"`
}

// WhereSubquery represents a subquery in WHERE IN clause (semi-join)
// Example: SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact)
type WhereSubquery struct {
	Pos        lexer.Position
	OpenParen  string      `parser:"'('"`
	Select     *Expression `parser:"'SELECT' @@"` // Single field only
	From       string      `parser:"'FROM' @Ident"`
	Where      *Expression `parser:"('WHERE' @@)?"`
	Limit      *int        `parser:"('LIMIT' @Integer)?"`
	CloseParen string      `parser:"')'"`
}

// OrderClause represents a single ORDER BY item
type OrderClause struct {
	OrderItem *OrderItem  `parser:"@@"`
	Direction *Direction  `parser:"@('ASC' | 'DESC')?"`
	Nulls     *NullsOrder `parser:"('NULLS' @('FIRST' | 'LAST'))?"`
}

// OrderItem represents what to order by (field or aggregate)
type OrderItem struct {
	Pos       lexer.Position
	Aggregate *AggregateExpression `parser:"  @@"`
	Field     []string             `parser:"| @Ident ('.' @Ident)*"`
}

// GroupClause represents a single GROUP BY item
type GroupClause struct {
	Pos   lexer.Position
	Field []string `parser:"@Ident ('.' @Ident)*"`
}

// Expression is the top-level expression node
type Expression struct {
	Pos       lexer.Position
	Or        *OrExpr `parser:"@@"`
	FieldType FieldType
}

// OrExpr represents OR expressions
type OrExpr struct {
	And       []*AndExpr `parser:"@@ ('OR' @@)*"`
	FieldType FieldType
}

// AndExpr represents AND expressions
type AndExpr struct {
	Not       []*NotExpr `parser:"@@ ('AND' @@)*"`
	FieldType FieldType
}

// NotExpr represents NOT expressions
type NotExpr struct {
	Not       bool         `parser:"@'NOT'?"`
	Compare   *CompareExpr `parser:"@@"`
	FieldType FieldType
}

// CompareExpr represents comparison expressions
type CompareExpr struct {
	Left      *InExpr   `parser:"@@"`
	Operator  *Operator `parser:"(@('=' | '==' | '!=' | '<>' | '>' | '<' | '>=' | '<=')"`
	Right     *InExpr   `parser:"@@)?"`
	FieldType FieldType
}

// InExpr represents IN/NOT IN expressions
// Supports both literal values and subqueries:
//   - Id IN ('001', '002', '003')
//   - Id IN (SELECT AccountId FROM Contact WHERE Status = 'Active')
type InExpr struct {
	Left      *LikeExpr      `parser:"@@"`
	Not       bool           `parser:"(@'NOT'?"`
	In        bool           `parser:"@'IN'"`
	Subquery  *WhereSubquery `parser:"( @@"`
	Values    []*Value       `parser:"| '(' @@ (',' @@)* ')' ))?"`
	FieldType FieldType
}

// LikeExpr represents LIKE expressions
type LikeExpr struct {
	Left      *IsExpr `parser:"@@"`
	Not       bool    `parser:"(@'NOT'?"`
	Like      bool    `parser:"@'LIKE'"`
	Pattern   *Value  `parser:"@@)?"`
	FieldType FieldType
}

// IsExpr represents IS NULL / IS NOT NULL expressions
type IsExpr struct {
	Left      *AddExpr `parser:"@@"`
	Is        bool     `parser:"(@'IS'"`
	Not       bool     `parser:"@'NOT'?"`
	Null      bool     `parser:"@'NULL')?"`
	FieldType FieldType
}

// AddExpr represents addition/subtraction expressions
type AddExpr struct {
	Left      *MulExpr `parser:"@@"`
	Right     []*AddOp `parser:"@@*"`
	FieldType FieldType
}

// AddOp represents a single addition/subtraction/concatenation operation
type AddOp struct {
	Operator  Operator `parser:"@('+' | '-' | '||')"`
	Right     *MulExpr `parser:"@@"`
	FieldType FieldType
}

// MulExpr represents multiplication/division expressions
type MulExpr struct {
	Left      *UnaryExpr `parser:"@@"`
	Right     []*MulOp   `parser:"@@*"`
	FieldType FieldType
}

// MulOp represents a single multiplication/division operation
type MulOp struct {
	Operator  Operator   `parser:"@('*' | '/' | '%')"`
	Right     *UnaryExpr `parser:"@@"`
	FieldType FieldType
}

// UnaryExpr represents unary expressions (+/-)
type UnaryExpr struct {
	Operator  *Operator `parser:"@('+' | '-')?"`
	Primary   *Primary  `parser:"@@"`
	FieldType FieldType
}

// Primary represents primary expressions (literals, fields, function calls, etc.)
type Primary struct {
	Subexpression *Expression          `parser:"  '(' @@ ')'"`
	Aggregate     *AggregateExpression `parser:"| @@"`
	FuncCall      *FuncCall            `parser:"| @@"`
	Const         *Const               `parser:"| @@"`
	Field         *Field               `parser:"| @@"`
	FieldType     FieldType
}

// Value represents a value in IN clause or LIKE pattern
type Value struct {
	Const     *Const `parser:"  @@"`
	Field     *Field `parser:"| @@"`
	FieldType FieldType
}

// Field represents a field reference (possibly with dot notation)
type Field struct {
	Pos       lexer.Position
	Path      []string `parser:"@Ident ('.' @Ident)*"`
	FieldType FieldType
}

// FuncCall represents a function call
type FuncCall struct {
	Name      Function      `parser:"@('COALESCE' | 'NULLIF' | 'CONCAT' | 'UPPER' | 'LOWER' | 'TRIM' | 'LENGTH' | 'LEN' | 'SUBSTRING' | 'SUBSTR' | 'ABS' | 'ROUND' | 'FLOOR' | 'CEIL' | 'CEILING')"`
	Args      []*Expression `parser:"'(' @@ (',' @@)* ')'"`
	FieldType FieldType
}

// Const represents a constant value
type Const struct {
	DynamicDate *DynamicDateLiteral `parser:"  @DynamicDateLiteral"`
	StaticDate  *StaticDateLiteral  `parser:"| @StaticDateLiteral"`
	DateTime    *DateTime           `parser:"| @DateTime"`
	Date        *Date               `parser:"| @Date"`
	String      *string             `parser:"| @String"`
	Float       *float64            `parser:"| @Float"`
	Integer     *int                `parser:"| @Integer"`
	Boolean     *Boolean            `parser:"| @('TRUE' | 'FALSE')"`
	Null        bool                `parser:"| @'NULL'"`
	FieldType   FieldType
}

// GetFieldType returns the inferred type of the constant
func (c *Const) GetFieldType() FieldType {
	if c.FieldType != FieldTypeUnknown {
		return c.FieldType
	}

	switch {
	case c.String != nil:
		c.FieldType = FieldTypeString
	case c.Integer != nil:
		c.FieldType = FieldTypeInteger
	case c.Float != nil:
		c.FieldType = FieldTypeFloat
	case c.Boolean != nil:
		c.FieldType = FieldTypeBoolean
	case c.Date != nil:
		c.FieldType = FieldTypeDate
	case c.DateTime != nil:
		c.FieldType = FieldTypeDateTime
	case c.StaticDate != nil:
		c.FieldType = FieldTypeDate
	case c.DynamicDate != nil:
		c.FieldType = FieldTypeDate
	case c.Null:
		c.FieldType = FieldTypeNull
	default:
		c.FieldType = FieldTypeUnknown
	}

	return c.FieldType
}

// IsRowQuery checks whether a SOQL string uses SELECT ROW syntax.
// Uses simple prefix matching because the SOQL may contain :param placeholders
// that the parser does not handle.
func IsRowQuery(soql string) bool {
	s := strings.TrimSpace(soql)
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "SELECT") {
		return false
	}
	rest := strings.TrimSpace(s[len("SELECT"):])
	return strings.HasPrefix(strings.ToUpper(rest), "ROW")
}
