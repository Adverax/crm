package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
)

var paramRegexp = regexp.MustCompile(`:(\w+)`)

// ViewHandler serves resolved Object View configs and per-query data.
type ViewHandler struct {
	cache       metadata.MetadataReader
	soqlService soql.QueryService
}

// NewViewHandler creates a new ViewHandler.
func NewViewHandler(cache metadata.MetadataReader, soqlService soql.QueryService) *ViewHandler {
	return &ViewHandler{
		cache:       cache,
		soqlService: soqlService,
	}
}

// RegisterRoutes registers the view routes on the given API group.
func (h *ViewHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/view/:ovApiName", h.GetByAPIName)
	rg.GET("/view/:ovApiName/query/:queryName", h.ExecuteQuery)
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
	for i := range ov.Config.View.Queries {
		if ov.Config.View.Queries[i].Name == queryName {
			query = &ov.Config.View.Queries[i]
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

	c.JSON(http.StatusOK, gin.H{"data": result})
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
