package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/ols"
)

// NavigationHandler handles admin CRUD and resolution for profile navigation.
type NavigationHandler struct {
	service     metadata.ProfileNavigationService
	cache       metadata.MetadataReader
	olsEnforcer ols.Enforcer
}

// NewNavigationHandler creates a new NavigationHandler.
func NewNavigationHandler(
	service metadata.ProfileNavigationService,
	cache metadata.MetadataReader,
	olsEnforcer ols.Enforcer,
) *NavigationHandler {
	return &NavigationHandler{
		service:     service,
		cache:       cache,
		olsEnforcer: olsEnforcer,
	}
}

// RegisterRoutes registers navigation routes on admin and public groups.
func (h *NavigationHandler) RegisterRoutes(adminGroup, apiGroup *gin.RouterGroup) {
	adminGroup.POST("/profile-navigation", h.Create)
	adminGroup.GET("/profile-navigation", h.List)
	adminGroup.GET("/profile-navigation/:id", h.Get)
	adminGroup.PUT("/profile-navigation/:id", h.Update)
	adminGroup.DELETE("/profile-navigation/:id", h.Delete)

	apiGroup.GET("/navigation", h.Resolve)
}

type createNavigationRequest struct {
	ProfileID string             `json:"profile_id" binding:"required"`
	Config    metadata.NavConfig `json:"config"`
}

type updateNavigationRequest struct {
	Config metadata.NavConfig `json:"config"`
}

func (h *NavigationHandler) Create(c *gin.Context) {
	var req createNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	profileID, err := uuid.Parse(req.ProfileID)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid profile_id"))
		return
	}

	input := metadata.CreateProfileNavigationInput{
		ProfileID: profileID,
		Config:    req.Config,
	}

	nav, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": nav})
}

func (h *NavigationHandler) List(c *gin.Context) {
	navs, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": navs})
}

func (h *NavigationHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid ID"))
		return
	}

	nav, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nav})
}

func (h *NavigationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid ID"))
		return
	}

	var req updateNavigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateProfileNavigationInput{
		Config: req.Config,
	}

	nav, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nav})
}

func (h *NavigationHandler) Delete(c *gin.Context) {
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

// resolvedNavItem is a navigation item with resolved label for object items.
type resolvedNavItem struct {
	Type          string `json:"type"`
	ObjectAPIName string `json:"object_api_name,omitempty"`
	Label         string `json:"label,omitempty"`
	PluralLabel   string `json:"plural_label,omitempty"`
	URL           string `json:"url,omitempty"`
	Icon          string `json:"icon,omitempty"`
}

type resolvedNavGroup struct {
	Key   string            `json:"key"`
	Label string            `json:"label"`
	Icon  string            `json:"icon,omitempty"`
	Items []resolvedNavItem `json:"items"`
}

type resolvedNavigation struct {
	Groups []resolvedNavGroup `json:"groups"`
}

// Resolve returns the navigation config for the current user's profile.
// If no config exists, returns an OLS-filtered fallback.
func (h *NavigationHandler) Resolve(c *gin.Context) {
	uc, ok := security.UserFromContext(c.Request.Context())
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("user context required"))
		return
	}

	nav, err := h.service.ResolveForProfile(c.Request.Context(), uc.ProfileID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	if nav != nil {
		resolved := h.resolveNavConfig(c, uc.UserID, nav.Config)
		c.JSON(http.StatusOK, gin.H{"data": resolved})
		return
	}

	// Fallback: OLS-filtered alphabetical list (no groups)
	fallback := h.buildFallbackNavigation(c, uc.UserID)
	c.JSON(http.StatusOK, gin.H{"data": fallback})
}

func (h *NavigationHandler) resolveNavConfig(c *gin.Context, userID uuid.UUID, cfg metadata.NavConfig) resolvedNavigation {
	var groups []resolvedNavGroup

	for _, g := range cfg.Groups {
		var items []resolvedNavItem
		for _, item := range g.Items {
			resolved := h.resolveNavItem(c, userID, item)
			if resolved != nil {
				items = append(items, *resolved)
			}
		}
		if len(items) > 0 {
			groups = append(groups, resolvedNavGroup{
				Key:   g.Key,
				Label: g.Label,
				Icon:  g.Icon,
				Items: items,
			})
		}
	}

	if groups == nil {
		groups = []resolvedNavGroup{}
	}
	return resolvedNavigation{Groups: groups}
}

func (h *NavigationHandler) resolveNavItem(c *gin.Context, userID uuid.UUID, item metadata.NavItem) *resolvedNavItem {
	switch item.Type {
	case "object":
		objDef, ok := h.cache.GetObjectByAPIName(item.ObjectAPIName)
		if !ok {
			return nil
		}
		if err := h.olsEnforcer.CanRead(c.Request.Context(), userID, objDef.ID); err != nil {
			return nil
		}
		return &resolvedNavItem{
			Type:          "object",
			ObjectAPIName: objDef.APIName,
			Label:         objDef.Label,
			PluralLabel:   objDef.PluralLabel,
		}
	case "link":
		return &resolvedNavItem{
			Type:  "link",
			Label: item.Label,
			URL:   item.URL,
			Icon:  item.Icon,
		}
	case "divider":
		return &resolvedNavItem{Type: "divider"}
	default:
		return nil
	}
}

func (h *NavigationHandler) buildFallbackNavigation(c *gin.Context, userID uuid.UUID) resolvedNavigation {
	names := h.cache.ListObjectAPINames()
	var items []resolvedNavItem

	for _, name := range names {
		objDef, ok := h.cache.GetObjectByAPIName(name)
		if !ok || !objDef.IsQueryable {
			continue
		}
		if err := h.olsEnforcer.CanRead(c.Request.Context(), userID, objDef.ID); err != nil {
			continue
		}
		items = append(items, resolvedNavItem{
			Type:          "object",
			ObjectAPIName: objDef.APIName,
			Label:         objDef.Label,
			PluralLabel:   objDef.PluralLabel,
		})
	}

	if items == nil {
		items = []resolvedNavItem{}
	}

	// Return a single unnamed group for flat rendering
	return resolvedNavigation{
		Groups: []resolvedNavGroup{
			{Key: "_default", Label: "", Items: items},
		},
	}
}
