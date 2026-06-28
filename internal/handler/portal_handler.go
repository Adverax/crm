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

// PortalHandler serves resolved Portal configs and gate execution (ADR-0037).
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

// RegisterRoutes registers the portal routes on the given API group.
func (h *PortalHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/portal/:portalApiName", h.GetByAPIName)
	rg.GET("/portal/:portalApiName/gate/:gateName", h.ExecuteGate)
	rg.POST("/portal/:portalApiName/gate/:gateName", h.ExecuteGate)
}

// GetByAPIName returns the portal config by api_name.
func (h *PortalHandler) GetByAPIName(c *gin.Context) {
	apiName := c.Param("portalApiName")

	ov, ok := h.cache.GetPortalByAPIName(apiName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("portal", apiName))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

type gatePostBody struct {
	Args map[string]any `json:"args"`
	Data map[string]any `json:"data"`
}

// ExecuteGate executes a named gate from a Portal (ADR-0037).
// GET: executes SOQL body steps only, returns layout + datasets + outcomes.
// POST: executes all body steps in order, returns datasets + outcomes.
func (h *PortalHandler) ExecuteGate(c *gin.Context) {
	portalAPIName := c.Param("portalApiName")
	gateName := c.Param("gateName")
	isPost := c.Request.Method == http.MethodPost

	portal, ok := h.cache.GetPortalByAPIName(portalAPIName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("portal", portalAPIName))
		return
	}

	gate, ok := portal.Config.Gates[gateName]
	if !ok {
		apperror.Respond(c, apperror.NotFound("gate", gateName))
		return
	}

	// Check HTTP method compatibility
	hasDML := gateHasDML(gate)
	hasSOQL := gateHasSOQL(gate)
	hasLayout := gate.Layout != nil

	if !isPost && !hasSOQL && hasDML {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "this gate only supports POST"})
		return
	}
	if isPost && !hasDML {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "this gate has no DML steps, use GET"})
		return
	}

	// Merge args (portal-level + gate-level)
	mergedArgDefs := mergeArgDefs(portal.Config.Args, gate.Args)

	// Resolve args
	var argsMap map[string]any
	var postData map[string]any

	if isPost {
		var body gatePostBody
		if err := c.ShouldBindJSON(&body); err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid request body"))
			return
		}
		var err error
		argsMap, err = resolveArgsFromMap(mergedArgDefs, body.Args)
		if err != nil {
			apperror.Respond(c, err)
			return
		}
		postData = body.Data
	} else {
		var err error
		argsMap, err = resolveArgs(mergedArgDefs, c)
		if err != nil {
			apperror.Respond(c, err)
			return
		}
	}

	// Validate arg rules runtime (portal-level, then gate-level)
	if h.celCache != nil {
		if err := validateArgRulesRuntime(portal.Config.ArgRules, argsMap, h.celCache); err != nil {
			apperror.Respond(c, err)
			return
		}
		if err := validateArgRulesRuntime(gate.ArgRules, argsMap, h.celCache); err != nil {
			apperror.Respond(c, err)
			return
		}
	}

	// Execute body steps
	datasets := make(map[string]any)

	// Merge args + data for substitution
	subMap := make(map[string]any, len(argsMap)+len(postData))
	for k, v := range argsMap {
		subMap[k] = v
	}
	for k, v := range postData {
		subMap[k] = v
	}

	for _, step := range gate.Body {
		// Evaluate when condition
		if step.When != "" && h.celCache != nil {
			evalVars := map[string]any{"args": argsMap, "datasets": datasets, "data": postData}
			result, evalErr := h.celCache.EvaluateBool(step.When, evalVars)
			if evalErr != nil {
				apperror.Respond(c, fmt.Errorf("portalHandler.ExecuteGate: step %q when eval: %w", step.Name, evalErr))
				return
			}
			if !result {
				datasets[step.Name] = nil
				continue
			}
		}

		switch step.Type {
		case "soql":
			soqlText := substituteTypedArgs(step.SOQL, subMap)

			perPage := step.PageSize
			if perPage == 0 {
				perPage = 200
			}
			if pp := c.Query("per_page"); pp != "" {
				if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 200 {
					perPage = v
				}
			}

			result, err := h.soqlService.Execute(c.Request.Context(), soqlText, &soql.QueryParams{
				PageSize: perPage,
			})
			if err != nil {
				apperror.Respond(c, fmt.Errorf("portalHandler.ExecuteGate: step %q: %w", step.Name, err))
				return
			}

			if result.IsRow {
				var record map[string]any
				if len(result.Records) > 0 {
					record = result.Records[0]
				}
				datasets[step.Name] = record
			} else {
				datasets[step.Name] = result
			}

		case "dml":
			if !isPost {
				// Skip DML steps on GET
				continue
			}

			dmlText := substituteTypedArgs(step.DML, subMap)
			results, err := h.dmlService.ExecuteBatch(c.Request.Context(), []string{dmlText})
			if err != nil {
				// Return error with layout for form re-render
				resp := gin.H{
					"gate":     gateName,
					"datasets": datasets,
					"outcomes": buildOutcomeResponse(gate.Outcomes),
					"errors":   []gin.H{{"message": err.Error()}},
				}
				if hasLayout {
					resp["layout"] = gate.Layout
				}
				c.JSON(http.StatusUnprocessableEntity, gin.H{"data": resp})
				return
			}

			if len(results) > 0 {
				item := make(map[string]any)
				r := results[0]
				if len(r.InsertedIds) > 0 {
					item["operation"] = "insert"
					item["ids"] = r.InsertedIds
					if len(r.InsertedIds) == 1 {
						item["Id"] = r.InsertedIds[0]
					}
				} else if len(r.UpdatedIds) > 0 {
					item["operation"] = "update"
					item["ids"] = r.UpdatedIds
				} else if len(r.DeletedIds) > 0 {
					item["operation"] = "delete"
					item["ids"] = r.DeletedIds
				}
				datasets[step.Name] = item
			}
		}
	}

	// Build response
	resp := gin.H{
		"gate":     gateName,
		"datasets": datasets,
		"outcomes": buildOutcomeResponse(gate.Outcomes),
		"errors":   nil,
	}

	if hasLayout {
		resp["layout"] = gate.Layout
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// buildOutcomeResponse converts GateOutcome slice to response format.
func buildOutcomeResponse(outcomes []metadata.GateOutcome) []gin.H {
	if len(outcomes) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, len(outcomes))
	for i, o := range outcomes {
		item := gin.H{
			"name": o.Name,
			"gate": o.Gate,
		}
		if o.Label != "" {
			item["label"] = o.Label
		}
		if o.Icon != "" {
			item["icon"] = o.Icon
		}
		if o.Type != "" {
			item["type"] = o.Type
		}
		if len(o.ArgsTemplate) > 0 {
			item["args_template"] = o.ArgsTemplate
		}
		result[i] = item
	}
	return result
}

// gateHasDML returns true if the gate has any DML body steps.
func gateHasDML(gate metadata.PortalGate) bool {
	for _, step := range gate.Body {
		if step.Type == "dml" {
			return true
		}
	}
	return false
}

// gateHasSOQL returns true if the gate has any SOQL body steps.
func gateHasSOQL(gate metadata.PortalGate) bool {
	for _, step := range gate.Body {
		if step.Type == "soql" {
			return true
		}
	}
	return false
}

// mergeArgDefs merges portal-level and gate-level arg definitions.
// Gate args shadow portal args on name collision.
func mergeArgDefs(portalArgs, gateArgs []metadata.PortalArg) []metadata.PortalArg {
	if len(gateArgs) == 0 {
		return portalArgs
	}
	if len(portalArgs) == 0 {
		return gateArgs
	}

	gateNames := make(map[string]bool, len(gateArgs))
	for _, a := range gateArgs {
		gateNames[a.Name] = true
	}

	merged := make([]metadata.PortalArg, 0, len(portalArgs)+len(gateArgs))
	for _, a := range portalArgs {
		if !gateNames[a.Name] {
			merged = append(merged, a)
		}
	}
	merged = append(merged, gateArgs...)
	return merged
}

// resolveArgsFromMap resolves args from a map (POST body).
func resolveArgsFromMap(argDefs []metadata.PortalArg, argsMap map[string]any) (map[string]any, error) {
	if len(argDefs) == 0 {
		return map[string]any{}, nil
	}

	result := make(map[string]any, len(argDefs))
	for _, arg := range argDefs {
		val, exists := argsMap[arg.Name]

		if !exists || val == nil {
			if arg.Default != nil {
				converted, err := convertArg(arg.Name, arg.Type, *arg.Default)
				if err != nil {
					return nil, err
				}
				result[arg.Name] = converted
				continue
			}
			return nil, apperror.BadRequest(fmt.Sprintf("required arg %q is missing", arg.Name))
		}

		// Convert to expected type
		raw := fmt.Sprintf("%v", val)
		converted, err := convertArg(arg.Name, arg.Type, raw)
		if err != nil {
			return nil, err
		}
		result[arg.Name] = converted
	}

	return result, nil
}

// validateArgRulesRuntime evaluates CEL arg rules at request time.
func validateArgRulesRuntime(rules []metadata.PortalArgRule, argsMap map[string]any, celCache *celutil.ProgramCache) error {
	if len(rules) == 0 {
		return nil
	}
	vars := map[string]any{"args": argsMap}
	for _, rule := range rules {
		result, err := celCache.EvaluateBool(rule.Condition, vars)
		if err != nil {
			return apperror.BadRequest(fmt.Sprintf("arg_rule %q: evaluation error: %v", rule.Name, err))
		}
		if !result {
			return apperror.BadRequest(rule.ErrorMessage)
		}
	}
	return nil
}

// resolveArgs parses URL query params against declared arg definitions.
// Returns a typed map for CEL and SOQL substitution.
func resolveArgs(argDefs []metadata.PortalArg, c *gin.Context) (map[string]any, error) {
	if len(argDefs) == 0 {
		return map[string]any{}, nil
	}

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

// substituteTypedArgs replaces :paramName in SOQL/DML with typed values.
// Strings are single-quoted (with escaping), numbers and bools are unquoted.
func substituteTypedArgs(text string, args map[string]any) string {
	return portalParamRegexp.ReplaceAllStringFunc(text, func(match string) string {
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
