package dml

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/fls"
	"github.com/adverax/crm/internal/platform/security/ols"
)

// MetadataAdapter bridges metadata.MetadataCache → engine.MetadataProvider.
type MetadataAdapter struct {
	cache *metadata.MetadataCache
}

// NewMetadataAdapter creates a new DML MetadataAdapter.
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
	return a.buildObjectMeta(objDef, fields), nil
}

func (a *MetadataAdapter) buildObjectMeta(objDef metadata.ObjectDefinition, fields []metadata.FieldDefinition) *engine.ObjectMeta {
	b := engine.NewObjectMeta(objDef.APIName, objDef.TableName)

	// System fields.
	b.FieldFull(&engine.FieldMeta{
		Name: "Id", Column: "id", Type: engine.FieldTypeID, ReadOnly: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "OwnerId", Column: "owner_id", Type: engine.FieldTypeID, Nullable: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "CreatedAt", Column: "created_at", Type: engine.FieldTypeDateTime, ReadOnly: true, HasDefault: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "UpdatedAt", Column: "updated_at", Type: engine.FieldTypeDateTime, ReadOnly: true, HasDefault: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "CreatedById", Column: "created_by_id", Type: engine.FieldTypeID, Nullable: true, HasDefault: true,
	})
	b.FieldFull(&engine.FieldMeta{
		Name: "UpdatedById", Column: "updated_by_id", Type: engine.FieldTypeID, Nullable: true, HasDefault: true,
	})

	// User-defined fields.
	for _, f := range fields {
		b.FieldFull(a.convertFieldMeta(f))
	}

	return b.Build()
}

func (a *MetadataAdapter) convertFieldMeta(f metadata.FieldDefinition) *engine.FieldMeta {
	return &engine.FieldMeta{
		Name:         f.APIName,
		Column:       strings.ToLower(f.APIName),
		Type:         mapFieldType(f.FieldType, f.FieldSubtype),
		Nullable:     !f.IsRequired,
		Required:     f.IsRequired,
		ReadOnly:     f.IsSystemField,
		HasDefault:   f.Config.DefaultValue != nil || f.Config.DefaultExpr != nil,
		IsExternalId: false,
		IsUnique:     f.IsUnique,
		DefaultValue: f.Config.DefaultValue,
		DefaultExpr:  f.Config.DefaultExpr,
		DefaultOn:    f.Config.DefaultOn,
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

// systemWriteFieldNames lists fields that bypass FLS write checks.
var systemWriteFieldNames = map[string]bool{
	"OwnerId":     true,
	"CreatedById": true,
	"UpdatedById": true,
}

// WriteAccessControllerAdapter bridges OLS/FLS enforcers → engine.WriteAccessController.
type WriteAccessControllerAdapter struct {
	cache       *metadata.MetadataCache
	olsEnforcer ols.Enforcer
	flsEnforcer fls.Enforcer
}

// NewWriteAccessControllerAdapter creates a new adapter.
func NewWriteAccessControllerAdapter(
	cache *metadata.MetadataCache,
	olsEnforcer ols.Enforcer,
	flsEnforcer fls.Enforcer,
) *WriteAccessControllerAdapter {
	return &WriteAccessControllerAdapter{
		cache:       cache,
		olsEnforcer: olsEnforcer,
		flsEnforcer: flsEnforcer,
	}
}

// CanWriteObject implements engine.WriteAccessController.
func (ac *WriteAccessControllerAdapter) CanWriteObject(ctx context.Context, object string, op engine.Operation) error {
	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return engine.NewWriteAccessError(object, op)
	}

	objDef, ok := ac.cache.GetObjectByAPIName(object)
	if !ok {
		return nil // unknown object → will be caught by validator
	}

	var err error
	switch op {
	case engine.OperationInsert:
		err = ac.olsEnforcer.CanCreate(ctx, uc.UserID, objDef.ID)
	case engine.OperationUpdate:
		err = ac.olsEnforcer.CanUpdate(ctx, uc.UserID, objDef.ID)
	case engine.OperationDelete:
		err = ac.olsEnforcer.CanDelete(ctx, uc.UserID, objDef.ID)
	case engine.OperationUpsert:
		// Upsert requires both create and update.
		if err = ac.olsEnforcer.CanCreate(ctx, uc.UserID, objDef.ID); err == nil {
			err = ac.olsEnforcer.CanUpdate(ctx, uc.UserID, objDef.ID)
		}
	}

	if err != nil {
		return engine.NewWriteAccessError(object, op)
	}
	return nil
}

// CheckWritableFields implements engine.WriteAccessController.
func (ac *WriteAccessControllerAdapter) CheckWritableFields(ctx context.Context, object string, fields []string) error {
	uc, ok := security.UserFromContext(ctx)
	if !ok {
		if len(fields) > 0 {
			return engine.NewFieldWriteAccessError(object, fields[0])
		}
		return nil
	}

	objDef, ok := ac.cache.GetObjectByAPIName(object)
	if !ok {
		return nil
	}

	for _, fieldName := range fields {
		if systemWriteFieldNames[fieldName] {
			continue
		}
		fieldID, ok := findFieldID(ac.cache, objDef.ID, fieldName)
		if !ok {
			continue
		}
		if err := ac.flsEnforcer.CanWriteField(ctx, uc.UserID, fieldID); err != nil {
			return engine.NewFieldWriteAccessError(object, fieldName)
		}
	}
	return nil
}

func findFieldID(cache *metadata.MetadataCache, objectID uuid.UUID, fieldAPIName string) (uuid.UUID, bool) {
	fields := cache.GetFieldsByObjectID(objectID)
	for _, f := range fields {
		if f.APIName == fieldAPIName {
			return f.ID, true
		}
	}
	return uuid.Nil, false
}
