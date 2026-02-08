package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata/ddl"
)

type fieldServiceImpl struct {
	txBeginner      TxBeginner
	objectRepo      ObjectRepository
	fieldRepo       FieldRepository
	polymorphicRepo PolymorphicTargetRepository
	ddlExec         DDLExecutor
	cache           CacheInvalidator
}

// NewFieldService creates a new FieldService.
func NewFieldService(
	txBeginner TxBeginner,
	objectRepo ObjectRepository,
	fieldRepo FieldRepository,
	polymorphicRepo PolymorphicTargetRepository,
	ddlExec DDLExecutor,
	cache CacheInvalidator,
) FieldService {
	return &fieldServiceImpl{
		txBeginner:      txBeginner,
		objectRepo:      objectRepo,
		fieldRepo:       fieldRepo,
		polymorphicRepo: polymorphicRepo,
		ddlExec:         ddlExec,
		cache:           cache,
	}
}

func (s *fieldServiceImpl) Create(ctx context.Context, input CreateFieldInput) (*FieldDefinition, error) {
	field := &FieldDefinition{
		ObjectID:           input.ObjectID,
		APIName:            input.APIName,
		Label:              input.Label,
		Description:        input.Description,
		HelpText:           input.HelpText,
		FieldType:          input.FieldType,
		FieldSubtype:       input.FieldSubtype,
		ReferencedObjectID: input.ReferencedObjectID,
		IsRequired:         input.IsRequired,
		IsUnique:           input.IsUnique,
		Config:             input.Config,
		IsCustom:           input.IsCustom,
		SortOrder:          input.SortOrder,
	}

	if err := ValidateFieldDefinition(field); err != nil {
		return nil, fmt.Errorf("fieldService.Create: %w", err)
	}

	obj, err := s.objectRepo.GetByID(ctx, input.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("fieldService.Create: get object: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("fieldService.Create: %w",
			apperror.NotFound("ObjectDefinition", input.ObjectID.String()))
	}

	if input.IsCustom && !obj.IsCustomFieldsAllowed {
		return nil, fmt.Errorf("fieldService.Create: %w",
			apperror.Forbidden("custom fields are not allowed on this object"))
	}

	existing, _ := s.fieldRepo.GetByObjectAndName(ctx, input.ObjectID, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("fieldService.Create: %w",
			apperror.Conflict(fmt.Sprintf("field '%s' already exists on this object", input.APIName)))
	}

	if input.FieldType == FieldTypeReference && input.FieldSubtype != nil && *input.FieldSubtype == SubtypeComposition {
		allFields, err := s.fieldRepo.ListAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("fieldService.Create: list fields for composition check: %w", err)
		}
		checker := NewCompositionChecker(allFields)
		allObjects, err := s.loadObjectMap(ctx)
		if err != nil {
			return nil, fmt.Errorf("fieldService.Create: load objects: %w", err)
		}
		if err := checker.ValidateNewComposition(*field, allObjects); err != nil {
			return nil, fmt.Errorf("fieldService.Create: %w", err)
		}
	}

	if input.ReferencedObjectID != nil {
		refObj, err := s.objectRepo.GetByID(ctx, *input.ReferencedObjectID)
		if err != nil {
			return nil, fmt.Errorf("fieldService.Create: get referenced object: %w", err)
		}
		if refObj == nil {
			return nil, fmt.Errorf("fieldService.Create: %w",
				apperror.NotFound("referenced ObjectDefinition", input.ReferencedObjectID.String()))
		}
	}

	var result *FieldDefinition
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.fieldRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("fieldService.Create: insert metadata: %w", err)
		}

		referencedTableName := ""
		if input.ReferencedObjectID != nil {
			refObj, _ := s.objectRepo.GetByID(ctx, *input.ReferencedObjectID)
			if refObj != nil {
				referencedTableName = refObj.TableName
			}
		}

		ddlStmts, err := ddl.AddColumn(obj.TableName, ToFieldInfo(*created), referencedTableName)
		if err != nil {
			return fmt.Errorf("fieldService.Create: generate DDL: %w", err)
		}

		if err := s.ddlExec.ExecInTx(ctx, tx, ddlStmts); err != nil {
			return fmt.Errorf("fieldService.Create: execute DDL: %w", err)
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.Invalidate(ctx); err != nil {
			return nil, fmt.Errorf("fieldService.Create: cache invalidate: %w", err)
		}
	}

	return result, nil
}

func (s *fieldServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*FieldDefinition, error) {
	field, err := s.fieldRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fieldService.GetByID: %w", err)
	}
	if field == nil {
		return nil, fmt.Errorf("fieldService.GetByID: %w",
			apperror.NotFound("FieldDefinition", id.String()))
	}
	return field, nil
}

func (s *fieldServiceImpl) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldDefinition, error) {
	obj, err := s.objectRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("fieldService.ListByObjectID: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("fieldService.ListByObjectID: %w",
			apperror.NotFound("ObjectDefinition", objectID.String()))
	}

	fields, err := s.fieldRepo.ListByObjectID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("fieldService.ListByObjectID: %w", err)
	}

	return fields, nil
}

func (s *fieldServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateFieldInput) (*FieldDefinition, error) {
	existing, err := s.fieldRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fieldService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("fieldService.Update: %w",
			apperror.NotFound("FieldDefinition", id.String()))
	}

	if existing.IsSystemField {
		return nil, fmt.Errorf("fieldService.Update: %w",
			apperror.Forbidden("cannot update system field"))
	}

	if existing.IsPlatformManaged {
		return nil, fmt.Errorf("fieldService.Update: %w",
			apperror.Forbidden("cannot update platform-managed field"))
	}

	if input.Label == "" {
		return nil, fmt.Errorf("fieldService.Update: %w",
			apperror.Validation("label is required"))
	}

	var result *FieldDefinition
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.fieldRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("fieldService.Update: %w", err)
		}

		obj, err := s.objectRepo.GetByID(ctx, existing.ObjectID)
		if err != nil {
			return fmt.Errorf("fieldService.Update: get object: %w", err)
		}

		var ddlStmts []string
		if input.IsRequired != existing.IsRequired {
			if input.IsRequired {
				ddlStmts = append(ddlStmts, ddl.AlterColumnSetNotNull(obj.TableName, existing.APIName))
			} else {
				ddlStmts = append(ddlStmts, ddl.AlterColumnDropNotNull(obj.TableName, existing.APIName))
			}
		}

		if len(ddlStmts) > 0 {
			if err := s.ddlExec.ExecInTx(ctx, tx, ddlStmts); err != nil {
				return fmt.Errorf("fieldService.Update: execute DDL: %w", err)
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
			return nil, fmt.Errorf("fieldService.Update: cache invalidate: %w", err)
		}
	}

	return result, nil
}

func (s *fieldServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.fieldRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fieldService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("fieldService.Delete: %w",
			apperror.NotFound("FieldDefinition", id.String()))
	}

	if existing.IsSystemField {
		return fmt.Errorf("fieldService.Delete: %w",
			apperror.Forbidden("cannot delete system field"))
	}

	if existing.IsPlatformManaged {
		return fmt.Errorf("fieldService.Delete: %w",
			apperror.Forbidden("cannot delete platform-managed field"))
	}

	obj, err := s.objectRepo.GetByID(ctx, existing.ObjectID)
	if err != nil {
		return fmt.Errorf("fieldService.Delete: get object: %w", err)
	}

	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		ddlStmts := ddl.DropColumn(obj.TableName, ToFieldInfo(*existing))
		if err := s.ddlExec.ExecInTx(ctx, tx, ddlStmts); err != nil {
			return fmt.Errorf("fieldService.Delete: DDL DROP COLUMN: %w", err)
		}

		if err := s.fieldRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("fieldService.Delete: delete metadata: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.Invalidate(ctx); err != nil {
			return fmt.Errorf("fieldService.Delete: cache invalidate: %w", err)
		}
	}

	return nil
}

func (s *fieldServiceImpl) loadObjectMap(ctx context.Context) (map[uuid.UUID]ObjectDefinition, error) {
	objects, err := s.objectRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]ObjectDefinition, len(objects))
	for _, obj := range objects {
		result[obj.ID] = obj
	}
	return result, nil
}
