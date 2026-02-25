package procedure

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"

	celengine "github.com/adverax/crm/internal/platform/cel"
)

// NewProcedureCELEnv creates a CEL environment suitable for procedure expressions.
// It includes: input (map), user (map), now (timestamp), and step results (dyn).
// Plus fn.* custom functions from the registry.
func NewProcedureCELEnv(registry *celengine.FunctionRegistry) (*cel.Env, error) {
	opts := []cel.EnvOption{
		cel.Variable("input", cel.DynType),
		cel.Variable("user", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		cel.Variable("error", cel.DynType),
		ext.Strings(),
	}

	if registry != nil {
		opts = append(opts, registry.EnvOptions()...)
	}

	return cel.NewEnv(opts...)
}
