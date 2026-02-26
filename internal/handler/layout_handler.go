package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// LayoutHandler handles admin CRUD for layouts.
type LayoutHandler struct {
	service metadata.LayoutService
}

// NewLayoutHandler creates a new LayoutHandler.
func NewLayoutHandler(service metadata.LayoutService) *LayoutHandler {
	return &LayoutHandler{service: service}
}

// RegisterRoutes registers layout routes on the admin group.
func (h *LayoutHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/layouts", h.Create)
	rg.GET("/layouts", h.List)
	rg.GET("/layouts/:layoutId", h.Get)
	rg.PUT("/layouts/:layoutId", h.Update)
	rg.DELETE("/layouts/:layoutId", h.Delete)
}

type createLayoutRequest struct {
	ObjectViewID string              `json:"object_view_id" binding:"required"`
	FormFactor   string              `json:"form_factor" binding:"required"`
	Mode         string              `json:"mode" binding:"required"`
	Config       metadata.LayoutConfig `json:"config"`
}

type updateLayoutRequest struct {
	Config metadata.LayoutConfig `json:"config"`
}

func (h *LayoutHandler) Create(c *gin.Context) {
	var req createLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	ovID, err := uuid.Parse(req.ObjectViewID)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object_view_id"))
		return
	}

	input := metadata.CreateLayoutInput{
		ObjectViewID: ovID,
		FormFactor:   req.FormFactor,
		Mode:         req.Mode,
		Config:       req.Config,
	}

	layout, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": layout})
}

func (h *LayoutHandler) List(c *gin.Context) {
	ovIDStr := c.Query("object_view_id")
	if ovIDStr != "" {
		ovID, err := uuid.Parse(ovIDStr)
		if err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid object_view_id"))
			return
		}
		layouts, err := h.service.ListByObjectViewID(c.Request.Context(), ovID)
		if err != nil {
			apperror.Respond(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": layouts})
		return
	}

	layouts, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": layouts})
}

func (h *LayoutHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("layoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid layout ID"))
		return
	}

	layout, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": layout})
}

func (h *LayoutHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("layoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid layout ID"))
		return
	}

	var req updateLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateLayoutInput{
		Config: req.Config,
	}

	layout, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": layout})
}

func (h *LayoutHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("layoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid layout ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
