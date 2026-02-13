package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/api"
	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// MetadataHandler handles admin CRUD for metadata resources.
type MetadataHandler struct {
	objectService metadata.ObjectService
	fieldService  metadata.FieldService
}

// NewMetadataHandler creates a new MetadataHandler.
func NewMetadataHandler(objectService metadata.ObjectService, fieldService metadata.FieldService) *MetadataHandler {
	return &MetadataHandler{
		objectService: objectService,
		fieldService:  fieldService,
	}
}

// RegisterRoutes registers metadata admin routes on the given router group.
func (h *MetadataHandler) RegisterRoutes(rg *gin.RouterGroup) {
	meta := rg.Group("/metadata")

	meta.POST("/objects", h.CreateObject)
	meta.GET("/objects", h.ListObjects)
	meta.GET("/objects/:objectId", h.GetObject)
	meta.PUT("/objects/:objectId", h.UpdateObject)
	meta.DELETE("/objects/:objectId", h.DeleteObject)

	meta.POST("/objects/:objectId/fields", h.CreateField)
	meta.GET("/objects/:objectId/fields", h.ListFields)
	meta.GET("/objects/:objectId/fields/:fieldId", h.GetField)
	meta.PUT("/objects/:objectId/fields/:fieldId", h.UpdateField)
	meta.DELETE("/objects/:objectId/fields/:fieldId", h.DeleteField)
}

func (h *MetadataHandler) CreateObject(c *gin.Context) {
	var req api.CreateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreateObjectInput{
		APIName:               req.ApiName,
		Label:                 req.Label,
		PluralLabel:           req.PluralLabel,
		Description:           derefStr(req.Description),
		ObjectType:            metadata.ObjectType(req.ObjectType),
		IsVisibleInSetup:      derefBool(req.IsVisibleInSetup),
		IsCustomFieldsAllowed: derefBool(req.IsCustomFieldsAllowed),
		IsDeleteableObject:    derefBool(req.IsDeleteableObject),
		IsCreateable:          derefBool(req.IsCreateable),
		IsUpdateable:          derefBool(req.IsUpdateable),
		IsDeleteable:          derefBool(req.IsDeleteable),
		IsQueryable:           derefBool(req.IsQueryable),
		IsSearchable:          derefBool(req.IsSearchable),
		HasActivities:         derefBool(req.HasActivities),
		HasNotes:              derefBool(req.HasNotes),
		HasHistoryTracking:    derefBool(req.HasHistoryTracking),
		HasSharingRules:       derefBool(req.HasSharingRules),
		Visibility:            metadata.Visibility(derefVisibility(req.Visibility)),
	}

	obj, err := h.objectService.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.ObjectResponse{Data: toAPIObject(obj)})
}

func (h *MetadataHandler) ListObjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "0"))
	filter := metadata.ObjectFilter{
		Page:    int32(page),
		PerPage: int32(perPage),
	}
	if ot := c.Query("object_type"); ot != "" {
		objectType := metadata.ObjectType(ot)
		filter.ObjectType = &objectType
	}

	objects, total, err := h.objectService.List(c.Request.Context(), filter)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	data := make([]api.ObjectDefinition, 0, len(objects))
	for i := range objects {
		data = append(data, *toAPIObject(&objects[i]))
	}

	page = int(filter.Page)
	perPage = int(filter.PerPage)
	if perPage == 0 {
		perPage = 20
	}
	totalPages := (total + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, api.ObjectListResponse{
		Data: &data,
		Pagination: &api.PaginationMeta{
			Page:       &page,
			PerPage:    &perPage,
			Total:      &total,
			TotalPages: &totalPages,
		},
	})
}

func (h *MetadataHandler) GetObject(c *gin.Context) {
	objectId, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	obj, err := h.objectService.GetByID(c.Request.Context(), objectId)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, api.ObjectResponse{Data: toAPIObject(obj)})
}

func (h *MetadataHandler) UpdateObject(c *gin.Context) {
	objectId, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	var req api.UpdateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateObjectInput{
		Label:                 req.Label,
		PluralLabel:           req.PluralLabel,
		Description:           derefStr(req.Description),
		IsVisibleInSetup:      derefBool(req.IsVisibleInSetup),
		IsCustomFieldsAllowed: derefBool(req.IsCustomFieldsAllowed),
		IsDeleteableObject:    derefBool(req.IsDeleteableObject),
		IsCreateable:          derefBool(req.IsCreateable),
		IsUpdateable:          derefBool(req.IsUpdateable),
		IsDeleteable:          derefBool(req.IsDeleteable),
		IsQueryable:           derefBool(req.IsQueryable),
		IsSearchable:          derefBool(req.IsSearchable),
		HasActivities:         derefBool(req.HasActivities),
		HasNotes:              derefBool(req.HasNotes),
		HasHistoryTracking:    derefBool(req.HasHistoryTracking),
		HasSharingRules:       derefBool(req.HasSharingRules),
		Visibility:            metadata.Visibility(derefVisibility(req.Visibility)),
	}

	obj, err := h.objectService.Update(c.Request.Context(), objectId, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, api.ObjectResponse{Data: toAPIObject(obj)})
}

func (h *MetadataHandler) DeleteObject(c *gin.Context) {
	objectId, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	if err := h.objectService.Delete(c.Request.Context(), objectId); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MetadataHandler) CreateField(c *gin.Context) {
	objectId, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	var req api.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreateFieldInput{
		ObjectID:  objectId,
		APIName:   req.ApiName,
		Label:     req.Label,
		FieldType: metadata.FieldType(req.FieldType),
		IsCustom:  derefBool(req.IsCustom),
	}

	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.HelpText != nil {
		input.HelpText = *req.HelpText
	}
	if req.FieldSubtype != nil {
		sub := metadata.FieldSubtype(*req.FieldSubtype)
		input.FieldSubtype = &sub
	}
	if req.ReferencedObjectId != nil {
		id := uuid.UUID(*req.ReferencedObjectId)
		input.ReferencedObjectID = &id
	}
	if req.IsRequired != nil {
		input.IsRequired = *req.IsRequired
	}
	if req.IsUnique != nil {
		input.IsUnique = *req.IsUnique
	}
	if req.SortOrder != nil {
		input.SortOrder = *req.SortOrder
	}
	if req.Config != nil {
		input.Config = toMetadataFieldConfig(req.Config)
	}

	field, err := h.fieldService.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.FieldResponse{Data: toAPIField(field)})
}

func (h *MetadataHandler) ListFields(c *gin.Context) {
	objectId, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	fields, err := h.fieldService.ListByObjectID(c.Request.Context(), objectId)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	data := make([]api.FieldDefinitionSchema, 0, len(fields))
	for i := range fields {
		data = append(data, *toAPIField(&fields[i]))
	}

	c.JSON(http.StatusOK, api.FieldListResponse{Data: &data})
}

func (h *MetadataHandler) GetField(c *gin.Context) {
	fieldId, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid field ID"))
		return
	}

	field, err := h.fieldService.GetByID(c.Request.Context(), fieldId)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, api.FieldResponse{Data: toAPIField(field)})
}

func (h *MetadataHandler) UpdateField(c *gin.Context) {
	fieldId, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid field ID"))
		return
	}

	var req api.UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateFieldInput{
		Label: req.Label,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.HelpText != nil {
		input.HelpText = *req.HelpText
	}
	if req.IsRequired != nil {
		input.IsRequired = *req.IsRequired
	}
	if req.IsUnique != nil {
		input.IsUnique = *req.IsUnique
	}
	if req.SortOrder != nil {
		input.SortOrder = *req.SortOrder
	}
	if req.Config != nil {
		input.Config = toMetadataFieldConfig(req.Config)
	}

	field, err := h.fieldService.Update(c.Request.Context(), fieldId, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, api.FieldResponse{Data: toAPIField(field)})
}

func (h *MetadataHandler) DeleteField(c *gin.Context) {
	fieldId, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid field ID"))
		return
	}

	if err := h.fieldService.Delete(c.Request.Context(), fieldId); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Conversion helpers

func toAPIObject(obj *metadata.ObjectDefinition) *api.ObjectDefinition {
	objType := api.ObjectDefinitionObjectType(obj.ObjectType)
	return &api.ObjectDefinition{
		Id:                    &obj.ID,
		ApiName:               &obj.APIName,
		Label:                 &obj.Label,
		PluralLabel:           &obj.PluralLabel,
		Description:           &obj.Description,
		ObjectType:            &objType,
		IsPlatformManaged:     &obj.IsPlatformManaged,
		IsVisibleInSetup:      &obj.IsVisibleInSetup,
		IsCustomFieldsAllowed: &obj.IsCustomFieldsAllowed,
		IsDeleteableObject:    &obj.IsDeleteableObject,
		IsCreateable:          &obj.IsCreateable,
		IsUpdateable:          &obj.IsUpdateable,
		IsDeleteable:          &obj.IsDeleteable,
		IsQueryable:           &obj.IsQueryable,
		IsSearchable:          &obj.IsSearchable,
		HasActivities:         &obj.HasActivities,
		HasNotes:              &obj.HasNotes,
		HasHistoryTracking:    &obj.HasHistoryTracking,
		HasSharingRules:       &obj.HasSharingRules,
		Visibility:            (*api.ObjectDefinitionVisibility)(&obj.Visibility),
		CreatedAt:             &obj.CreatedAt,
		UpdatedAt:             &obj.UpdatedAt,
	}
}

func toAPIField(f *metadata.FieldDefinition) *api.FieldDefinitionSchema {
	ft := string(f.FieldType)
	result := &api.FieldDefinitionSchema{
		Id:                 &f.ID,
		ObjectId:           &f.ObjectID,
		ApiName:            &f.APIName,
		Label:              &f.Label,
		Description:        &f.Description,
		HelpText:           &f.HelpText,
		FieldType:          &ft,
		ReferencedObjectId: f.ReferencedObjectID,
		IsRequired:         &f.IsRequired,
		IsUnique:           &f.IsUnique,
		IsSystemField:      &f.IsSystemField,
		IsCustom:           &f.IsCustom,
		IsPlatformManaged:  &f.IsPlatformManaged,
		SortOrder:          &f.SortOrder,
		CreatedAt:          &f.CreatedAt,
		UpdatedAt:          &f.UpdatedAt,
	}

	if f.FieldSubtype != nil {
		sub := string(*f.FieldSubtype)
		result.FieldSubtype = &sub
	}

	result.Config = toAPIFieldConfig(&f.Config)

	return result
}

func toAPIFieldConfig(fc *metadata.FieldConfig) *api.FieldConfig {
	result := &api.FieldConfig{
		MaxLength:        fc.MaxLength,
		Precision:        fc.Precision,
		Scale:            fc.Scale,
		Format:           fc.Format,
		StartValue:       fc.StartValue,
		RelationshipName: fc.RelationshipName,
		IsReparentable:   fc.IsReparentable,
		DefaultValue:     fc.DefaultValue,
	}
	if fc.OnDelete != nil {
		od := api.FieldConfigOnDelete(*fc.OnDelete)
		result.OnDelete = &od
	}
	return result
}

func toMetadataFieldConfig(fc *api.FieldConfig) metadata.FieldConfig {
	result := metadata.FieldConfig{
		MaxLength:        fc.MaxLength,
		Precision:        fc.Precision,
		Scale:            fc.Scale,
		Format:           fc.Format,
		StartValue:       fc.StartValue,
		RelationshipName: fc.RelationshipName,
		IsReparentable:   fc.IsReparentable,
		DefaultValue:     fc.DefaultValue,
	}
	if fc.OnDelete != nil {
		od := string(*fc.OnDelete)
		result.OnDelete = &od
	}
	return result
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func derefVisibility[T ~string](v *T) string {
	if v == nil {
		return "private"
	}
	return string(*v)
}
