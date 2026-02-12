package engine

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// Lexer defines tokens for DML statements
var Lexer = lexer.MustSimple([]lexer.SimpleRule{
	// DML Keywords and Functions (must come before Ident to take precedence)
	{Name: "Keyword", Pattern: `\b(?i:INSERT|INTO|VALUES|UPDATE|SET|DELETE|FROM|WHERE|UPSERT|ON|AND|OR|NOT|IN|LIKE|IS|NULL|TRUE|FALSE|COALESCE|NULLIF|CONCAT|UPPER|LOWER|TRIM|LENGTH|LEN|SUBSTRING|SUBSTR|ABS|ROUND|FLOOR|CEIL|CEILING)\b`},

	// DateTime: 2024-01-15T10:30:00Z (must come before Date)
	{Name: "DateTime", Pattern: `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})`},

	// Date: 2024-01-15
	{Name: "Date", Pattern: `\d{4}-\d{2}-\d{2}`},

	// Identifiers
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},

	// Quoted identifiers: "Field Name"
	{Name: "QuotedIdent", Pattern: `"(?:\\.|[^"])*"`},

	// Float numbers (must come before Integer)
	{Name: "Float", Pattern: `[-+]?\d+\.\d+(?:[eE][-+]?\d+)?`},

	// Integer numbers
	{Name: "Integer", Pattern: `\d+`},

	// String literals (single quotes)
	{Name: "String", Pattern: `'(?:''|[^'])*'`},

	// Operators
	{Name: "Operators", Pattern: `<>|!=|<=|>=|==|[+\-*/%=<>(),.]`},

	// Whitespace (ignored)
	{Name: "Whitespace", Pattern: `[\s\n\r\t]+`},

	// Comments (SQL style)
	{Name: "Comment", Pattern: `--[^\n]*`},
})
