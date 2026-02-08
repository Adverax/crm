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
