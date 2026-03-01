package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
)

var paramRegexp = regexp.MustCompile(`:(\w+)`)

// PortalHandler serves resolved Object View configs, per-query data, and action execution.
type PortalHandler struct {
	cache       metadata.MetadataReader
	soqlService soql.QueryService
	dmlService  dml.DMLService
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(cache metadata.MetadataReader, soqlService soql.QueryService, dmlService dml.DMLService) *PortalHandler {
	return &PortalHandler{
		cache:       cache,
		soqlService: soqlService,
		dmlService:  dmlService,
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

	// Substitute URL query params into SOQL :paramName placeholders
	soqlText := substituteParams(query.SOQL, c)

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
		apperror.Respond(c, fmt.Errorf("viewHandler.ExecuteQuery: %w", err))
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
		apperror.Respond(c, fmt.Errorf("viewHandler.ExecuteAction: %w", err))
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

// substituteParams replaces :paramName in SOQL with URL query parameter values.
func substituteParams(soqlText string, c *gin.Context) string {
	return paramRegexp.ReplaceAllStringFunc(soqlText, func(match string) string {
		paramName := match[1:] // strip leading ':'
		if val := c.Query(paramName); val != "" {
			return "'" + val + "'"
		}
		return match
	})
}
