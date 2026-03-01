package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// PortalAdminHandler handles admin CRUD for object views.
type PortalAdminHandler struct {
	service metadata.PortalService
}

// NewPortalAdminHandler creates a new PortalAdminHandler.
func NewPortalAdminHandler(service metadata.PortalService) *PortalAdminHandler {
	return &PortalAdminHandler{service: service}
}

// RegisterRoutes registers object view routes on the admin group.
func (h *PortalAdminHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/portals", h.Create)
	rg.GET("/portals", h.List)
	rg.GET("/portals/:portalId", h.Get)
	rg.PUT("/portals/:portalId", h.Update)
	rg.DELETE("/portals/:portalId", h.Delete)
}

type createPortalRequest struct {
	ProfileID   *string               `json:"profile_id"`
	APIName     string                `json:"api_name" binding:"required"`
	Label       string                `json:"label" binding:"required"`
	Description *string               `json:"description"`
	Config      metadata.PortalConfig `json:"config"`
}

type updatePortalRequest struct {
	Label       string                `json:"label" binding:"required"`
	Description *string               `json:"description"`
	Config      metadata.PortalConfig `json:"config"`
}

func (h *PortalAdminHandler) Create(c *gin.Context) {
	var req createPortalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreatePortalInput{
		APIName: req.APIName,
		Label:   req.Label,
		Config:  req.Config,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.ProfileID != nil {
		pid, err := uuid.Parse(*req.ProfileID)
		if err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid profile_id"))
			return
		}
		input.ProfileID = &pid
	}

	ov, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ov})
}

func (h *PortalAdminHandler) List(c *gin.Context) {
	views, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": views})
}

func (h *PortalAdminHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("portalId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	ov, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

func (h *PortalAdminHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("portalId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	var req updatePortalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdatePortalInput{
		Label:  req.Label,
		Config: req.Config,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}

	ov, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

func (h *PortalAdminHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("portalId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
