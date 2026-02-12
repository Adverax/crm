package engine

import (
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Parser is the SOQL parser
var Parser = participle.MustBuild[Grammar](
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

// Parse parses a SOQL query string into an AST
func Parse(query string) (*Grammar, error) {
	return Parser.ParseString("", query)
}

// MustParse parses a SOQL query string into an AST, panics on error
func MustParse(query string) *Grammar {
	g, err := Parse(query)
	if err != nil {
		panic(err)
	}
	return g
}
