package soql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/fls"
	"github.com/adverax/crm/internal/platform/security/ols"
	"github.com/adverax/crm/internal/platform/soql/engine"
)

// MetadataAdapter bridges metadata.MetadataCache → engine.MetadataProvider.
type MetadataAdapter struct {
	cache *metadata.MetadataCache
}

// NewMetadataAdapter creates a new MetadataAdapter.
func NewMetadataAdapter(cache *metadata.MetadataCache) *MetadataAdapter {
	return &MetadataAdapter{cache: cache}
}

// GetObject implements engine.MetadataProvider.
func (a *MetadataAdapter) GetObject(_ context.Context, name string) (*engine.ObjectMeta, error) {
	objDef, ok := a.cache.GetObjectByAPIName(name)
	if !ok {
		return nil, nil
	}

	fields := a.cache.GetFieldsByObjectID(objDef.ID)
	forwardRels := a.cache.GetForwardRelationships(objDef.ID)
	reverseRels := a.cache.GetReverseRelationships(objDef.ID)

	return a.buildObjectMeta(objDef, fields, forwardRels, reverseRels), nil
}

// ListObjects implements engine.MetadataProvider.
func (a *MetadataAdapter) ListObjects(_ context.Context) ([]string, error) {
	return a.cache.ListObjectAPINames(), nil
}

func (a *MetadataAdapter) buildObjectMeta(
	objDef metadata.ObjectDefinition,
	fields []metadata.FieldDefinition,
	forwardRels []metadata.RelationshipInfo,
	reverseRels []metadata.RelationshipInfo,
) *engine.ObjectMeta {
	b := engine.NewObjectMeta(objDef.APIName, "public", objDef.TableName)

	// System fields always present on every object.
	a.addSystemFields(b)

	// User-defined fields.
	for _, f := range fields {
		b.FieldFull(a.convertFieldMeta(f))
	}

	// Lookups (child → parent, via forward relationships).
	for _, rel := range forwardRels {
		parentObj, ok := a.cache.GetObjectByID(rel.ParentObjectID)
		if !ok {
			continue
		}
		lookupName := rel.RelationshipName
		if lookupName == "" {
			lookupName = parentObj.APIName
		}
		b.Lookup(lookupName, rel.FieldAPIName, parentObj.APIName, "id")
	}

	// Relationships (parent → child, via reverse relationships).
	for _, rel := range reverseRels {
		childObj, ok := a.cache.GetObjectByID(rel.ChildObjectID)
		if !ok {
			continue
		}
		relName := rel.RelationshipName
		if relName == "" {
			relName = childObj.APIName + "s"
		}
		b.Relationship(relName, childObj.APIName, rel.FieldAPIName, "id")
	}

	return b.Build()
}

func (a *MetadataAdapter) addSystemFields(b *engine.ObjectMetaBuilder) {
	b.FieldFull(&engine.FieldMeta{
		Name: "Id", Column: "id", Type: engine.FieldTypeID,
		Filterable: true, Sortable: true, Groupable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "OwnerId", Column: "owner_id", Type: engine.FieldTypeID,
		Nullable: true, Filterable: true, Sortable: true, Groupable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "CreatedAt", Column: "created_at", Type: engine.FieldTypeDateTime,
		Filterable: true, Sortable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "UpdatedAt", Column: "updated_at", Type: engine.FieldTypeDateTime,
		Filterable: true, Sortable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "CreatedById", Column: "created_by_id", Type: engine.FieldTypeID,
		Nullable: true, Filterable: true, Sortable: true, Groupable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "UpdatedById", Column: "updated_by_id", Type: engine.FieldTypeID,
		Nullable: true, Filterable: true, Sortable: true, Groupable: true,
	})
}

func (a *MetadataAdapter) convertFieldMeta(f metadata.FieldDefinition) *engine.FieldMeta {
	ft := mapFieldType(f.FieldType, f.FieldSubtype)
	return &engine.FieldMeta{
		Name:         f.APIName,
		Column:       strings.ToLower(f.APIName),
		Type:         ft,
		Nullable:     !f.IsRequired,
		Filterable:   isFilterable(ft),
		Sortable:     isSortable(ft),
		Groupable:    isGroupable(ft),
		Aggregatable: ft == engine.FieldTypeInteger || ft == engine.FieldTypeFloat,
	}
}

func mapFieldType(ft metadata.FieldType, sub *metadata.FieldSubtype) engine.FieldType {
	switch ft {
	case metadata.FieldTypeText:
		return engine.FieldTypeString
	case metadata.FieldTypeNumber:
		if sub != nil && (*sub == metadata.SubtypeDecimal || *sub == metadata.SubtypeCurrency || *sub == metadata.SubtypePercent) {
			return engine.FieldTypeFloat
		}
		return engine.FieldTypeInteger
	case metadata.FieldTypeBoolean:
		return engine.FieldTypeBoolean
	case metadata.FieldTypeDatetime:
		if sub != nil && *sub == metadata.SubtypeDate {
			return engine.FieldTypeDate
		}
		return engine.FieldTypeDateTime
	case metadata.FieldTypePicklist:
		return engine.FieldTypeString
	case metadata.FieldTypeReference:
		return engine.FieldTypeID
	default:
		return engine.FieldTypeString
	}
}

func isFilterable(ft engine.FieldType) bool { return true }
func isSortable(ft engine.FieldType) bool {
	return ft != engine.FieldTypeObject && ft != engine.FieldTypeArray
}
func isGroupable(ft engine.FieldType) bool { return isSortable(ft) }

// systemFieldNames lists fields that always pass access checks.
var systemFieldNames = map[string]bool{
	"Id": true, "OwnerId": true, "CreatedAt": true, "UpdatedAt": true,
	"CreatedById": true, "UpdatedById": true,
}

// AccessControllerAdapter bridges OLS/FLS enforcers → engine.AccessController.
type AccessControllerAdapter struct {
	cache       *metadata.MetadataCache
	olsEnforcer ols.Enforcer
	flsEnforcer fls.Enforcer
}

// NewAccessControllerAdapter creates a new adapter.
func NewAccessControllerAdapter(
	cache *metadata.MetadataCache,
	olsEnforcer ols.Enforcer,
	flsEnforcer fls.Enforcer,
) *AccessControllerAdapter {
	return &AccessControllerAdapter{
		cache:       cache,
		olsEnforcer: olsEnforcer,
		flsEnforcer: flsEnforcer,
	}
}

// CanAccessObject implements engine.AccessController.
func (ac *AccessControllerAdapter) CanAccessObject(ctx context.Context, object string) error {
	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return engine.NewAccessError(object)
	}

	objDef, ok := ac.cache.GetObjectByAPIName(object)
	if !ok {
		return nil // unknown object → will be caught by validator
	}

	if err := ac.olsEnforcer.CanRead(ctx, uc.UserID, objDef.ID); err != nil {
		return engine.NewAccessError(object)
	}
	return nil
}

// CanAccessField implements engine.AccessController.
func (ac *AccessControllerAdapter) CanAccessField(ctx context.Context, object, field string) error {
	if systemFieldNames[field] {
		return nil
	}

	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return engine.NewFieldAccessError(object, field)
	}

	objDef, ok := ac.cache.GetObjectByAPIName(object)
	if !ok {
		return nil
	}

	fieldID, ok := ac.findFieldID(objDef.ID, field)
	if !ok {
		return nil // unknown field → will be caught by validator
	}

	if err := ac.flsEnforcer.CanReadField(ctx, uc.UserID, fieldID); err != nil {
		return engine.NewFieldAccessError(object, field)
	}
	return nil
}

func (ac *AccessControllerAdapter) findFieldID(objectID uuid.UUID, fieldAPIName string) (uuid.UUID, bool) {
	fields := ac.cache.GetFieldsByObjectID(objectID)
	for _, f := range fields {
		if f.APIName == fieldAPIName {
			return f.ID, true
		}
	}
	return uuid.Nil, false
}

// resolveObjectID resolves an API name to a UUID via the metadata cache.
func resolveObjectID(cache *metadata.MetadataCache, apiName string) (uuid.UUID, error) {
	objDef, ok := cache.GetObjectByAPIName(apiName)
	if !ok {
		return uuid.Nil, fmt.Errorf("object %q not found", apiName)
	}
	return objDef.ID, nil
}
