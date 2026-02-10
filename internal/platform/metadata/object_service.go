package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata/ddl"
)

type objectServiceImpl struct {
	txBeginner TxBeginner
	objectRepo ObjectRepository
	fieldRepo  FieldRepository
	ddlExec    DDLExecutor
	cache      CacheInvalidator
}

// NewObjectService creates a new ObjectService.
func NewObjectService(
	txBeginner TxBeginner,
	objectRepo ObjectRepository,
	fieldRepo FieldRepository,
	ddlExec DDLExecutor,
	cache CacheInvalidator,
) ObjectService {
	return &objectServiceImpl{
		txBeginner: txBeginner,
		objectRepo: objectRepo,
		fieldRepo:  fieldRepo,
		ddlExec:    ddlExec,
		cache:      cache,
	}
}

func (s *objectServiceImpl) Create(ctx context.Context, input CreateObjectInput) (*ObjectDefinition, error) {
	tableName := GenerateTableName(input.APIName)

	if input.Visibility == "" {
		input.Visibility = VisibilityPrivate
	}

	obj := &ObjectDefinition{
		APIName:               input.APIName,
		Label:                 input.Label,
		PluralLabel:           input.PluralLabel,
		Description:           input.Description,
		TableName:             tableName,
		ObjectType:            input.ObjectType,
		IsVisibleInSetup:      input.IsVisibleInSetup,
		IsCustomFieldsAllowed: input.IsCustomFieldsAllowed,
		IsDeleteableObject:    input.IsDeleteableObject,
		IsCreateable:          input.IsCreateable,
		IsUpdateable:          input.IsUpdateable,
		IsDeleteable:          input.IsDeleteable,
		IsQueryable:           input.IsQueryable,
		IsSearchable:          input.IsSearchable,
		HasActivities:         input.HasActivities,
		HasNotes:              input.HasNotes,
		HasHistoryTracking:    input.HasHistoryTracking,
		HasSharingRules:       input.HasSharingRules,
		Visibility:            input.Visibility,
	}

	if err := ValidateObjectDefinition(obj); err != nil {
		return nil, fmt.Errorf("objectService.Create: %w", err)
	}

	if err := ValidateVisibility(input.Visibility); err != nil {
		return nil, fmt.Errorf("objectService.Create: %w", err)
	}

	existing, _ := s.objectRepo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("objectService.Create: %w",
			apperror.Conflict(fmt.Sprintf("object with api_name '%s' already exists", input.APIName)))
	}

	var result *ObjectDefinition
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.objectRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("objectService.Create: insert metadata: %w", err)
		}

		createTableDDL := ddl.CreateObjectTable(tableName)
		if err := s.ddlExec.ExecInTx(ctx, tx, []string{createTableDDL}); err != nil {
			return fmt.Errorf("objectService.Create: DDL CREATE TABLE: %w", err)
		}

		// Create share table if OWD is not public_read_write
		if input.Visibility != VisibilityPublicReadWrite {
			shareTableDDL := ddl.CreateShareTable(tableName)
			if err := s.ddlExec.ExecInTx(ctx, tx, []string{shareTableDDL}); err != nil {
				return fmt.Errorf("objectService.Create: DDL CREATE SHARE TABLE: %w", err)
			}
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.Invalidate(ctx); err != nil {
			return nil, fmt.Errorf("objectService.Create: cache invalidate: %w", err)
		}
	}

	return result, nil
}

func (s *objectServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*ObjectDefinition, error) {
	obj, err := s.objectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("objectService.GetByID: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("objectService.GetByID: %w",
			apperror.NotFound("ObjectDefinition", id.String()))
	}
	return obj, nil
}

func (s *objectServiceImpl) List(ctx context.Context, filter ObjectFilter) ([]ObjectDefinition, int64, error) {
	if filter.PerPage <= 0 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PerPage

	objects, err := s.objectRepo.List(ctx, filter.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("objectService.List: %w", err)
	}

	total, err := s.objectRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("objectService.List: count: %w", err)
	}

	return objects, total, nil
}

func (s *objectServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateObjectInput) (*ObjectDefinition, error) {
	existing, err := s.objectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("objectService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("objectService.Update: %w",
			apperror.NotFound("ObjectDefinition", id.String()))
	}

	if existing.IsPlatformManaged {
		return nil, fmt.Errorf("objectService.Update: %w",
			apperror.Forbidden("cannot update platform-managed object"))
	}

	if input.Label == "" {
		return nil, fmt.Errorf("objectService.Update: %w",
			apperror.Validation("label is required"))
	}
	if input.PluralLabel == "" {
		return nil, fmt.Errorf("objectService.Update: %w",
			apperror.Validation("plural_label is required"))
	}

	if input.Visibility == "" {
		input.Visibility = existing.Visibility
	}
	if err := ValidateVisibility(input.Visibility); err != nil {
		return nil, fmt.Errorf("objectService.Update: %w", err)
	}

	visibilityChanged := existing.Visibility != input.Visibility
	wasPublicRW := existing.Visibility == VisibilityPublicReadWrite
	willBePublicRW := input.Visibility == VisibilityPublicReadWrite

	var result *ObjectDefinition
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.objectRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("objectService.Update: %w", err)
		}

		// Handle share table lifecycle on visibility change
		if visibilityChanged {
			if wasPublicRW && !willBePublicRW {
				// Need to create share table
				shareTableDDL := ddl.CreateShareTable(existing.TableName)
				if err := s.ddlExec.ExecInTx(ctx, tx, []string{shareTableDDL}); err != nil {
					return fmt.Errorf("objectService.Update: DDL CREATE SHARE TABLE: %w", err)
				}
			} else if !wasPublicRW && willBePublicRW {
				// Need to drop share table
				dropShareDDL := ddl.DropShareTable(existing.TableName)
				if err := s.ddlExec.ExecInTx(ctx, tx, []string{dropShareDDL}); err != nil {
					return fmt.Errorf("objectService.Update: DDL DROP SHARE TABLE: %w", err)
				}
			}
		}

		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.Invalidate(ctx); err != nil {
			return nil, fmt.Errorf("objectService.Update: cache invalidate: %w", err)
		}
	}

	return result, nil
}

func (s *objectServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.objectRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("objectService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("objectService.Delete: %w",
			apperror.NotFound("ObjectDefinition", id.String()))
	}

	if !existing.IsDeleteableObject {
		return fmt.Errorf("objectService.Delete: %w",
			apperror.Forbidden("this object cannot be deleted"))
	}

	if existing.IsPlatformManaged {
		return fmt.Errorf("objectService.Delete: %w",
			apperror.Forbidden("cannot delete platform-managed object"))
	}

	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		// Drop share table first (if it exists)
		dropShareDDL := ddl.DropShareTable(existing.TableName)
		if err := s.ddlExec.ExecInTx(ctx, tx, []string{dropShareDDL}); err != nil {
			return fmt.Errorf("objectService.Delete: DDL DROP SHARE TABLE: %w", err)
		}

		dropTableDDL := ddl.DropObjectTable(existing.TableName)
		if err := s.ddlExec.ExecInTx(ctx, tx, []string{dropTableDDL}); err != nil {
			return fmt.Errorf("objectService.Delete: DDL DROP TABLE: %w", err)
		}

		if err := s.objectRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("objectService.Delete: delete metadata: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.Invalidate(ctx); err != nil {
			return fmt.Errorf("objectService.Delete: cache invalidate: %w", err)
		}
	}

	return nil
}
