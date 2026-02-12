package engine

import (
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Parser is the DML parser
var Parser = participle.MustBuild[DMLStatement](
	participle.Lexer(Lexer),
	participle.CaseInsensitive("Keyword"),
	participle.Elide("Whitespace", "Comment"),

	// Unquote string literals (remove surrounding quotes and handle escapes)
	participle.Map(func(token lexer.Token) (lexer.Token, error) {
		// Remove surrounding single quotes and unescape ''
		s := token.Value[1 : len(token.Value)-1]
		s = strings.ReplaceAll(s, "''", "'")
		token.Value = s
		return token, nil
	}, "String"),

	// Handle quoted identifiers
	participle.Map(func(token lexer.Token) (lexer.Token, error) {
		// Unquote the identifier
		value, err := strconv.Unquote(token.Value)
		if err != nil {
			return token, participle.Errorf(token.Pos, "invalid quoted identifier %q: %s", token.Value, err.Error())
		}
		token.Type = Lexer.Symbols()["Ident"]
		token.Value = value
		return token, nil
	}, "QuotedIdent"),
)

// Parse parses a DML statement string into an AST.
func Parse(statement string) (*DMLStatement, error) {
	return Parser.ParseString("", statement)
}

// MustParse parses a DML statement string into an AST, panics on error.
func MustParse(statement string) *DMLStatement {
	s, err := Parse(statement)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseInsert parses an INSERT statement.
func ParseInsert(statement string) (*InsertStatement, error) {
	ast, err := Parse(statement)
	if err != nil {
		return nil, err
	}
	if ast.Insert == nil {
		return nil, NewValidationError(ErrCodeInvalidExpression, "not an INSERT statement")
	}
	return ast.Insert, nil
}

// ParseUpdate parses an UPDATE statement.
func ParseUpdate(statement string) (*UpdateStatement, error) {
	ast, err := Parse(statement)
	if err != nil {
		return nil, err
	}
	if ast.Update == nil {
		return nil, NewValidationError(ErrCodeInvalidExpression, "not an UPDATE statement")
	}
	return ast.Update, nil
}

// ParseDelete parses a DELETE statement.
func ParseDelete(statement string) (*DeleteStatement, error) {
	ast, err := Parse(statement)
	if err != nil {
		return nil, err
	}
	if ast.Delete == nil {
		return nil, NewValidationError(ErrCodeInvalidExpression, "not a DELETE statement")
	}
	return ast.Delete, nil
}

// ParseUpsert parses an UPSERT statement.
func ParseUpsert(statement string) (*UpsertStatement, error) {
	ast, err := Parse(statement)
	if err != nil {
		return nil, err
	}
	if ast.Upsert == nil {
		return nil, NewValidationError(ErrCodeInvalidExpression, "not an UPSERT statement")
	}
	return ast.Upsert, nil
}
