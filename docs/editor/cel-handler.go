// CEL Expression Editor — Backend Validation Handler (Go)
//
// HTTP endpoint for server-side CEL expression validation.
// Uses cel-go to compile expressions and return type information or errors.
//
// Endpoint: POST /api/v1/admin/cel/validate
//
// Dependencies:
//   go get github.com/google/cel-go
//   go get github.com/gin-gonic/gin

package cel

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CELHandler handles CEL expression validation requests.
type CELHandler struct {
	registry *FunctionRegistry
}

// NewCELHandler creates a new CELHandler.
func NewCELHandler(registry *FunctionRegistry) *CELHandler {
	return &CELHandler{registry: registry}
}

// RegisterRoutes registers CEL routes on the admin group.
func (h *CELHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.POST("/cel/validate", h.Validate)
}

// SetRegistry updates the function registry (called after function changes).
func (h *CELHandler) SetRegistry(registry *FunctionRegistry) {
	h.registry = registry
}

type celValidateRequest struct {
	Expression string     `json:"expression" binding:"required"`
	Context    string     `json:"context" binding:"required"`
	Params     []ParamDef `json:"params"`
}

type celValidateResponse struct {
	Valid      bool               `json:"valid"`
	ReturnType *string            `json:"return_type,omitempty"`
	Errors     []celValidateError `json:"errors,omitempty"`
}

type celValidateError struct {
	Message string `json:"message"`
	Line    *int   `json:"line,omitempty"`
	Column  *int   `json:"column,omitempty"`
}

// Validate compiles a CEL expression and returns validation result.
//
// Request body:
//
//	{
//	  "expression": "record.FirstName != ''",
//	  "context": "validation_rule",
//	  "params": [{"name": "x", "type": "number"}]
//	}
//
// Response (valid):
//
//	{"valid": true, "return_type": "bool"}
//
// Response (invalid):
//
//	{"valid": false, "errors": [{"message": "...", "line": 1, "column": 5}]}
func (h *CELHandler) Validate(c *gin.Context) {
	var req celValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	env, err := h.buildEnv(req.Context, req.Params)
	if err != nil {
		c.JSON(http.StatusOK, celValidateResponse{
			Valid:  false,
			Errors: []celValidateError{{Message: err.Error()}},
		})
		return
	}

	ast, issues := env.Compile(req.Expression)
	if issues != nil && issues.Err() != nil {
		errs := make([]celValidateError, 0, len(issues.Errors()))
		for _, e := range issues.Errors() {
			ve := celValidateError{Message: e.Message}
			if loc := e.Location; loc != nil {
				line := loc.Line()
				col := loc.Column() + 1 // 0-based → 1-based for API
				ve.Line = &line
				ve.Column = &col
			}
			errs = append(errs, ve)
		}
		if len(errs) == 0 {
			errs = []celValidateError{{Message: issues.Err().Error()}}
		}
		c.JSON(http.StatusOK, celValidateResponse{
			Valid:  false,
			Errors: errs,
		})
		return
	}

	returnType := ast.OutputType().String()
	c.JSON(http.StatusOK, celValidateResponse{
		Valid:      true,
		ReturnType: &returnType,
	})
}

func (h *CELHandler) buildEnv(context string, params []ParamDef) (*Env, error) {
	switch context {
	case "validation_rule", "when_expression":
		if h.registry != nil {
			return StandardEnvWithFunctions(h.registry)
		}
		return StandardEnv()
	case "default_expr":
		if h.registry != nil {
			return DefaultEnvWithFunctions(h.registry)
		}
		return DefaultEnv()
	case "function_body":
		return FunctionBodyEnv(params, h.registry)
	case "portal_when":
		return PortalEnv()
	case "gate_when":
		return GateEnv()
	default:
		if h.registry != nil {
			return StandardEnvWithFunctions(h.registry)
		}
		return StandardEnv()
	}
}
