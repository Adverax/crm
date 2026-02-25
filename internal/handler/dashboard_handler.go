package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/soql"
)

const (
	dashboardWidgetTimeout = 5 * time.Second
)

// DashboardHandler handles admin CRUD and resolution for profile dashboards.
type DashboardHandler struct {
	service     metadata.ProfileDashboardService
	soqlService soql.QueryService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	service metadata.ProfileDashboardService,
	soqlService soql.QueryService,
) *DashboardHandler {
	return &DashboardHandler{
		service:     service,
		soqlService: soqlService,
	}
}

// RegisterRoutes registers dashboard routes on admin and public groups.
func (h *DashboardHandler) RegisterRoutes(adminGroup, apiGroup *gin.RouterGroup) {
	adminGroup.POST("/profile-dashboards", h.Create)
	adminGroup.GET("/profile-dashboards", h.List)
	adminGroup.GET("/profile-dashboards/:id", h.Get)
	adminGroup.PUT("/profile-dashboards/:id", h.Update)
	adminGroup.DELETE("/profile-dashboards/:id", h.Delete)

	apiGroup.GET("/dashboard", h.Resolve)
}

type createDashboardRequest struct {
	ProfileID string                   `json:"profile_id" binding:"required"`
	Config    metadata.DashboardConfig `json:"config"`
}

type updateDashboardRequest struct {
	Config metadata.DashboardConfig `json:"config"`
}

func (h *DashboardHandler) Create(c *gin.Context) {
	var req createDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	profileID, err := uuid.Parse(req.ProfileID)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid profile_id"))
		return
	}

	input := metadata.CreateProfileDashboardInput{
		ProfileID: profileID,
		Config:    req.Config,
	}

	dash, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dash})
}

func (h *DashboardHandler) List(c *gin.Context) {
	dashes, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dashes})
}

func (h *DashboardHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid ID"))
		return
	}

	dash, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dash})
}

func (h *DashboardHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid ID"))
		return
	}

	var req updateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateProfileDashboardInput{
		Config: req.Config,
	}

	dash, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dash})
}

func (h *DashboardHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type resolvedWidget struct {
	Key           string              `json:"key"`
	Type          string              `json:"type"`
	Label         string              `json:"label"`
	Size          string              `json:"size"`
	ObjectAPIName string              `json:"object_api_name,omitempty"`
	Columns       []string            `json:"columns,omitempty"`
	Format        string              `json:"format,omitempty"`
	Links         []metadata.DashLink `json:"links,omitempty"`
	Data          any                 `json:"data"`
}

type listWidgetData struct {
	Records    []map[string]any `json:"records"`
	TotalCount int              `json:"total_count"`
}

type metricWidgetData struct {
	Value any `json:"value"`
}

type resolvedDashboard struct {
	Widgets []resolvedWidget `json:"widgets"`
}

// Resolve returns the executed dashboard for the current user's profile.
func (h *DashboardHandler) Resolve(c *gin.Context) {
	uc, ok := security.UserFromContext(c.Request.Context())
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("user context required"))
		return
	}

	dash, err := h.service.ResolveForProfile(c.Request.Context(), uc.ProfileID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	if dash == nil {
		// Empty dashboard fallback
		c.JSON(http.StatusOK, gin.H{"data": resolvedDashboard{Widgets: []resolvedWidget{}}})
		return
	}

	widgets := make([]resolvedWidget, 0, len(dash.Config.Widgets))
	for _, w := range dash.Config.Widgets {
		resolved := h.executeWidget(c.Request.Context(), uc.UserID, w)
		widgets = append(widgets, resolved)
	}

	c.JSON(http.StatusOK, gin.H{"data": resolvedDashboard{Widgets: widgets}})
}

func (h *DashboardHandler) executeWidget(ctx context.Context, userID uuid.UUID, w metadata.DashboardWidget) resolvedWidget {
	rw := resolvedWidget{
		Key:           w.Key,
		Type:          w.Type,
		Label:         w.Label,
		Size:          w.Size,
		ObjectAPIName: w.ObjectAPIName,
		Columns:       w.Columns,
		Format:        w.Format,
		Links:         w.Links,
	}

	switch w.Type {
	case "list":
		rw.Data = h.executeListWidget(ctx, userID, w)
	case "metric":
		rw.Data = h.executeMetricWidget(ctx, userID, w)
	case "link_list":
		rw.Data = nil
	}

	return rw
}

func (h *DashboardHandler) executeListWidget(ctx context.Context, userID uuid.UUID, w metadata.DashboardWidget) listWidgetData {
	queryCtx, cancel := context.WithTimeout(ctx, dashboardWidgetTimeout)
	defer cancel()

	query := substituteVars(w.Query, userID)
	result, err := h.soqlService.Execute(queryCtx, query, nil)
	if err != nil {
		slog.Warn("dashboard: list widget query failed", "widget", w.Key, "error", err)
		return listWidgetData{Records: []map[string]any{}, TotalCount: 0}
	}

	return listWidgetData{
		Records:    result.Records,
		TotalCount: result.TotalSize,
	}
}

func (h *DashboardHandler) executeMetricWidget(ctx context.Context, userID uuid.UUID, w metadata.DashboardWidget) metricWidgetData {
	queryCtx, cancel := context.WithTimeout(ctx, dashboardWidgetTimeout)
	defer cancel()

	query := substituteVars(w.Query, userID)
	result, err := h.soqlService.Execute(queryCtx, query, nil)
	if err != nil {
		slog.Warn("dashboard: metric widget query failed", "widget", w.Key, "error", err)
		return metricWidgetData{Value: 0}
	}

	if len(result.Records) > 0 {
		// Return the first value from the first record
		for _, v := range result.Records[0] {
			return metricWidgetData{Value: v}
		}
	}

	return metricWidgetData{Value: 0}
}

// substituteVars replaces dashboard-specific variables in SOQL queries.
func substituteVars(query string, userID uuid.UUID) string {
	return strings.ReplaceAll(query, ":currentUserId", "'"+userID.String()+"'")
}
