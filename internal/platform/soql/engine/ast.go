package engine

import "github.com/alecthomas/participle/v2/lexer"

// Grammar is the root AST node representing a SOQL query
type Grammar struct {
	Pos                  lexer.Position
	Select               []*SelectExpression `"SELECT" @@ ("," @@)*`
	From                 string              `"FROM" @Ident`
	Where                *Expression         `("WHERE" @@)?`
	WithSecurityEnforced bool                `@("WITH" "SECURITY_ENFORCED")?`
	GroupBy              []*GroupClause      `("GROUP" "BY" @@ ("," @@)*)?`
	Having               *Expression         `("HAVING" @@)?`
	OrderBy              []*OrderClause      `("ORDER" "BY" @@ ("," @@)*)?`
	Limit                *int                `("LIMIT" @Integer)?`
	Offset               *int                `("OFFSET" @Integer)?`
	ForUpdate            bool                `@("FOR" "UPDATE")?`
}

// SelectExpression represents a single item in SELECT clause
type SelectExpression struct {
	Pos   lexer.Position
	Item  *SelectItem `@@`
	Alias *string     `(("AS" @Ident) | (?! "FROM" | "," | "WHERE" | "GROUP" | "ORDER" | "LIMIT" | "OFFSET" | "HAVING") @Ident)?`
}

// SelectItem represents the actual content of a SELECT item
type SelectItem struct {
	Pos       lexer.Position
	Typeof    *TypeofExpression     `  @@`
	Aggregate *AggregateExpression  `| @@`
	Subquery  *RelationshipSubquery `| @@`
	Expr      *Expression           `| @@`
}

// TypeofExpression represents a TYPEOF expression for polymorphic fields
// Example: TYPEOF What WHEN Account THEN Name, Industry WHEN Opportunity THEN Name, Amount ELSE Name END
type TypeofExpression struct {
	Pos         lexer.Position
	Field       string        `"TYPEOF" @Ident`
	WhenClauses []*WhenClause `@@+`
	ElseFields  []string      `("ELSE" @Ident ("," @Ident)*)?`
	End         string        `"END"`
}

// WhenClause represents a single WHEN clause in TYPEOF
// Example: WHEN Account THEN Name, Industry
type WhenClause struct {
	Pos        lexer.Position
	ObjectType string   `"WHEN" @Ident`
	Fields     []string `"THEN" @Ident ("," @Ident)*`
}

// AggregateExpression represents an aggregate function call
type AggregateExpression struct {
	Pos        lexer.Position
	Function   Aggregate   `@("COUNT" | "COUNT_DISTINCT" | "SUM" | "AVG" | "MIN" | "MAX")`
	OpenParen  string      `"("`
	Expression *Expression `@@`
	CloseParen string      `")"`
	FieldType  FieldType   // inferred type
}

// RelationshipSubquery represents a Parent-to-Child subquery in SELECT
// Example: (SELECT FirstName, Email FROM Contacts)
type RelationshipSubquery struct {
	Pos        lexer.Position
	OpenParen  string              `"("`
	Select     []*SelectExpression `"SELECT" @@ ("," @@)*`
	From       string              `"FROM" @Ident`
	Where      *Expression         `("WHERE" @@)?`
	OrderBy    []*OrderClause      `("ORDER" "BY" @@ ("," @@)*)?`
	Limit      *int                `("LIMIT" @Integer)?`
	CloseParen string              `")"`
}

// WhereSubquery represents a subquery in WHERE IN clause (semi-join)
// Example: SELECT Name FROM Account WHERE Id IN (SELECT AccountId FROM Contact)
type WhereSubquery struct {
	Pos        lexer.Position
	OpenParen  string      `"("`
	Select     *Expression `"SELECT" @@` // Single field only
	From       string      `"FROM" @Ident`
	Where      *Expression `("WHERE" @@)?`
	Limit      *int        `("LIMIT" @Integer)?`
	CloseParen string      `")"`
}

// OrderClause represents a single ORDER BY item
type OrderClause struct {
	OrderItem *OrderItem  `@@`
	Direction *Direction  `@("ASC" | "DESC")?`
	Nulls     *NullsOrder `("NULLS" @("FIRST" | "LAST"))?`
}

// OrderItem represents what to order by (field or aggregate)
type OrderItem struct {
	Pos       lexer.Position
	Aggregate *AggregateExpression `  @@`
	Field     []string             `| @Ident ("." @Ident)*`
}

// GroupClause represents a single GROUP BY item
type GroupClause struct {
	Pos   lexer.Position
	Field []string `@Ident ("." @Ident)*`
}

// Expression is the top-level expression node
type Expression struct {
	Pos       lexer.Position
	Or        *OrExpr `@@`
	FieldType FieldType
}

// OrExpr represents OR expressions
type OrExpr struct {
	And       []*AndExpr `@@ ("OR" @@)*`
	FieldType FieldType
}

// AndExpr represents AND expressions
type AndExpr struct {
	Not       []*NotExpr `@@ ("AND" @@)*`
	FieldType FieldType
}

// NotExpr represents NOT expressions
type NotExpr struct {
	Not       bool         `@"NOT"?`
	Compare   *CompareExpr `@@`
	FieldType FieldType
}

// CompareExpr represents comparison expressions
type CompareExpr struct {
	Left      *InExpr   `@@`
	Operator  *Operator `(@("=" | "==" | "!=" | "<>" | ">" | "<" | ">=" | "<=")`
	Right     *InExpr   `@@)?`
	FieldType FieldType
}

// InExpr represents IN/NOT IN expressions
// Supports both literal values and subqueries:
//   - Id IN ('001', '002', '003')
//   - Id IN (SELECT AccountId FROM Contact WHERE Status = 'Active')
type InExpr struct {
	Left      *LikeExpr      `@@`
	Not       bool           `(@"NOT"?`
	In        bool           `@"IN"`
	Subquery  *WhereSubquery `( @@`
	Values    []*Value       `| "(" @@ ("," @@)* ")" ))?`
	FieldType FieldType
}

// LikeExpr represents LIKE expressions
type LikeExpr struct {
	Left      *IsExpr `@@`
	Not       bool    `(@"NOT"?`
	Like      bool    `@"LIKE"`
	Pattern   *Value  `@@)?`
	FieldType FieldType
}

// IsExpr represents IS NULL / IS NOT NULL expressions
type IsExpr struct {
	Left      *AddExpr `@@`
	Is        bool     `(@"IS"`
	Not       bool     `@"NOT"?`
	Null      bool     `@"NULL")?`
	FieldType FieldType
}

// AddExpr represents addition/subtraction expressions
type AddExpr struct {
	Left      *MulExpr `@@`
	Right     []*AddOp `@@*`
	FieldType FieldType
}

// AddOp represents a single addition/subtraction/concatenation operation
type AddOp struct {
	Operator  Operator `@("+" | "-" | "||")`
	Right     *MulExpr `@@`
	FieldType FieldType
}

// MulExpr represents multiplication/division expressions
type MulExpr struct {
	Left      *UnaryExpr `@@`
	Right     []*MulOp   `@@*`
	FieldType FieldType
}

// MulOp represents a single multiplication/division operation
type MulOp struct {
	Operator  Operator   `@("*" | "/" | "%")`
	Right     *UnaryExpr `@@`
	FieldType FieldType
}

// UnaryExpr represents unary expressions (+/-)
type UnaryExpr struct {
	Operator  *Operator `@("+" | "-")?`
	Primary   *Primary  `@@`
	FieldType FieldType
}

// Primary represents primary expressions (literals, fields, function calls, etc.)
type Primary struct {
	Subexpression *Expression          `  "(" @@ ")"`
	Aggregate     *AggregateExpression `| @@`
	FuncCall      *FuncCall            `| @@`
	Const         *Const               `| @@`
	Field         *Field               `| @@`
	FieldType     FieldType
}

// Value represents a value in IN clause or LIKE pattern
type Value struct {
	Const     *Const `  @@`
	Field     *Field `| @@`
	FieldType FieldType
}

// Field represents a field reference (possibly with dot notation)
type Field struct {
	Pos       lexer.Position
	Path      []string `@Ident ("." @Ident)*`
	FieldType FieldType
}

// FuncCall represents a function call
type FuncCall struct {
	Name      Function      `@("COALESCE" | "NULLIF" | "CONCAT" | "UPPER" | "LOWER" | "TRIM" | "LENGTH" | "LEN" | "SUBSTRING" | "SUBSTR" | "ABS" | "ROUND" | "FLOOR" | "CEIL" | "CEILING")`
	Args      []*Expression `"(" @@ ("," @@)* ")"`
	FieldType FieldType
}

// Const represents a constant value
type Const struct {
	DynamicDate *DynamicDateLiteral `  @DynamicDateLiteral`
	StaticDate  *StaticDateLiteral  `| @StaticDateLiteral`
	DateTime    *DateTime           `| @DateTime`
	Date        *Date               `| @Date`
	String      *string             `| @String`
	Float       *float64            `| @Float`
	Integer     *int                `| @Integer`
	Boolean     *Boolean            `| @("TRUE" | "FALSE")`
	Null        bool                `| @"NULL"`
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
