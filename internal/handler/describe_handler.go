package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	celutil "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/fls"
	"github.com/adverax/crm/internal/platform/security/ols"
	"github.com/adverax/crm/internal/platform/soql/engine"
)

// DMLTargetExtractor extracts target objects and modified fields from DML statements.
type DMLTargetExtractor interface {
	ExtractTargets(statements []string) []dml.DMLTargetInfo
}

// DescribeHandler exposes public metadata for frontend consumption.
type DescribeHandler struct {
	cache        metadata.MetadataReader
	olsEnforcer  ols.Enforcer
	flsEnforcer  fls.Enforcer
	dmlExtractor DMLTargetExtractor
}

// NewDescribeHandler creates a new DescribeHandler.
func NewDescribeHandler(
	cache metadata.MetadataReader,
	olsEnforcer ols.Enforcer,
	flsEnforcer fls.Enforcer,
	dmlExtractor DMLTargetExtractor,
) *DescribeHandler {
	return &DescribeHandler{
		cache:        cache,
		olsEnforcer:  olsEnforcer,
		flsEnforcer:  flsEnforcer,
		dmlExtractor: dmlExtractor,
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
	Form         *formDescribe   `json:"form,omitempty"`
}

type formDescribe struct {
	Sections          []formSection                    `json:"sections"`
	HighlightFields   []string                         `json:"highlight_fields"`
	Actions           []formAction                     `json:"actions"`
	RelatedLists      []formRelatedList                `json:"related_lists"`
	ListFields        []string                         `json:"list_fields"`
	ListDefaultSort   string                           `json:"list_default_sort"`
	Root              *metadata.LayoutComponent        `json:"root,omitempty"`
	ListConfig        *metadata.ListConfig             `json:"list_config,omitempty"`
	FieldPresentation map[string]formFieldPresentation `json:"field_presentation,omitempty"`
	Queries           []formQuery                      `json:"queries,omitempty"`
}

type formQuery struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type formSection struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	Columns        int      `json:"columns"`
	Collapsed      bool     `json:"collapsed"`
	Collapsible    bool     `json:"collapsible,omitempty"`
	VisibilityExpr string   `json:"visibility_expr,omitempty"`
	Fields         []string `json:"fields"`
}

type formFieldPresentation struct {
	ColSpan        int                 `json:"col_span,omitempty"`
	UIKind         json.RawMessage     `json:"ui_kind,omitempty"`
	RequiredExpr   string              `json:"required_expr,omitempty"`
	ReadonlyExpr   string              `json:"readonly_expr,omitempty"`
	VisibilityExpr string              `json:"visibility_expr,omitempty"`
	Reference      *metadata.RefConfig `json:"reference,omitempty"`
}

type formAction struct {
	Key             string               `json:"key"`
	Label           string               `json:"label"`
	Type            string               `json:"type"`
	Icon            string               `json:"icon"`
	VisibilityExpr  string               `json:"visibility_expr"`
	ValidationRules []formValidationRule `json:"validation_rules,omitempty"`
}

type formValidationRule struct {
	Expression     string `json:"expression"`
	ErrorMessage   string `json:"error_message"`
	ErrorCode      string `json:"error_code"`
	WhenExpression string `json:"when_expression,omitempty"`
}

type formRelatedList struct {
	Object string   `json:"object"`
	Label  string   `json:"label"`
	Fields []string `json:"fields"`
	Sort   string   `json:"sort"`
	Limit  int      `json:"limit"`
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

	// Read layout hints from headers
	formFactor := c.GetHeader("X-Form-Factor")
	if formFactor == "" {
		formFactor = "desktop"
	}
	formMode := c.GetHeader("X-Form-Mode")
	if formMode == "" {
		formMode = "read"
	}

	// Try to resolve form via OV + Layout, fallback to auto-generated
	form := h.resolveForm(objDef, allFields, formFactor, formMode)

	desc := objectDescribe{
		APIName:      objDef.APIName,
		Label:        objDef.Label,
		PluralLabel:  objDef.PluralLabel,
		IsCreateable: objDef.IsCreateable,
		IsUpdateable: objDef.IsUpdateable,
		IsDeleteable: objDef.IsDeleteable,
		Fields:       allFields,
		Form:         form,
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

func buildFallbackForm(fields []fieldDescribe) *formDescribe {
	editableNames := make([]string, 0)
	for _, f := range fields {
		if !f.IsSystemField && !f.IsReadOnly {
			editableNames = append(editableNames, f.APIName)
		}
	}

	highlightFields := make([]string, 0, 3)
	for i, f := range fields {
		if i >= 3 {
			break
		}
		if !f.IsSystemField {
			highlightFields = append(highlightFields, f.APIName)
		}
	}

	listFields := make([]string, 0, 5)
	for _, f := range fields {
		if len(listFields) >= 5 {
			break
		}
		if !f.IsSystemField || f.APIName == "Id" {
			listFields = append(listFields, f.APIName)
		}
	}

	return &formDescribe{
		Sections: []formSection{{
			Key:     "details",
			Label:   "Details",
			Columns: 2,
			Fields:  editableNames,
		}},
		HighlightFields: highlightFields,
		Actions:         []formAction{},
		RelatedLists:    []formRelatedList{},
		ListFields:      listFields,
	}
}

// resolveForm attempts to merge OV + Layout into a Form.
// If no OV or Layout is found for this object, falls back to auto-generated form.
func (h *DescribeHandler) resolveForm(
	objDef metadata.ObjectDefinition,
	fields []fieldDescribe,
	formFactor string,
	mode string,
) *formDescribe {
	// Find OV for this object via object api_name convention (ov api_name = object api_name)
	ov, hasOV := h.cache.GetObjectViewByAPIName(objDef.APIName)
	if !hasOV {
		return buildFallbackForm(fields)
	}

	// Find Layout for this OV with fallback chain
	layout := h.resolveLayout(ov.ID, formFactor, mode)

	// Build form from OV + Layout merge
	return h.mergeOVAndLayout(ov, layout, fields)
}

// resolveLayout finds the best matching layout with fallback chain:
// 1. Exact match (form_factor + mode)
// 2. Same form_factor, any mode
// 3. desktop + same mode
// 4. desktop + read
// 5. nil (auto-generate)
func (h *DescribeHandler) resolveLayout(ovID uuid.UUID, formFactor string, mode string) *metadata.Layout {
	layouts := h.cache.GetLayoutsForOV(ovID)
	if len(layouts) == 0 {
		return nil
	}

	var sameFFAnyMode, desktopSameMode, desktopRead *metadata.Layout
	for i := range layouts {
		l := &layouts[i]
		if l.FormFactor == formFactor && l.Mode == mode {
			return l // exact match
		}
		if l.FormFactor == formFactor && sameFFAnyMode == nil {
			sameFFAnyMode = l
		}
		if l.FormFactor == "desktop" && l.Mode == mode && desktopSameMode == nil {
			desktopSameMode = l
		}
		if l.FormFactor == "desktop" && l.Mode == "read" && desktopRead == nil {
			desktopRead = l
		}
	}

	if sameFFAnyMode != nil {
		return sameFFAnyMode
	}
	if desktopSameMode != nil {
		return desktopSameMode
	}
	return desktopRead // may be nil → auto-generate
}

// mergeOVAndLayout merges OV config + Layout config into formDescribe.
func (h *DescribeHandler) mergeOVAndLayout(
	ov metadata.ObjectView,
	layout *metadata.Layout,
	fields []fieldDescribe,
) *formDescribe {
	form := buildFallbackForm(fields)

	// Apply OV sections if OV has view config with fields
	fieldNames := metadata.FieldNames(ov.Config.Read.Fields)
	if len(fieldNames) > 0 {
		form.Sections = []formSection{{
			Key:     "details",
			Label:   "Details",
			Columns: 2,
			Fields:  fieldNames,
		}}

		form.ListFields = fieldNames
		if len(form.ListFields) > 5 {
			form.ListFields = form.ListFields[:5]
		}
	}

	// Map queries to form (without SOQL for security).
	// Infer type from SOQL syntax: SELECT ROW = scalar, SELECT = list.
	if len(ov.Config.Read.Queries) > 0 {
		queries := make([]formQuery, len(ov.Config.Read.Queries))
		for i, q := range ov.Config.Read.Queries {
			qType := "list"
			if engine.IsRowQuery(q.SOQL) {
				qType = "scalar"
			}
			queries[i] = formQuery{
				Name: q.Name,
				Type: qType,
			}
		}
		form.Queries = queries
	}

	// Apply OV actions (NOT apply — server-side only)
	if len(ov.Config.Read.Actions) > 0 {
		actions := make([]formAction, len(ov.Config.Read.Actions))
		for i, a := range ov.Config.Read.Actions {
			actions[i] = formAction{
				Key:            a.Key,
				Label:          a.Label,
				Type:           a.Type,
				Icon:           a.Icon,
				VisibilityExpr: a.VisibilityExpr,
			}
			// Extract validation rules for DML actions
			if a.Apply != nil && a.Apply.Type == "dml" && h.dmlExtractor != nil {
				actions[i].ValidationRules = h.extractActionValidationRules(a.Apply.DML)
			}
		}
		form.Actions = actions
	}

	if layout == nil {
		return form
	}

	// Apply Layout root component tree
	form.Root = layout.Config.Root

	// Apply Layout section config
	if layout.Config.SectionConfig != nil {
		for i := range form.Sections {
			sc, ok := layout.Config.SectionConfig[form.Sections[i].Key]
			if !ok {
				continue
			}
			if sc.Columns > 0 {
				form.Sections[i].Columns = sc.Columns
			}
			form.Sections[i].Collapsed = sc.Collapsed
			form.Sections[i].Collapsible = sc.Collapsible
			form.Sections[i].VisibilityExpr = sc.VisibilityExpr
		}
	}

	// Apply Layout field config (with layout_ref resolution)
	if layout.Config.FieldConfig != nil {
		presentation := make(map[string]formFieldPresentation)
		for fieldName, fc := range layout.Config.FieldConfig {
			resolved := h.resolveFieldConfig(fc)
			presentation[fieldName] = resolved
		}
		form.FieldPresentation = presentation
	}

	// Apply Layout list config
	if layout.Config.ListConfig != nil {
		form.ListConfig = layout.Config.ListConfig
	}

	return form
}

// extractActionValidationRules extracts validation rules from metadata for DML target objects,
// filtered by the fields each DML statement modifies.
func (h *DescribeHandler) extractActionValidationRules(dmlStatements []string) []formValidationRule {
	targets := h.dmlExtractor.ExtractTargets(dmlStatements)
	if len(targets) == 0 {
		return nil
	}

	// Collect modified fields per object (merge if same object in multiple DMLs)
	objectFields := make(map[string]map[string]bool)
	for _, t := range targets {
		if t.Operation == "delete" || len(t.Fields) == 0 {
			continue
		}
		if objectFields[t.Object] == nil {
			objectFields[t.Object] = make(map[string]bool)
		}
		for _, f := range t.Fields {
			objectFields[t.Object][f] = true
		}
	}

	if len(objectFields) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]bool)
	var rules []formValidationRule

	for objName, dmlFields := range objectFields {
		objDef, ok := h.cache.GetObjectByAPIName(objName)
		if !ok {
			continue
		}

		vrs := h.cache.GetValidationRules(objDef.ID)
		for _, vr := range vrs {
			if !vr.IsActive || seen[vr.ID] {
				continue
			}

			ruleFields := celutil.ExtractRecordFieldRefs(vr.Expression)
			if len(ruleFields) == 0 {
				continue
			}

			// Include rule if intersection(rule_fields, dml_fields) is non-empty
			hasOverlap := false
			for _, rf := range ruleFields {
				if dmlFields[rf] {
					hasOverlap = true
					break
				}
			}
			if !hasOverlap {
				continue
			}

			seen[vr.ID] = true
			rule := formValidationRule{
				Expression:   vr.Expression,
				ErrorMessage: vr.ErrorMessage,
				ErrorCode:    vr.ErrorCode,
			}
			if vr.WhenExpression != nil {
				rule.WhenExpression = *vr.WhenExpression
			}
			rules = append(rules, rule)
		}
	}

	return rules
}

// resolveFieldConfig resolves a LayoutFieldConfig, merging shared layout if layout_ref is present.
func (h *DescribeHandler) resolveFieldConfig(fc metadata.LayoutFieldConfig) formFieldPresentation {
	result := formFieldPresentation{
		ColSpan:        fc.ColSpan,
		UIKind:         fc.UIKind,
		RequiredExpr:   fc.RequiredExpr,
		ReadonlyExpr:   fc.ReadonlyExpr,
		VisibilityExpr: fc.VisibilityExpr,
		Reference:      fc.Reference,
	}

	// Resolve layout_ref: shared layout provides base, inline config wins
	if fc.LayoutRef != "" {
		sl, ok := h.cache.GetSharedLayoutByAPIName(fc.LayoutRef)
		if ok {
			var shared metadata.LayoutFieldConfig
			if err := json.Unmarshal(sl.Config, &shared); err == nil {
				// Shared provides defaults, inline overrides
				if result.ColSpan == 0 {
					result.ColSpan = shared.ColSpan
				}
				if result.UIKind == nil {
					result.UIKind = shared.UIKind
				}
				if result.RequiredExpr == "" {
					result.RequiredExpr = shared.RequiredExpr
				}
				if result.ReadonlyExpr == "" {
					result.ReadonlyExpr = shared.ReadonlyExpr
				}
				if result.VisibilityExpr == "" {
					result.VisibilityExpr = shared.VisibilityExpr
				}
				if result.Reference == nil {
					result.Reference = shared.Reference
				}
			}
		}
	}

	return result
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
