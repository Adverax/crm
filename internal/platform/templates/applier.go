package templates

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
)

// Applier creates objects and fields from a template using existing services.
type Applier struct {
	objectService     metadata.ObjectService
	fieldService      metadata.FieldService
	objectRepo        metadata.ObjectRepository
	permissionService security.PermissionService
	metadataCache     metadata.CacheInvalidator
}

// NewApplier creates a new template applier.
func NewApplier(
	objectService metadata.ObjectService,
	fieldService metadata.FieldService,
	objectRepo metadata.ObjectRepository,
	permissionService security.PermissionService,
	metadataCache metadata.CacheInvalidator,
) *Applier {
	return &Applier{
		objectService:     objectService,
		fieldService:      fieldService,
		objectRepo:        objectRepo,
		permissionService: permissionService,
		metadataCache:     metadataCache,
	}
}

// Apply creates all objects and fields defined in the template.
func (a *Applier) Apply(ctx context.Context, tmpl Template) error {
	count, err := a.objectRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("applier.Apply: count objects: %w", err)
	}
	if count > 0 {
		return apperror.Conflict("cannot apply template: objects already exist")
	}

	// Pass 1: create all objects, collect apiName → UUID mapping.
	objectMap := make(map[string]uuid.UUID, len(tmpl.Objects))
	for _, obj := range tmpl.Objects {
		created, createErr := a.objectService.Create(ctx, metadata.CreateObjectInput{
			APIName:               obj.APIName,
			Label:                 obj.Label,
			PluralLabel:           obj.PluralLabel,
			Description:           obj.Description,
			ObjectType:            metadata.ObjectTypeStandard,
			Visibility:            obj.Visibility,
			IsVisibleInSetup:      true,
			IsCustomFieldsAllowed: obj.IsCustomFieldsAllowed,
			IsDeleteableObject:    false,
			IsCreateable:          obj.IsCreateable,
			IsUpdateable:          obj.IsUpdateable,
			IsDeleteable:          obj.IsDeleteable,
			IsQueryable:           obj.IsQueryable,
			IsSearchable:          obj.IsSearchable,
			HasActivities:         obj.HasActivities,
			HasNotes:              obj.HasNotes,
			HasHistoryTracking:    obj.HasHistoryTracking,
			HasSharingRules:       obj.HasSharingRules,
		})
		if createErr != nil {
			return fmt.Errorf("applier.Apply: create object %s: %w", obj.APIName, createErr)
		}
		objectMap[obj.APIName] = created.ID
	}

	// Pass 2: create all fields, resolving reference apiName → UUID.
	for _, f := range tmpl.Fields {
		objectID, ok := objectMap[f.ObjectAPIName]
		if !ok {
			return fmt.Errorf("applier.Apply: object %s not found for field %s", f.ObjectAPIName, f.APIName)
		}

		input := metadata.CreateFieldInput{
			ObjectID:     objectID,
			APIName:      f.APIName,
			Label:        f.Label,
			Description:  f.Description,
			FieldType:    f.FieldType,
			FieldSubtype: f.FieldSubtype,
			IsRequired:   f.IsRequired,
			IsUnique:     f.IsUnique,
			Config:       f.Config,
			SortOrder:    f.SortOrder,
		}

		if f.ReferencedObjectAPIName != "" {
			refID, refOK := objectMap[f.ReferencedObjectAPIName]
			if !refOK {
				return fmt.Errorf("applier.Apply: referenced object %s not found for field %s.%s", f.ReferencedObjectAPIName, f.ObjectAPIName, f.APIName)
			}
			input.ReferencedObjectID = &refID
		}

		if _, createErr := a.fieldService.Create(ctx, input); createErr != nil {
			return fmt.Errorf("applier.Apply: create field %s.%s: %w", f.ObjectAPIName, f.APIName, createErr)
		}
	}

	// Grant full OLS on all new objects to the System Admin base PS.
	adminPSID := security.SystemAdminBasePermissionSetID
	for _, objectID := range objectMap {
		if _, permErr := a.permissionService.SetObjectPermission(ctx, adminPSID, security.SetObjectPermissionInput{
			ObjectID:    objectID,
			Permissions: security.OLSAll,
		}); permErr != nil {
			return fmt.Errorf("applier.Apply: set OLS: %w", permErr)
		}
	}

	// Invalidate metadata cache so fields are available for FLS.
	if err := a.metadataCache.Invalidate(ctx); err != nil {
		return fmt.Errorf("applier.Apply: invalidate cache: %w", err)
	}

	return nil
}
