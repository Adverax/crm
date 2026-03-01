package cel

import (
	"sort"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/ext"
)

var (
	parseEnv     *cel.Env
	parseEnvOnce sync.Once
)

func getParseEnv() *cel.Env {
	parseEnvOnce.Do(func() {
		env, err := cel.NewEnv(
			cel.Variable("record", cel.DynType),
			cel.Variable("old", cel.DynType),
			cel.Variable("user", cel.DynType),
			cel.Variable("now", cel.TimestampType),
			ext.Strings(),
		)
		if err != nil {
			return
		}
		parseEnv = env
	})
	return parseEnv
}

// ExtractRecordFieldRefs parses a CEL expression and extracts field names
// referenced as record.field_name. Returns nil on parse error.
func ExtractRecordFieldRefs(expression string) []string {
	env := getParseEnv()
	if env == nil {
		return nil
	}

	parsed, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return nil
	}

	native := parsed.NativeRep()
	if native == nil {
		return nil
	}

	nav := ast.NavigateAST(native)

	seen := make(map[string]bool)
	visitor := ast.NewExprVisitor(func(e ast.Expr) {
		if e.Kind() != ast.SelectKind {
			return
		}
		sel := e.AsSelect()
		operand := sel.Operand()
		if operand.Kind() == ast.IdentKind && operand.AsIdent() == "record" {
			seen[sel.FieldName()] = true
		}
	})
	ast.PreOrderVisit(nav, visitor)

	if len(seen) == 0 {
		return nil
	}

	fields := make([]string, 0, len(seen))
	for f := range seen {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	return fields
}
