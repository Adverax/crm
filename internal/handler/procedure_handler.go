package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/procedure"
)

// ProcedureHandler handles admin CRUD, versioning, and execution for procedures.
type ProcedureHandler struct {
	service  metadata.ProcedureService
	executor procedure.ProcedureExecutor
}

// NewProcedureHandler creates a new ProcedureHandler.
func NewProcedureHandler(service metadata.ProcedureService, executor procedure.ProcedureExecutor) *ProcedureHandler {
	return &ProcedureHandler{service: service, executor: executor}
}

// RegisterRoutes registers procedure routes on the admin group.
func (h *ProcedureHandler) RegisterRoutes(admin *gin.RouterGroup) {
	g := admin.Group("/procedures")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:procedureId", h.Get)
	g.PUT("/:procedureId", h.UpdateMetadata)
	g.DELETE("/:procedureId", h.Delete)
	g.PUT("/:procedureId/draft", h.SaveDraft)
	g.DELETE("/:procedureId/draft", h.DiscardDraft)
	g.POST("/:procedureId/draft/from-published", h.CreateDraftFromPublished)
	g.POST("/:procedureId/publish", h.Publish)
	g.POST("/:procedureId/rollback", h.Rollback)
	g.GET("/:procedureId/versions", h.ListVersions)
	g.POST("/:procedureId/execute", h.Execute)
	g.POST("/:procedureId/dry-run", h.DryRunExec)
}

type createProcedureRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type updateProcedureMetadataRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type saveDraftRequest struct {
	Definition    metadata.ProcedureDefinition `json:"definition" binding:"required"`
	ChangeSummary string                       `json:"change_summary"`
}

func (h *ProcedureHandler) Create(c *gin.Context) {
	var req createProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	result, err := h.service.Create(c.Request.Context(), metadata.CreateProcedureInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *ProcedureHandler) List(c *gin.Context) {
	procedures, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": procedures})
}

func (h *ProcedureHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ProcedureHandler) UpdateMetadata(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	var req updateProcedureMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	proc, err := h.service.UpdateMetadata(c.Request.Context(), id, metadata.UpdateProcedureMetadataInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proc})
}

func (h *ProcedureHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProcedureHandler) SaveDraft(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	var req saveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	version, err := h.service.SaveDraft(c.Request.Context(), id, metadata.SaveDraftInput{
		Definition:    req.Definition,
		ChangeSummary: req.ChangeSummary,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": version})
}

func (h *ProcedureHandler) DiscardDraft(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	if err := h.service.DiscardDraft(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProcedureHandler) CreateDraftFromPublished(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	version, err := h.service.CreateDraftFromPublished(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": version})
}

func (h *ProcedureHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	version, err := h.service.Publish(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": version})
}

func (h *ProcedureHandler) Rollback(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	version, err := h.service.Rollback(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": version})
}

func (h *ProcedureHandler) ListVersions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	versions, err := h.service.ListVersions(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": versions})
}

type executeRequest struct {
	Input map[string]any `json:"input"`
}

func (h *ProcedureHandler) Execute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	var req executeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	proc, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	if proc.Procedure.PublishedVersionID == nil {
		apperror.Respond(c, apperror.BadRequest("no published version to execute"))
		return
	}

	if h.executor == nil {
		apperror.Respond(c, apperror.Internal("procedure executor not configured"))
		return
	}

	result, err := h.executor.Execute(c.Request.Context(), proc.Procedure.Code, req.Input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ProcedureHandler) DryRunExec(c *gin.Context) {
	id, err := uuid.Parse(c.Param("procedureId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid procedure ID"))
		return
	}

	var req executeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	proc, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	if h.executor == nil {
		apperror.Respond(c, apperror.Internal("procedure executor not configured"))
		return
	}

	// Use draft if exists, else published
	var def *metadata.ProcedureDefinition
	if proc.DraftVersion != nil {
		def = &proc.DraftVersion.Definition
	} else if proc.PublishedVersion != nil {
		def = &proc.PublishedVersion.Definition
	} else {
		apperror.Respond(c, apperror.BadRequest("no version available for dry-run"))
		return
	}

	result, err := h.executor.DryRun(c.Request.Context(), def, req.Input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
