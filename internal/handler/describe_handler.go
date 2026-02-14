package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/fls"
	"github.com/adverax/crm/internal/platform/security/ols"
)

// DescribeHandler exposes public metadata for frontend consumption.
type DescribeHandler struct {
	cache       *metadata.MetadataCache
	olsEnforcer ols.Enforcer
	flsEnforcer fls.Enforcer
}

// NewDescribeHandler creates a new DescribeHandler.
func NewDescribeHandler(
	cache *metadata.MetadataCache,
	olsEnforcer ols.Enforcer,
	flsEnforcer fls.Enforcer,
) *DescribeHandler {
	return &DescribeHandler{
		cache:       cache,
		olsEnforcer: olsEnforcer,
		flsEnforcer: flsEnforcer,
	}
}

// RegisterRoutes registers describe routes on the given group.
func (h *DescribeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/describe", h.ListObjects)
	rg.GET("/describe/:objectName", h.DescribeObject)
}

type objectNavItem struct {
	APIName      string `json:"api_name"`
	Label        string `json:"label"`
	PluralLabel  string `json:"plural_label"`
	IsCreateable bool   `json:"is_createable"`
	IsQueryable  bool   `json:"is_queryable"`
}

type fieldDescribe struct {
	APIName       string              `json:"api_name"`
	Label         string              `json:"label"`
	FieldType     string              `json:"field_type"`
	FieldSubtype  *string             `json:"field_subtype"`
	IsRequired    bool                `json:"is_required"`
	IsReadOnly    bool                `json:"is_read_only"`
	IsSystemField bool                `json:"is_system_field"`
	SortOrder     int                 `json:"sort_order"`
	Config        fieldConfigDescribe `json:"config"`
}

type fieldConfigDescribe struct {
	MaxLength    *int                     `json:"max_length,omitempty"`
	Precision    *int                     `json:"precision,omitempty"`
	Scale        *int                     `json:"scale,omitempty"`
	DefaultValue *string                  `json:"default_value,omitempty"`
	Values       []metadata.PicklistValue `json:"values,omitempty"`
}

type objectDescribe struct {
	APIName      string          `json:"api_name"`
	Label        string          `json:"label"`
	PluralLabel  string          `json:"plural_label"`
	IsCreateable bool            `json:"is_createable"`
	IsUpdateable bool            `json:"is_updateable"`
	IsDeleteable bool            `json:"is_deleteable"`
	Fields       []fieldDescribe `json:"fields"`
}

// ListObjects returns all objects the current user can read (for navigation).
func (h *DescribeHandler) ListObjects(c *gin.Context) {
	uc, ok := security.UserFromContext(c.Request.Context())
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("user context required"))
		return
	}

	names := h.cache.ListObjectAPINames()
	items := make([]objectNavItem, 0, len(names))

	for _, name := range names {
		objDef, ok := h.cache.GetObjectByAPIName(name)
		if !ok {
			continue
		}
		if !objDef.IsQueryable {
			continue
		}
		if err := h.olsEnforcer.CanRead(c.Request.Context(), uc.UserID, objDef.ID); err != nil {
			continue
		}
		items = append(items, objectNavItem{
			APIName:      objDef.APIName,
			Label:        objDef.Label,
			PluralLabel:  objDef.PluralLabel,
			IsCreateable: objDef.IsCreateable,
			IsQueryable:  objDef.IsQueryable,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// DescribeObject returns full object description with fields (filtered by FLS).
func (h *DescribeHandler) DescribeObject(c *gin.Context) {
	objectName := c.Param("objectName")

	uc, ok := security.UserFromContext(c.Request.Context())
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("user context required"))
		return
	}

	objDef, ok := h.cache.GetObjectByAPIName(objectName)
	if !ok {
		apperror.Respond(c, apperror.NotFound("object", objectName))
		return
	}

	if err := h.olsEnforcer.CanRead(c.Request.Context(), uc.UserID, objDef.ID); err != nil {
		apperror.Respond(c, err)
		return
	}

	// Build system fields
	systemFields := buildSystemFieldDescriptions()

	// Build user-defined fields, filtered by FLS
	userFields := h.buildUserFields(c, uc.UserID, objDef)

	allFields := append(systemFields, userFields...)

	desc := objectDescribe{
		APIName:      objDef.APIName,
		Label:        objDef.Label,
		PluralLabel:  objDef.PluralLabel,
		IsCreateable: objDef.IsCreateable,
		IsUpdateable: objDef.IsUpdateable,
		IsDeleteable: objDef.IsDeleteable,
		Fields:       allFields,
	}

	c.JSON(http.StatusOK, gin.H{"data": desc})
}

func buildSystemFieldDescriptions() []fieldDescribe {
	return []fieldDescribe{
		{APIName: "Id", Label: "ID", FieldType: "text", IsReadOnly: true, IsSystemField: true, SortOrder: -6},
		{APIName: "OwnerId", Label: "Владелец", FieldType: "reference", IsReadOnly: false, IsSystemField: true, SortOrder: -5},
		{APIName: "CreatedAt", Label: "Дата создания", FieldType: "datetime", IsReadOnly: true, IsSystemField: true, SortOrder: -4},
		{APIName: "UpdatedAt", Label: "Дата обновления", FieldType: "datetime", IsReadOnly: true, IsSystemField: true, SortOrder: -3},
		{APIName: "CreatedById", Label: "Кем создано", FieldType: "reference", IsReadOnly: true, IsSystemField: true, SortOrder: -2},
		{APIName: "UpdatedById", Label: "Кем обновлено", FieldType: "reference", IsReadOnly: true, IsSystemField: true, SortOrder: -1},
	}
}

func (h *DescribeHandler) buildUserFields(c *gin.Context, userID uuid.UUID, objDef metadata.ObjectDefinition) []fieldDescribe {
	fields := h.cache.GetFieldsByObjectID(objDef.ID)
	result := make([]fieldDescribe, 0, len(fields))

	for _, f := range fields {
		if err := h.flsEnforcer.CanReadField(c.Request.Context(), userID, f.ID); err != nil {
			continue
		}

		var subtype *string
		if f.FieldSubtype != nil {
			s := string(*f.FieldSubtype)
			subtype = &s
		}

		fd := fieldDescribe{
			APIName:       f.APIName,
			Label:         f.Label,
			FieldType:     string(f.FieldType),
			FieldSubtype:  subtype,
			IsRequired:    f.IsRequired,
			IsReadOnly:    f.IsSystemField,
			IsSystemField: f.IsSystemField,
			SortOrder:     f.SortOrder,
			Config: fieldConfigDescribe{
				MaxLength:    f.Config.MaxLength,
				Precision:    f.Config.Precision,
				Scale:        f.Config.Scale,
				DefaultValue: f.Config.DefaultValue,
				Values:       f.Config.Values,
			},
		}
		result = append(result, fd)
	}

	return result
}
