package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/metadata"
)

// CELHandler handles CEL expression validation requests.
type CELHandler struct {
	cache    metadata.MetadataReader
	registry *celengine.FunctionRegistry
}

// NewCELHandler creates a new CELHandler.
func NewCELHandler(cache metadata.MetadataReader, registry *celengine.FunctionRegistry) *CELHandler {
	return &CELHandler{cache: cache, registry: registry}
}

// RegisterRoutes registers CEL routes on the admin group.
func (h *CELHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.POST("/cel/validate", h.Validate)
}

// SetRegistry updates the function registry (called after function changes).
func (h *CELHandler) SetRegistry(registry *celengine.FunctionRegistry) {
	h.registry = registry
}

type celValidateRequest struct {
	Expression    string               `json:"expression" binding:"required"`
	Context       string               `json:"context" binding:"required"`
	ObjectAPIName *string              `json:"object_api_name"`
	Params        []celengine.ParamDef `json:"params"`
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

func (h *CELHandler) buildEnv(context string, params []celengine.ParamDef) (*celengine.Env, error) {
	switch context {
	case "validation_rule", "when_expression":
		if h.registry != nil {
			return celengine.StandardEnvWithFunctions(h.registry)
		}
		return celengine.StandardEnv()
	case "default_expr":
		if h.registry != nil {
			return celengine.DefaultEnvWithFunctions(h.registry)
		}
		return celengine.DefaultEnv()
	case "function_body":
		return celengine.FunctionBodyEnv(params, h.registry)
	default:
		// Default to standard env
		if h.registry != nil {
			return celengine.StandardEnvWithFunctions(h.registry)
		}
		return celengine.StandardEnv()
	}
}
