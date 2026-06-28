// CEL Expression Editor — Environment Builders (Go)
//
// Creates CEL environments per evaluation context.
// Each context determines which variables are available.
//
// Dependencies:
//   go get github.com/google/cel-go

package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

// Env is a type alias for cel.Env to simplify imports.
type Env = cel.Env

// StandardEnv creates a CEL environment for validation rules and when expressions.
// Variables: record (dyn), old (dyn), user (dyn), now (timestamp).
func StandardEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("record", cel.DynType),
		cel.Variable("old", cel.DynType),
		cel.Variable("user", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		ext.Strings(),
	)
}

// DefaultEnv creates a CEL environment for default expressions.
// Variables: record (dyn), user (dyn), now (timestamp).
func DefaultEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("record", cel.DynType),
		cel.Variable("user", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		ext.Strings(),
	)
}

// PortalEnv creates a CEL environment for portal-level arg validation.
// Variables: args (dyn).
func PortalEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("args", cel.DynType),
		ext.Strings(),
	)
}

// GateEnv creates a CEL environment for gate-level expressions.
// Variables: args (dyn), datasets (dyn), data (dyn), user (dyn).
func GateEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("args", cel.DynType),
		cel.Variable("datasets", cel.DynType),
		cel.Variable("data", cel.DynType),
		cel.Variable("user", cel.DynType),
		ext.Strings(),
	)
}
