package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgFieldPermissionRepository implements FieldPermissionRepository and AllFieldPermissions.
type PgFieldPermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPgFieldPermissionRepository creates a new PgFieldPermissionRepository.
func NewPgFieldPermissionRepository(pool *pgxpool.Pool) *PgFieldPermissionRepository {
	return &PgFieldPermissionRepository{pool: pool}
}

func (r *PgFieldPermissionRepository) Upsert(ctx context.Context, tx pgx.Tx, psID, fieldID uuid.UUID, permissions int) (*FieldPermission, error) {
	var fp FieldPermission
	err := tx.QueryRow(ctx, `
		INSERT INTO security.field_permissions (permission_set_id, field_id, permissions)
		VALUES ($1, $2, $3)
		ON CONFLICT (permission_set_id, field_id)
		DO UPDATE SET permissions = $3, updated_at = now()
		RETURNING id, permission_set_id, field_id, permissions, created_at, updated_at
	`, psID, fieldID, permissions).Scan(
		&fp.ID, &fp.PermissionSetID, &fp.FieldID,
		&fp.Permissions, &fp.CreatedAt, &fp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgFieldPermRepo.Upsert: %w", err)
	}
	return &fp, nil
}

func (r *PgFieldPermissionRepository) GetByPSAndField(ctx context.Context, psID, fieldID uuid.UUID) (*FieldPermission, error) {
	var fp FieldPermission
	err := r.pool.QueryRow(ctx, `
		SELECT id, permission_set_id, field_id, permissions, created_at, updated_at
		FROM security.field_permissions
		WHERE permission_set_id = $1 AND field_id = $2
	`, psID, fieldID).Scan(
		&fp.ID, &fp.PermissionSetID, &fp.FieldID,
		&fp.Permissions, &fp.CreatedAt, &fp.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgFieldPermRepo.GetByPSAndField: %w", err)
	}
	return &fp, nil
}

func (r *PgFieldPermissionRepository) ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]FieldPermission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, field_id, permissions, created_at, updated_at
		FROM security.field_permissions
		WHERE permission_set_id = $1
		ORDER BY created_at
	`, psID)
	if err != nil {
		return nil, fmt.Errorf("pgFieldPermRepo.ListByPermissionSetID: %w", err)
	}
	defer rows.Close()

	perms := make([]FieldPermission, 0)
	for rows.Next() {
		var fp FieldPermission
		if err := rows.Scan(
			&fp.ID, &fp.PermissionSetID, &fp.FieldID,
			&fp.Permissions, &fp.CreatedAt, &fp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgFieldPermRepo.ListByPermissionSetID: scan: %w", err)
		}
		perms = append(perms, fp)
	}
	return perms, rows.Err()
}

func (r *PgFieldPermissionRepository) ListByPermissionSetIDs(ctx context.Context, psIDs []uuid.UUID) ([]FieldPermission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, field_id, permissions, created_at, updated_at
		FROM security.field_permissions
		WHERE permission_set_id = ANY($1)
	`, psIDs)
	if err != nil {
		return nil, fmt.Errorf("pgFieldPermRepo.ListByPermissionSetIDs: %w", err)
	}
	defer rows.Close()

	perms := make([]FieldPermission, 0)
	for rows.Next() {
		var fp FieldPermission
		if err := rows.Scan(
			&fp.ID, &fp.PermissionSetID, &fp.FieldID,
			&fp.Permissions, &fp.CreatedAt, &fp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgFieldPermRepo.ListByPermissionSetIDs: scan: %w", err)
		}
		perms = append(perms, fp)
	}
	return perms, rows.Err()
}

func (r *PgFieldPermissionRepository) Delete(ctx context.Context, tx pgx.Tx, psID, fieldID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM security.field_permissions
		WHERE permission_set_id = $1 AND field_id = $2
	`, psID, fieldID)
	if err != nil {
		return fmt.Errorf("pgFieldPermRepo.Delete: %w", err)
	}
	return nil
}
