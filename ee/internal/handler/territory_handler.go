//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/ee/internal/platform/territory"
	"github.com/adverax/crm/internal/pkg/apperror"
)

// TerritoryHandler handles admin CRUD for territory management resources.
type TerritoryHandler struct {
	modelService      territory.ModelService
	territoryService  territory.TerritoryService
	objDefaultService territory.ObjectDefaultService
	userAssignService territory.UserAssignmentService
	recAssignService  territory.RecordAssignmentService
	ruleService       territory.AssignmentRuleService
}

// NewTerritoryHandler creates a new TerritoryHandler.
func NewTerritoryHandler(
	modelService territory.ModelService,
	territoryService territory.TerritoryService,
	objDefaultService territory.ObjectDefaultService,
	userAssignService territory.UserAssignmentService,
	recAssignService territory.RecordAssignmentService,
	ruleService territory.AssignmentRuleService,
) *TerritoryHandler {
	return &TerritoryHandler{
		modelService:      modelService,
		territoryService:  territoryService,
		objDefaultService: objDefaultService,
		userAssignService: userAssignService,
		recAssignService:  recAssignService,
		ruleService:       ruleService,
	}
}

// RegisterRoutes registers territory admin routes on the given router group.
func (h *TerritoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	terr := rg.Group("/territory")

	// Models
	terr.POST("/models", h.CreateModel)
	terr.GET("/models", h.ListModels)
	terr.GET("/models/:id", h.GetModel)
	terr.PUT("/models/:id", h.UpdateModel)
	terr.DELETE("/models/:id", h.DeleteModel)
	terr.POST("/models/:id/activate", h.ActivateModel)
	terr.POST("/models/:id/archive", h.ArchiveModel)

	// Territories
	terr.POST("/territories", h.CreateTerritory)
	terr.GET("/territories", h.ListTerritories)
	terr.GET("/territories/:id", h.GetTerritory)
	terr.PUT("/territories/:id", h.UpdateTerritory)
	terr.DELETE("/territories/:id", h.DeleteTerritory)

	// Object defaults
	terr.POST("/territories/:id/object-defaults", h.SetObjectDefault)
	terr.GET("/territories/:id/object-defaults", h.ListObjectDefaults)
	terr.DELETE("/territories/:id/object-defaults/:objectId", h.RemoveObjectDefault)

	// User assignments
	terr.POST("/territories/:id/users", h.AssignUser)
	terr.GET("/territories/:id/users", h.ListTerritoryUsers)
	terr.DELETE("/territories/:id/users/:userId", h.UnassignUser)

	// Record assignments
	terr.POST("/territories/:id/records", h.AssignRecord)
	terr.GET("/territories/:id/records", h.ListTerritoryRecords)
	terr.DELETE("/territories/:id/records/:recordId", h.UnassignRecord)

	// Assignment rules
	terr.POST("/assignment-rules", h.CreateAssignmentRule)
	terr.GET("/assignment-rules", h.ListAssignmentRules)
	terr.GET("/assignment-rules/:id", h.GetAssignmentRule)
	terr.PUT("/assignment-rules/:id", h.UpdateAssignmentRule)
	terr.DELETE("/assignment-rules/:id", h.DeleteAssignmentRule)
}

// --- Models ---

func (h *TerritoryHandler) CreateModel(c *gin.Context) {
	var req territory.CreateModelInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	model, err := h.modelService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": model})
}

func (h *TerritoryHandler) ListModels(c *gin.Context) {
	page, perPage := parsePagination(c)
	models, total, err := h.modelService.List(c.Request.Context(), page, perPage)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       models,
		"pagination": paginationMeta(page, perPage, total),
	})
}

func (h *TerritoryHandler) GetModel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	model, err := h.modelService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": model})
}

func (h *TerritoryHandler) UpdateModel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req territory.UpdateModelInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	model, err := h.modelService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": model})
}

func (h *TerritoryHandler) DeleteModel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.modelService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TerritoryHandler) ActivateModel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.modelService.Activate(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TerritoryHandler) ArchiveModel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.modelService.Archive(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Territories ---

func (h *TerritoryHandler) CreateTerritory(c *gin.Context) {
	var req territory.CreateTerritoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	t, err := h.territoryService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (h *TerritoryHandler) ListTerritories(c *gin.Context) {
	modelIDStr := c.Query("model_id")
	if modelIDStr == "" {
		apperror.Respond(c, apperror.BadRequest("model_id query parameter is required"))
		return
	}
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid model_id"))
		return
	}
	territories, err := h.territoryService.ListByModelID(c.Request.Context(), modelID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": territories})
}

func (h *TerritoryHandler) GetTerritory(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	t, err := h.territoryService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (h *TerritoryHandler) UpdateTerritory(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req territory.UpdateTerritoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	t, err := h.territoryService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (h *TerritoryHandler) DeleteTerritory(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.territoryService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Object Defaults ---

type setObjectDefaultRequest struct {
	ObjectID    uuid.UUID `json:"object_id"`
	AccessLevel string    `json:"access_level"`
}

func (h *TerritoryHandler) SetObjectDefault(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req setObjectDefaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	input := territory.SetObjectDefaultInput{
		TerritoryID: territoryID,
		ObjectID:    req.ObjectID,
		AccessLevel: req.AccessLevel,
	}
	result, err := h.objDefaultService.Set(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TerritoryHandler) ListObjectDefaults(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	defaults, err := h.objDefaultService.ListByTerritoryID(c.Request.Context(), territoryID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": defaults})
}

func (h *TerritoryHandler) RemoveObjectDefault(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	objectID, err := parseUUID(c, "objectId")
	if err != nil {
		return
	}
	if err := h.objDefaultService.Remove(c.Request.Context(), territoryID, objectID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- User Assignments ---

type assignUserRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

func (h *TerritoryHandler) AssignUser(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req assignUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	input := territory.AssignUserInput{
		UserID:      req.UserID,
		TerritoryID: territoryID,
	}
	result, err := h.userAssignService.Assign(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *TerritoryHandler) ListTerritoryUsers(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	assignments, err := h.userAssignService.ListByTerritoryID(c.Request.Context(), territoryID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

func (h *TerritoryHandler) UnassignUser(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := parseUUID(c, "userId")
	if err != nil {
		return
	}
	if err := h.userAssignService.Unassign(c.Request.Context(), userID, territoryID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Record Assignments ---

type assignRecordRequest struct {
	RecordID uuid.UUID `json:"record_id"`
	ObjectID uuid.UUID `json:"object_id"`
	Reason   string    `json:"reason"`
}

func (h *TerritoryHandler) AssignRecord(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req assignRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	input := territory.AssignRecordInput{
		RecordID:    req.RecordID,
		ObjectID:    req.ObjectID,
		TerritoryID: territoryID,
		Reason:      req.Reason,
	}
	result, err := h.recAssignService.Assign(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *TerritoryHandler) ListTerritoryRecords(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	assignments, err := h.recAssignService.ListByTerritoryID(c.Request.Context(), territoryID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

func (h *TerritoryHandler) UnassignRecord(c *gin.Context) {
	territoryID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	recordID, err := parseUUID(c, "recordId")
	if err != nil {
		return
	}
	objectIDStr := c.Query("object_id")
	if objectIDStr == "" {
		apperror.Respond(c, apperror.BadRequest("object_id query parameter is required"))
		return
	}
	objectID, err := uuid.Parse(objectIDStr)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object_id"))
		return
	}
	if err := h.recAssignService.Unassign(c.Request.Context(), recordID, objectID, territoryID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Assignment Rules ---

func (h *TerritoryHandler) CreateAssignmentRule(c *gin.Context) {
	var req territory.CreateAssignmentRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	rule, err := h.ruleService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

func (h *TerritoryHandler) ListAssignmentRules(c *gin.Context) {
	territoryIDStr := c.Query("territory_id")
	if territoryIDStr == "" {
		apperror.Respond(c, apperror.BadRequest("territory_id query parameter is required"))
		return
	}
	territoryID, err := uuid.Parse(territoryIDStr)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid territory_id"))
		return
	}
	rules, err := h.ruleService.ListByTerritoryID(c.Request.Context(), territoryID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *TerritoryHandler) GetAssignmentRule(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	rule, err := h.ruleService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *TerritoryHandler) UpdateAssignmentRule(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req territory.UpdateAssignmentRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	rule, err := h.ruleService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *TerritoryHandler) DeleteAssignmentRule(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.ruleService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Helpers ---

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid UUID: "+param))
		return uuid.Nil, err
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int32, int32) {
	var page, perPage int32 = 1, 20
	if v := c.Query("page"); v != "" {
		if p := parseInt32(v); p > 0 {
			page = p
		}
	}
	if v := c.Query("per_page"); v != "" {
		if p := parseInt32(v); p > 0 {
			perPage = p
		}
	}
	return page, perPage
}

func parseInt32(s string) int32 {
	var n int32
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int32(c-'0')
	}
	return n
}

func paginationMeta(page, perPage int32, total int64) gin.H {
	totalPages := (total + int64(perPage) - 1) / int64(perPage)
	return gin.H{
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	}
}
