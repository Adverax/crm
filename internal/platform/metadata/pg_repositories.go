package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgObjectRepository implements ObjectRepository using pgx.
type PgObjectRepository struct {
	pool *pgxpool.Pool
}

// NewPgObjectRepository creates a new PgObjectRepository.
func NewPgObjectRepository(pool *pgxpool.Pool) *PgObjectRepository {
	return &PgObjectRepository{pool: pool}
}

func (r *PgObjectRepository) Create(ctx context.Context, tx pgx.Tx, input CreateObjectInput) (*ObjectDefinition, error) {
	var obj ObjectDefinition
	tableName := GenerateTableName(input.APIName)
	err := tx.QueryRow(ctx, `
		INSERT INTO metadata.object_definitions (
			api_name, label, plural_label, description, table_name, object_type,
			is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, api_name, label, plural_label, description, table_name, object_type,
			is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules,
			created_at, updated_at
	`,
		input.APIName, input.Label, input.PluralLabel, input.Description, tableName, input.ObjectType,
		input.IsVisibleInSetup, input.IsCustomFieldsAllowed, input.IsDeleteableObject,
		input.IsCreateable, input.IsUpdateable, input.IsDeleteable, input.IsQueryable, input.IsSearchable,
		input.HasActivities, input.HasNotes, input.HasHistoryTracking, input.HasSharingRules,
	).Scan(
		&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description, &obj.TableName, &obj.ObjectType,
		&obj.IsPlatformManaged, &obj.IsVisibleInSetup, &obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
		&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable, &obj.IsQueryable, &obj.IsSearchable,
		&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
		&obj.CreatedAt, &obj.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgObjectRepo.Create: %w", err)
	}
	return &obj, nil
}

func (r *PgObjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*ObjectDefinition, error) {
	var obj ObjectDefinition
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, plural_label, description, table_name, object_type,
			is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules,
			created_at, updated_at
		FROM metadata.object_definitions WHERE id = $1
	`, id).Scan(
		&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description, &obj.TableName, &obj.ObjectType,
		&obj.IsPlatformManaged, &obj.IsVisibleInSetup, &obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
		&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable, &obj.IsQueryable, &obj.IsSearchable,
		&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
		&obj.CreatedAt, &obj.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgObjectRepo.GetByID: %w", err)
	}
	return &obj, nil
}

func (r *PgObjectRepository) GetByAPIName(ctx context.Context, apiName string) (*ObjectDefinition, error) {
	var obj ObjectDefinition
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, plural_label, description, table_name, object_type,
			is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules,
			created_at, updated_at
		FROM metadata.object_definitions WHERE api_name = $1
	`, apiName).Scan(
		&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description, &obj.TableName, &obj.ObjectType,
		&obj.IsPlatformManaged, &obj.IsVisibleInSetup, &obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
		&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable, &obj.IsQueryable, &obj.IsSearchable,
		&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
		&obj.CreatedAt, &obj.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgObjectRepo.GetByAPIName: %w", err)
	}
	return &obj, nil
}

func (r *PgObjectRepository) List(ctx context.Context, limit, offset int32) ([]ObjectDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, plural_label, description, table_name, object_type,
			is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules,
			created_at, updated_at
		FROM metadata.object_definitions
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgObjectRepo.List: %w", err)
	}
	defer rows.Close()

	objects := make([]ObjectDefinition, 0)
	for rows.Next() {
		var obj ObjectDefinition
		if err := rows.Scan(
			&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description, &obj.TableName, &obj.ObjectType,
			&obj.IsPlatformManaged, &obj.IsVisibleInSetup, &obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
			&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable, &obj.IsQueryable, &obj.IsSearchable,
			&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
			&obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgObjectRepo.List: scan: %w", err)
		}
		objects = append(objects, obj)
	}
	return objects, rows.Err()
}

func (r *PgObjectRepository) ListAll(ctx context.Context) ([]ObjectDefinition, error) {
	return r.List(ctx, 10000, 0)
}

func (r *PgObjectRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateObjectInput) (*ObjectDefinition, error) {
	var obj ObjectDefinition
	err := tx.QueryRow(ctx, `
		UPDATE metadata.object_definitions SET
			label = $2, plural_label = $3, description = $4,
			is_visible_in_setup = $5, is_custom_fields_allowed = $6, is_deleteable_object = $7,
			is_createable = $8, is_updateable = $9, is_deleteable = $10,
			is_queryable = $11, is_searchable = $12,
			has_activities = $13, has_notes = $14, has_history_tracking = $15, has_sharing_rules = $16,
			updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, label, plural_label, description, table_name, object_type,
			is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
			is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
			has_activities, has_notes, has_history_tracking, has_sharing_rules,
			created_at, updated_at
	`,
		id, input.Label, input.PluralLabel, input.Description,
		input.IsVisibleInSetup, input.IsCustomFieldsAllowed, input.IsDeleteableObject,
		input.IsCreateable, input.IsUpdateable, input.IsDeleteable,
		input.IsQueryable, input.IsSearchable,
		input.HasActivities, input.HasNotes, input.HasHistoryTracking, input.HasSharingRules,
	).Scan(
		&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description, &obj.TableName, &obj.ObjectType,
		&obj.IsPlatformManaged, &obj.IsVisibleInSetup, &obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
		&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable, &obj.IsQueryable, &obj.IsSearchable,
		&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
		&obj.CreatedAt, &obj.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgObjectRepo.Update: %w", err)
	}
	return &obj, nil
}

func (r *PgObjectRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM metadata.object_definitions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgObjectRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgObjectRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM metadata.object_definitions`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgObjectRepo.Count: %w", err)
	}
	return count, nil
}

// PgFieldRepository implements FieldRepository using pgx.
type PgFieldRepository struct {
	pool *pgxpool.Pool
}

// NewPgFieldRepository creates a new PgFieldRepository.
func NewPgFieldRepository(pool *pgxpool.Pool) *PgFieldRepository {
	return &PgFieldRepository{pool: pool}
}

func (r *PgFieldRepository) Create(ctx context.Context, tx pgx.Tx, input CreateFieldInput) (*FieldDefinition, error) {
	var f FieldDefinition
	var subtype *string
	if input.FieldSubtype != nil {
		s := string(*input.FieldSubtype)
		subtype = &s
	}
	err := tx.QueryRow(ctx, `
		INSERT INTO metadata.field_definitions (
			object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config, is_custom, sort_order
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
	`,
		input.ObjectID, input.APIName, input.Label, input.Description, input.HelpText,
		input.FieldType, subtype, input.ReferencedObjectID,
		input.IsRequired, input.IsUnique, input.Config, input.IsCustom, input.SortOrder,
	).Scan(
		&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
		&f.FieldType, &subtype, &f.ReferencedObjectID,
		&f.IsRequired, &f.IsUnique, &f.Config,
		&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
		&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.Create: %w", err)
	}
	if subtype != nil {
		st := FieldSubtype(*subtype)
		f.FieldSubtype = &st
	}
	return &f, nil
}

func (r *PgFieldRepository) GetByID(ctx context.Context, id uuid.UUID) (*FieldDefinition, error) {
	var f FieldDefinition
	var subtype *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
		FROM metadata.field_definitions WHERE id = $1
	`, id).Scan(
		&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
		&f.FieldType, &subtype, &f.ReferencedObjectID,
		&f.IsRequired, &f.IsUnique, &f.Config,
		&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
		&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.GetByID: %w", err)
	}
	if subtype != nil {
		st := FieldSubtype(*subtype)
		f.FieldSubtype = &st
	}
	return &f, nil
}

func (r *PgFieldRepository) GetByObjectAndName(ctx context.Context, objectID uuid.UUID, apiName string) (*FieldDefinition, error) {
	var f FieldDefinition
	var subtype *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
		FROM metadata.field_definitions WHERE object_id = $1 AND api_name = $2
	`, objectID, apiName).Scan(
		&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
		&f.FieldType, &subtype, &f.ReferencedObjectID,
		&f.IsRequired, &f.IsUnique, &f.Config,
		&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
		&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.GetByObjectAndName: %w", err)
	}
	if subtype != nil {
		st := FieldSubtype(*subtype)
		f.FieldSubtype = &st
	}
	return &f, nil
}

func (r *PgFieldRepository) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
		FROM metadata.field_definitions WHERE object_id = $1
		ORDER BY sort_order, created_at
	`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.ListByObjectID: %w", err)
	}
	defer rows.Close()

	fields := make([]FieldDefinition, 0)
	for rows.Next() {
		var f FieldDefinition
		var subtype *string
		if err := rows.Scan(
			&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
			&f.FieldType, &subtype, &f.ReferencedObjectID,
			&f.IsRequired, &f.IsUnique, &f.Config,
			&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
			&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgFieldRepo.ListByObjectID: scan: %w", err)
		}
		if subtype != nil {
			st := FieldSubtype(*subtype)
			f.FieldSubtype = &st
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (r *PgFieldRepository) ListAll(ctx context.Context) ([]FieldDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
		FROM metadata.field_definitions
	`)
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.ListAll: %w", err)
	}
	defer rows.Close()

	fields := make([]FieldDefinition, 0)
	for rows.Next() {
		var f FieldDefinition
		var subtype *string
		if err := rows.Scan(
			&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
			&f.FieldType, &subtype, &f.ReferencedObjectID,
			&f.IsRequired, &f.IsUnique, &f.Config,
			&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
			&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgFieldRepo.ListAll: scan: %w", err)
		}
		if subtype != nil {
			st := FieldSubtype(*subtype)
			f.FieldSubtype = &st
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (r *PgFieldRepository) ListReferenceFields(ctx context.Context) ([]FieldDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
		FROM metadata.field_definitions WHERE field_type = 'reference'
	`)
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.ListReferenceFields: %w", err)
	}
	defer rows.Close()

	fields := make([]FieldDefinition, 0)
	for rows.Next() {
		var f FieldDefinition
		var subtype *string
		if err := rows.Scan(
			&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
			&f.FieldType, &subtype, &f.ReferencedObjectID,
			&f.IsRequired, &f.IsUnique, &f.Config,
			&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
			&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgFieldRepo.ListReferenceFields: scan: %w", err)
		}
		if subtype != nil {
			st := FieldSubtype(*subtype)
			f.FieldSubtype = &st
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (r *PgFieldRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateFieldInput) (*FieldDefinition, error) {
	var f FieldDefinition
	var subtype *string
	err := tx.QueryRow(ctx, `
		UPDATE metadata.field_definitions SET
			label = $2, description = $3, help_text = $4,
			is_required = $5, is_unique = $6, config = $7,
			sort_order = $8, updated_at = now()
		WHERE id = $1
		RETURNING id, object_id, api_name, label, description, help_text,
			field_type, field_subtype, referenced_object_id,
			is_required, is_unique, config,
			is_system_field, is_custom, is_platform_managed,
			sort_order, created_at, updated_at
	`,
		id, input.Label, input.Description, input.HelpText,
		input.IsRequired, input.IsUnique, input.Config,
		input.SortOrder,
	).Scan(
		&f.ID, &f.ObjectID, &f.APIName, &f.Label, &f.Description, &f.HelpText,
		&f.FieldType, &subtype, &f.ReferencedObjectID,
		&f.IsRequired, &f.IsUnique, &f.Config,
		&f.IsSystemField, &f.IsCustom, &f.IsPlatformManaged,
		&f.SortOrder, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgFieldRepo.Update: %w", err)
	}
	if subtype != nil {
		st := FieldSubtype(*subtype)
		f.FieldSubtype = &st
	}
	return &f, nil
}

func (r *PgFieldRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM metadata.field_definitions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgFieldRepo.Delete: %w", err)
	}
	return nil
}

// PgPolymorphicTargetRepository implements PolymorphicTargetRepository using pgx.
type PgPolymorphicTargetRepository struct {
	pool *pgxpool.Pool
}

// NewPgPolymorphicTargetRepository creates a new PgPolymorphicTargetRepository.
func NewPgPolymorphicTargetRepository(pool *pgxpool.Pool) *PgPolymorphicTargetRepository {
	return &PgPolymorphicTargetRepository{pool: pool}
}

func (r *PgPolymorphicTargetRepository) Create(ctx context.Context, tx pgx.Tx, fieldID, objectID uuid.UUID) (*PolymorphicTarget, error) {
	var pt PolymorphicTarget
	err := tx.QueryRow(ctx, `
		INSERT INTO metadata.polymorphic_targets (field_id, object_id)
		VALUES ($1, $2)
		RETURNING id, field_id, object_id, created_at
	`, fieldID, objectID).Scan(&pt.ID, &pt.FieldID, &pt.ObjectID, &pt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("pgPolymorphicTargetRepo.Create: %w", err)
	}
	return &pt, nil
}

func (r *PgPolymorphicTargetRepository) ListByFieldID(ctx context.Context, fieldID uuid.UUID) ([]PolymorphicTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, field_id, object_id, created_at
		FROM metadata.polymorphic_targets WHERE field_id = $1
	`, fieldID)
	if err != nil {
		return nil, fmt.Errorf("pgPolymorphicTargetRepo.ListByFieldID: %w", err)
	}
	defer rows.Close()

	var targets []PolymorphicTarget
	for rows.Next() {
		var pt PolymorphicTarget
		if err := rows.Scan(&pt.ID, &pt.FieldID, &pt.ObjectID, &pt.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgPolymorphicTargetRepo.ListByFieldID: scan: %w", err)
		}
		targets = append(targets, pt)
	}
	return targets, rows.Err()
}

func (r *PgPolymorphicTargetRepository) ListAll(ctx context.Context) ([]PolymorphicTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, field_id, object_id, created_at
		FROM metadata.polymorphic_targets
	`)
	if err != nil {
		return nil, fmt.Errorf("pgPolymorphicTargetRepo.ListAll: %w", err)
	}
	defer rows.Close()

	var targets []PolymorphicTarget
	for rows.Next() {
		var pt PolymorphicTarget
		if err := rows.Scan(&pt.ID, &pt.FieldID, &pt.ObjectID, &pt.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgPolymorphicTargetRepo.ListAll: scan: %w", err)
		}
		targets = append(targets, pt)
	}
	return targets, rows.Err()
}

func (r *PgPolymorphicTargetRepository) DeleteByFieldID(ctx context.Context, tx pgx.Tx, fieldID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM metadata.polymorphic_targets WHERE field_id = $1`, fieldID)
	if err != nil {
		return fmt.Errorf("pgPolymorphicTargetRepo.DeleteByFieldID: %w", err)
	}
	return nil
}
