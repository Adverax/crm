package engine

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// Lexer defines tokens for SOQL
var Lexer = lexer.MustSimple([]lexer.SimpleRule{
	// Keywords (must come before Ident to take precedence)
	{Name: "Keyword", Pattern: `\b(?i:SELECT|FROM|WHERE|WITH|SECURITY_ENFORCED|GROUP|BY|HAVING|ORDER|LIMIT|OFFSET|AND|OR|NOT|IN|LIKE|IS|NULL|TRUE|FALSE|AS|ASC|DESC|NULLS|FIRST|LAST|FOR|UPDATE|TYPEOF|WHEN|THEN|ELSE|END|COUNT|COUNT_DISTINCT|SUM|AVG|MIN|MAX|COALESCE|NULLIF|CONCAT|UPPER|LOWER|TRIM|LENGTH|LEN|SUBSTRING|SUBSTR|ABS|ROUND|FLOOR|CEIL|CEILING)\b`},

	// Static date literals (including fiscal periods)
	{Name: "StaticDateLiteral", Pattern: `\b(?i:TODAY|YESTERDAY|TOMORROW|THIS_WEEK|LAST_WEEK|NEXT_WEEK|THIS_MONTH|LAST_MONTH|NEXT_MONTH|THIS_QUARTER|LAST_QUARTER|NEXT_QUARTER|THIS_YEAR|LAST_YEAR|NEXT_YEAR|LAST_90_DAYS|NEXT_90_DAYS|THIS_FISCAL_QUARTER|LAST_FISCAL_QUARTER|NEXT_FISCAL_QUARTER|THIS_FISCAL_YEAR|LAST_FISCAL_YEAR|NEXT_FISCAL_YEAR)\b`},

	// Dynamic date literals: LAST_N_DAYS:30, NEXT_N_MONTHS:3, LAST_N_FISCAL_QUARTERS:2, etc.
	{Name: "DynamicDateLiteral", Pattern: `\b(?i:LAST_N_DAYS|NEXT_N_DAYS|LAST_N_WEEKS|NEXT_N_WEEKS|LAST_N_MONTHS|NEXT_N_MONTHS|LAST_N_QUARTERS|NEXT_N_QUARTERS|LAST_N_YEARS|NEXT_N_YEARS|LAST_N_FISCAL_QUARTERS|NEXT_N_FISCAL_QUARTERS|LAST_N_FISCAL_YEARS|NEXT_N_FISCAL_YEARS):\d+\b`},

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

	// Operators (|| must come before single |)
	{Name: "Operators", Pattern: `<>|!=|<=|>=|==|\|\||[+\-*/%=<>(),.]`},

	// Whitespace (ignored)
	{Name: "Whitespace", Pattern: `[\s\n\r\t]+`},

	// Comments (SQL style)
	{Name: "Comment", Pattern: `--[^\n]*`},
})
