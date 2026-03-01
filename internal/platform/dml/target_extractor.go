package dml

import (
	"github.com/adverax/crm/internal/platform/dml/engine"
)

// DMLTargetInfo describes the target object and modified fields of a DML statement.
type DMLTargetInfo struct {
	Object    string   // target object API name
	Fields    []string // modified fields (INSERT columns, UPDATE SET fields, UPSERT columns)
	Operation string   // "insert", "update", "delete", "upsert"
}

// ExtractTargets parses DML statements and extracts target objects with modified fields.
// Invalid statements are silently skipped.
func ExtractTargets(statements []string) []DMLTargetInfo {
	var targets []DMLTargetInfo
	for _, stmt := range statements {
		ast, err := engine.Parse(stmt)
		if err != nil {
			continue
		}

		info := DMLTargetInfo{
			Object:    ast.GetObject(),
			Operation: operationString(ast.GetOperation()),
			Fields:    extractFields(ast),
		}
		targets = append(targets, info)
	}
	return targets
}

func operationString(op engine.Operation) string {
	switch op {
	case engine.OperationInsert:
		return "insert"
	case engine.OperationUpdate:
		return "update"
	case engine.OperationDelete:
		return "delete"
	case engine.OperationUpsert:
		return "upsert"
	default:
		return "unknown"
	}
}

func extractFields(ast *engine.DMLStatement) []string {
	switch {
	case ast.Insert != nil:
		return ast.Insert.Fields
	case ast.Update != nil:
		fields := make([]string, len(ast.Update.Assignments))
		for i, a := range ast.Update.Assignments {
			fields[i] = a.Field
		}
		return fields
	case ast.Upsert != nil:
		return ast.Upsert.Fields
	case ast.Delete != nil:
		return nil
	default:
		return nil
	}
}
