package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// ViewHandler serves resolved Object View configs.
type ViewHandler struct {
	cache metadata.MetadataReader
}

// NewViewHandler creates a new ViewHandler.
func NewViewHandler(cache metadata.MetadataReader) *ViewHandler {
	return &ViewHandler{cache: cache}
}

// RegisterRoutes registers the view route on the given API group.
func (h *ViewHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/view/:ovApiName", h.GetByAPIName)
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
