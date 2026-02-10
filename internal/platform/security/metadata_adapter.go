package security

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgMetadataFieldLister implements MetadataFieldLister using direct DB queries.
// This avoids circular dependency with the metadata package.
type PgMetadataFieldLister struct {
	pool *pgxpool.Pool
}

// NewPgMetadataFieldLister creates a new PgMetadataFieldLister.
func NewPgMetadataFieldLister(pool *pgxpool.Pool) *PgMetadataFieldLister {
	return &PgMetadataFieldLister{pool: pool}
}

func (l *PgMetadataFieldLister) ListFieldsByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldInfo, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT id, api_name
		FROM metadata.field_definitions
		WHERE object_id = $1
		ORDER BY sort_order, created_at
	`, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []FieldInfo
	for rows.Next() {
		var f FieldInfo
		if err := rows.Scan(&f.ID, &f.APIName); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (l *PgMetadataFieldLister) ListAllObjectIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := l.pool.Query(ctx, `SELECT id FROM metadata.object_definitions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (l *PgMetadataFieldLister) GetObjectVisibility(ctx context.Context, objectID uuid.UUID) (string, error) {
	var visibility string
	err := l.pool.QueryRow(ctx, `
		SELECT visibility FROM metadata.object_definitions WHERE id = $1
	`, objectID).Scan(&visibility)
	if err != nil {
		return "", err
	}
	return visibility, nil
}

func (l *PgMetadataFieldLister) GetObjectTableName(ctx context.Context, objectID uuid.UUID) (string, error) {
	var tableName string
	err := l.pool.QueryRow(ctx, `
		SELECT table_name FROM metadata.object_definitions WHERE id = $1
	`, objectID).Scan(&tableName)
	if err != nil {
		return "", err
	}
	return tableName, nil
}

func (l *PgMetadataFieldLister) ListCompositionFields(ctx context.Context) ([]CompositionFieldInfo, error) {
	rows, err := l.pool.Query(ctx, `
		SELECT fd.object_id, fd.referenced_object_id
		FROM metadata.field_definitions fd
		WHERE fd.field_type = 'reference'
		  AND fd.field_subtype = 'composition'
		  AND fd.referenced_object_id IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []CompositionFieldInfo
	for rows.Next() {
		var f CompositionFieldInfo
		if err := rows.Scan(&f.ChildObjectID, &f.ParentObjectID); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}
