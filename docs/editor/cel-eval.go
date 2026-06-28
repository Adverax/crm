// CEL Expression Editor — ProgramCache and Evaluation Helpers (Go)
//
// Thread-safe compiled program cache with convenience evaluation methods.
// Compiles CEL expressions once, reuses the program for subsequent evaluations.
//
// Dependencies:
//   go get github.com/google/cel-go

package cel

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
)

// RecordMap is a type alias for record data used in CEL expressions.
type RecordMap = map[string]any

// ProgramCache caches compiled CEL programs for reuse.
// Thread-safe: multiple goroutines can call GetOrCompile and Evaluate* concurrently.
type ProgramCache struct {
	env      *cel.Env
	mu       sync.RWMutex
	programs map[string]cel.Program
}

// NewProgramCache creates a new ProgramCache with the given CEL environment.
func NewProgramCache(env *cel.Env) *ProgramCache {
	return &ProgramCache{
		env:      env,
		programs: make(map[string]cel.Program),
	}
}

// GetOrCompile returns a cached program or compiles and caches a new one.
func (c *ProgramCache) GetOrCompile(expr string) (cel.Program, error) {
	c.mu.RLock()
	prog, ok := c.programs[expr]
	c.mu.RUnlock()
	if ok {
		return prog, nil
	}

	ast, issues := c.env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, &CompileError{Expression: expr, Message: issues.Err().Error()}
	}

	prog, err := c.env.Program(ast)
	if err != nil {
		return nil, &CompileError{Expression: expr, Message: err.Error()}
	}

	c.mu.Lock()
	c.programs[expr] = prog
	c.mu.Unlock()

	return prog, nil
}

// EvaluateBool compiles (or retrieves from cache) and evaluates an expression,
// expecting a bool result. Returns an error if the result is not boolean.
func (c *ProgramCache) EvaluateBool(expr string, vars map[string]any) (bool, error) {
	prog, err := c.GetOrCompile(expr)
	if err != nil {
		return false, err
	}

	out, _, err := prog.Eval(vars)
	if err != nil {
		return false, &EvalError{Expression: expr, Cause: err}
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, &EvalError{
			Expression: expr,
			Cause:      fmt.Errorf("expected bool, got %T", out.Value()),
		}
	}

	return result, nil
}

// EvaluateAny compiles (or retrieves from cache) and evaluates an expression,
// returning any result.
func (c *ProgramCache) EvaluateAny(expr string, vars map[string]any) (any, error) {
	prog, err := c.GetOrCompile(expr)
	if err != nil {
		return nil, err
	}

	out, _, err := prog.Eval(vars)
	if err != nil {
		return nil, &EvalError{Expression: expr, Cause: err}
	}

	return out.Value(), nil
}

// Reset replaces the CEL environment and clears all cached programs.
// Used when custom functions change and environments need to be rebuilt.
func (c *ProgramCache) Reset(env *cel.Env) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.env = env
	c.programs = make(map[string]cel.Program)
}

// Env returns the underlying CEL environment.
func (c *ProgramCache) Env() *cel.Env {
	return c.env
}

// --- Variable builders ---

// UserVars builds a CEL variable map from user context fields.
func UserVars(userID, profileID, roleID string) map[string]any {
	return map[string]any{
		"id":         userID,
		"profile_id": profileID,
		"role_id":    roleID,
	}
}

// DefaultVars builds CEL variables for default expressions.
func DefaultVars(record RecordMap, user map[string]any) map[string]any {
	return map[string]any{
		"record": record,
		"user":   user,
		"now":    time.Now().UTC(),
	}
}

// ValidationVars builds CEL variables for validation rule expressions.
func ValidationVars(record, old RecordMap, user map[string]any) map[string]any {
	vars := map[string]any{
		"record": record,
		"user":   user,
		"now":    time.Now().UTC(),
	}
	if old != nil {
		vars["old"] = old
	}
	return vars
}
