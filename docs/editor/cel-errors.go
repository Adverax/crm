// CEL Expression Editor — Typed Error Types (Go)
//
// Structured error types for CEL compilation and evaluation errors.

package cel

import "fmt"

// CompileError represents a CEL expression compilation error.
type CompileError struct {
	Expression string
	Message    string
}

func (e *CompileError) Error() string {
	return fmt.Sprintf("CEL compile error for expression %q: %s", e.Expression, e.Message)
}

// EvalError represents a CEL expression evaluation error.
type EvalError struct {
	Expression string
	Cause      error
}

func (e *EvalError) Error() string {
	return fmt.Sprintf("CEL eval error for expression %q: %s", e.Expression, e.Cause)
}

func (e *EvalError) Unwrap() error {
	return e.Cause
}
