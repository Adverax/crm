package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	celutil "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
)

var portalParamRegexp = regexp.MustCompile(`:(\w+)`)

// PortalHandler serves resolved Object View configs, per-query data, and action execution.
type PortalHandler struct {
	cache       metadata.MetadataReader
	soqlService soql.QueryService
	dmlService  dml.DMLService
	celCache    *celutil.ProgramCache
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(
	cache metadata.MetadataReader,
	soqlService soql.QueryService,
	dmlService dml.DMLService,
	celCache *celutil.ProgramCache,
) *PortalHandler {
	return &PortalHandler{
		cache:       cache,
		soqlService: soqlService,
		dmlService:  dmlService,
		celCache:    celCache,
	}
}

// RegisterRoutes registers the view routes on the given API group.
func (h *PortalHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/portal/:portalApiName", h.GetByAPIName)
	rg.GET("/portal/:portalApiName/query/:queryName", h.ExecuteQuery)
	rg.POST("/portal/:portalApiName/action/:actionKey", h.ExecuteAction)
}

// GetByAPIName returns the OV config by api_name.
func (h *PortalHandler) GetByAPIName(c *gin.Context) {
	apiName := c.Param("portalApiName")

	ov, ok := h.cache.GetPortalByAPIName(apiName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("portal", apiName))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

// ExecuteQuery executes a named query from an Object View.
func (h *PortalHandler) ExecuteQuery(c *gin.Context) {
	portalAPIName := c.Param("portalApiName")
	queryName := c.Param("queryName")

	ov, ok := h.cache.GetPortalByAPIName(portalAPIName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("portal", portalAPIName))
		return
	}

	// Find the query
	var query *metadata.PortalQuery
	for i := range ov.Config.Read.Queries {
		if ov.Config.Read.Queries[i].Name == queryName {
			query = &ov.Config.Read.Queries[i]
			break
		}
	}
	if query == nil {
		apperror.Respond(c, apperror.NotFound("query", queryName))
		return
	}

	var soqlText string
	if len(ov.Config.Args) > 0 {
		// Typed args path
		argsMap, err := resolveArgs(ov.Config.Args, c)
		if err != nil {
			apperror.Respond(c, err)
			return
		}

		// Validate args
		if h.celCache != nil {
			if err := validateArgsRuntime(ov.Config.Args, argsMap, h.celCache); err != nil {
				apperror.Respond(c, err)
				return
			}
		}

		// Evaluate when condition
		if query.When != "" && h.celCache != nil {
			result, evalErr := h.celCache.EvaluateBool(query.When, map[string]any{"args": argsMap})
			if evalErr != nil {
				apperror.Respond(c, fmt.Errorf("portalHandler.ExecuteQuery: when eval: %w", evalErr))
				return
			}
			if !result {
				c.JSON(http.StatusOK, gin.H{"data": nil})
				return
			}
		}

		soqlText = substituteTypedArgs(query.SOQL, argsMap)
	} else {
		// Backward compat: untyped substitution
		soqlText = substituteParams(query.SOQL, c)
	}

	// Parse pagination
	perPage := 20
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 200 {
			perPage = v
		}
	}

	result, err := h.soqlService.Execute(c.Request.Context(), soqlText, &soql.QueryParams{
		PageSize: perPage,
	})
	if err != nil {
		apperror.Respond(c, fmt.Errorf("portalHandler.ExecuteQuery: %w", err))
		return
	}

	// For SELECT ROW queries, return single record instead of array.
	if result.IsRow {
		var record map[string]any
		if len(result.Records) > 0 {
			record = result.Records[0]
		}
		c.JSON(http.StatusOK, gin.H{"data": record})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

type executeActionRequest struct {
	Data     map[string]any `json:"data"`
	RecordID string         `json:"record_id"`
}

type actionResultItem struct {
	Operation string   `json:"operation"`
	Object    string   `json:"object"`
	IDs       []string `json:"ids,omitempty"`
}

// ExecuteAction executes a named action from an Object View (ADR-0036).
func (h *PortalHandler) ExecuteAction(c *gin.Context) {
	portalAPIName := c.Param("portalApiName")
	actionKey := c.Param("actionKey")

	ov, ok := h.cache.GetPortalByAPIName(portalAPIName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("portal", portalAPIName))
		return
	}

	// Find the action
	var action *metadata.PortalAction
	for i := range ov.Config.Read.Actions {
		if ov.Config.Read.Actions[i].Key == actionKey {
			action = &ov.Config.Read.Actions[i]
			break
		}
	}
	if action == nil {
		apperror.Respond(c, apperror.NotFound("action", actionKey))
		return
	}

	if action.Apply == nil {
		apperror.Respond(c, apperror.BadRequest("action is not executable"))
		return
	}

	var req executeActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	if action.Apply.Type == "scenario" {
		apperror.Respond(c, apperror.BadRequest("scenario actions are not yet implemented"))
		return
	}

	// Execute DML batch
	results, err := h.dmlService.ExecuteBatch(c.Request.Context(), action.Apply.DML)
	if err != nil {
		apperror.Respond(c, fmt.Errorf("portalHandler.ExecuteAction: %w", err))
		return
	}

	items := make([]actionResultItem, len(results))
	for i, r := range results {
		item := actionResultItem{}
		if len(r.InsertedIds) > 0 {
			item.Operation = "insert"
			item.IDs = r.InsertedIds
		} else if len(r.UpdatedIds) > 0 {
			item.Operation = "update"
			item.IDs = r.UpdatedIds
		} else if len(r.DeletedIds) > 0 {
			item.Operation = "delete"
			item.IDs = r.DeletedIds
		}
		items[i] = item
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": items,
	})
}

// validateArgsRuntime evaluates CEL validation expressions for each arg that has one.
func validateArgsRuntime(argDefs []metadata.PortalArg, argsMap map[string]any, celCache *celutil.ProgramCache) error {
	vars := map[string]any{"args": argsMap}
	for _, arg := range argDefs {
		if arg.Validation == "" {
			continue
		}
		result, err := celCache.EvaluateBool(arg.Validation, vars)
		if err != nil {
			return apperror.BadRequest(fmt.Sprintf("arg %q: validation error: %v", arg.Name, err))
		}
		if !result {
			msg := arg.ErrorMessage
			if msg == "" {
				msg = fmt.Sprintf("arg %q failed validation", arg.Name)
			}
			return apperror.BadRequest(msg)
		}
	}
	return nil
}

// resolveArgs parses URL query params against declared arg definitions.
// Returns a typed map for CEL and SOQL substitution.
func resolveArgs(argDefs []metadata.PortalArg, c *gin.Context) (map[string]any, error) {
	result := make(map[string]any, len(argDefs))

	for _, arg := range argDefs {
		raw := c.Query(arg.Name)

		if raw == "" {
			if arg.Default != nil {
				raw = *arg.Default
			} else {
				return nil, apperror.BadRequest(fmt.Sprintf("required arg %q is missing", arg.Name))
			}
		}

		val, err := convertArg(arg.Name, arg.Type, raw)
		if err != nil {
			return nil, err
		}
		result[arg.Name] = val
	}

	return result, nil
}

// convertArg converts a raw string value to the declared arg type.
func convertArg(name, argType, raw string) (any, error) {
	switch argType {
	case "string":
		return raw, nil
	case "int":
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, apperror.BadRequest(fmt.Sprintf("arg %q: %q is not a valid int", name, raw))
		}
		return v, nil
	case "float":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, apperror.BadRequest(fmt.Sprintf("arg %q: %q is not a valid float", name, raw))
		}
		return v, nil
	case "bool":
		switch raw {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, apperror.BadRequest(fmt.Sprintf("arg %q: %q is not a valid bool", name, raw))
		}
	default:
		return nil, apperror.BadRequest(fmt.Sprintf("arg %q: unknown type %q", name, argType))
	}
}

// substituteTypedArgs replaces :paramName in SOQL with typed values.
// Strings are single-quoted (with escaping), numbers and bools are unquoted.
func substituteTypedArgs(soqlText string, args map[string]any) string {
	return portalParamRegexp.ReplaceAllStringFunc(soqlText, func(match string) string {
		paramName := match[1:]
		val, ok := args[paramName]
		if !ok {
			return match
		}

		switch v := val.(type) {
		case string:
			escaped := strings.ReplaceAll(v, "'", "''")
			return "'" + escaped + "'"
		case int64:
			return strconv.FormatInt(v, 10)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(v)
		default:
			return match
		}
	})
}

// substituteParams replaces :paramName in SOQL with URL query parameter values.
// Legacy path for portals without declared args.
func substituteParams(soqlText string, c *gin.Context) string {
	return portalParamRegexp.ReplaceAllStringFunc(soqlText, func(match string) string {
		paramName := match[1:] // strip leading ':'
		if val := c.Query(paramName); val != "" {
			return "'" + val + "'"
		}
		return match
	})
}
