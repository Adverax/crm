package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgCacheLoader loads metadata from PostgreSQL for cache population.
type PgCacheLoader struct {
	pool *pgxpool.Pool
}

// NewPgCacheLoader creates a new PgCacheLoader.
func NewPgCacheLoader(pool *pgxpool.Pool) *PgCacheLoader {
	return &PgCacheLoader{pool: pool}
}

// LoadAllObjects loads all object definitions from the database.
func (l *PgCacheLoader) LoadAllObjects(ctx context.Context) ([]ObjectDefinition, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, api_name, label, plural_label, description,
		       table_name, object_type,
		       is_platform_managed, is_visible_in_setup,
		       is_custom_fields_allowed, is_deleteable_object,
		       is_createable, is_updateable, is_deleteable,
		       is_queryable, is_searchable,
		       has_activities, has_notes, has_history_tracking, has_sharing_rules,
		       visibility, created_at, updated_at
		FROM metadata.object_definitions
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadAllObjects: %w", err)
	}
	defer rows.Close()

	var objects []ObjectDefinition
	for rows.Next() {
		var obj ObjectDefinition
		if err := rows.Scan(
			&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel, &obj.Description,
			&obj.TableName, &obj.ObjectType,
			&obj.IsPlatformManaged, &obj.IsVisibleInSetup,
			&obj.IsCustomFieldsAllowed, &obj.IsDeleteableObject,
			&obj.IsCreateable, &obj.IsUpdateable, &obj.IsDeleteable,
			&obj.IsQueryable, &obj.IsSearchable,
			&obj.HasActivities, &obj.HasNotes, &obj.HasHistoryTracking, &obj.HasSharingRules,
			&obj.Visibility, &obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgCacheLoader.LoadAllObjects: scan: %w", err)
		}
		objects = append(objects, obj)
	}

	return objects, rows.Err()
}

// LoadAllFields loads all field definitions from the database.
func (l *PgCacheLoader) LoadAllFields(ctx context.Context) ([]FieldDefinition, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, help_text,
		       field_type, field_subtype, referenced_object_id,
		       is_required, is_unique, config,
		       is_system_field, is_custom, is_platform_managed,
		       sort_order, created_at, updated_at
		FROM metadata.field_definitions
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadAllFields: %w", err)
	}
	defer rows.Close()

	var fields []FieldDefinition
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
			return nil, fmt.Errorf("pgCacheLoader.LoadAllFields: scan: %w", err)
		}
		if subtype != nil {
			st := FieldSubtype(*subtype)
			f.FieldSubtype = &st
		}
		fields = append(fields, f)
	}

	return fields, rows.Err()
}

// LoadRelationships loads the relationship registry from the materialized view.
func (l *PgCacheLoader) LoadRelationships(ctx context.Context) ([]RelationshipInfo, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT field_id, field_api_name, relationship_name,
		       child_object_id, child_object_api_name,
		       parent_object_id, parent_object_api_name,
		       reference_subtype, on_delete
		FROM metadata.relationship_registry
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadRelationships: %w", err)
	}
	defer rows.Close()

	var rels []RelationshipInfo
	for rows.Next() {
		var rel RelationshipInfo
		var subtype string
		var parentID *uuid.UUID
		var parentAPIName *string
		var relName *string
		if err := rows.Scan(
			&rel.FieldID, &rel.FieldAPIName, &relName,
			&rel.ChildObjectID, &rel.ChildObjectAPIName,
			&parentID, &parentAPIName,
			&subtype, &rel.OnDelete,
		); err != nil {
			return nil, fmt.Errorf("pgCacheLoader.LoadRelationships: scan: %w", err)
		}
		rel.ReferenceSubtype = FieldSubtype(subtype)
		if parentID != nil {
			rel.ParentObjectID = *parentID
		}
		if parentAPIName != nil {
			rel.ParentObjectAPIName = *parentAPIName
		}
		if relName != nil {
			rel.RelationshipName = *relName
		}
		rels = append(rels, rel)
	}

	return rels, rows.Err()
}

// LoadAllValidationRules loads all validation rules from the database.
func (l *PgCacheLoader) LoadAllValidationRules(ctx context.Context) ([]ValidationRule, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at
		FROM metadata.validation_rules
		ORDER BY object_id, sort_order, api_name
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadAllValidationRules: %w", err)
	}
	defer rows.Close()

	return scanValidationRules(rows)
}

// LoadAllFunctions loads all custom functions from the database.
func (l *PgCacheLoader) LoadAllFunctions(ctx context.Context) ([]Function, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, name, description, params, return_type, body,
			created_at, updated_at
		FROM metadata.functions
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadAllFunctions: %w", err)
	}
	defer rows.Close()

	return scanFunctions(rows)
}

// LoadAllObjectViews loads all object views from the database.
func (l *PgCacheLoader) LoadAllObjectViews(ctx context.Context) ([]ObjectView, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, object_id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at
		FROM metadata.object_views
		ORDER BY object_id, api_name
	`)
	if err != nil {
		return nil, fmt.Errorf("pgCacheLoader.LoadAllObjectViews: %w", err)
	}
	defer rows.Close()

	return scanObjectViews(rows)
}

// RefreshMaterializedView refreshes the relationship_registry materialized view concurrently.
func (l *PgCacheLoader) RefreshMaterializedView(ctx context.Context) error {
	_, err := l.pool.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY metadata.relationship_registry")
	if err != nil {
		return fmt.Errorf("pgCacheLoader.RefreshMaterializedView: %w", err)
	}
	return nil
}
