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

// ViewHandler serves resolved Object View configs, per-query data, and action execution.
type ViewHandler struct {
	cache       metadata.MetadataReader
	soqlService soql.QueryService
	dmlService  dml.DMLService
}

// NewViewHandler creates a new ViewHandler.
func NewViewHandler(cache metadata.MetadataReader, soqlService soql.QueryService, dmlService dml.DMLService) *ViewHandler {
	return &ViewHandler{
		cache:       cache,
		soqlService: soqlService,
		dmlService:  dmlService,
	}
}

// RegisterRoutes registers the view routes on the given API group.
func (h *ViewHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/view/:ovApiName", h.GetByAPIName)
	rg.GET("/view/:ovApiName/query/:queryName", h.ExecuteQuery)
	rg.POST("/view/:ovApiName/action/:actionKey", h.ExecuteAction)
}

// GetByAPIName returns the OV config by api_name.
func (h *ViewHandler) GetByAPIName(c *gin.Context) {
	apiName := c.Param("ovApiName")

	ov, ok := h.cache.GetObjectViewByAPIName(apiName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("object_view", apiName))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

// ExecuteQuery executes a named query from an Object View.
func (h *ViewHandler) ExecuteQuery(c *gin.Context) {
	ovAPIName := c.Param("ovApiName")
	queryName := c.Param("queryName")

	ov, ok := h.cache.GetObjectViewByAPIName(ovAPIName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("object_view", ovAPIName))
		return
	}

	// Find the query
	var query *metadata.OVQuery
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
func (h *ViewHandler) ExecuteAction(c *gin.Context) {
	ovAPIName := c.Param("ovApiName")
	actionKey := c.Param("actionKey")

	ov, ok := h.cache.GetObjectViewByAPIName(ovAPIName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("object_view", ovAPIName))
		return
	}

	// Find the action
	var action *metadata.OVAction
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
