package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/templates"
)

// TemplateHandler handles admin endpoints for application templates.
type TemplateHandler struct {
	registry *templates.Registry
	applier  *templates.Applier
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(registry *templates.Registry, applier *templates.Applier) *TemplateHandler {
	return &TemplateHandler{
		registry: registry,
		applier:  applier,
	}
}

// RegisterRoutes registers template admin routes on the given router group.
func (h *TemplateHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/templates", h.ListTemplates)
	rg.POST("/templates/:templateId/apply", h.ApplyTemplate)
}

// ListTemplates returns all available templates with their status.
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	all := h.registry.List()
	result := make([]templates.TemplateInfo, 0, len(all))

	for _, tmpl := range all {
		result = append(result, templates.TemplateInfo{
			ID:          tmpl.ID,
			Label:       tmpl.Label,
			Description: tmpl.Description,
			Status:      templates.TemplateStatusAvailable,
			Objects:     len(tmpl.Objects),
			Fields:      len(tmpl.Fields),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ApplyTemplate applies a template, creating all its objects and fields.
func (h *TemplateHandler) ApplyTemplate(c *gin.Context) {
	templateID := c.Param("templateId")

	tmpl, ok := h.registry.Get(templateID)
	if !ok {
		apperror.Respond(c, apperror.NotFound("template", templateID))
		return
	}

	if err := h.applier.Apply(c.Request.Context(), tmpl); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"template_id": tmpl.ID,
		"message":     "template applied successfully",
	}})
}
