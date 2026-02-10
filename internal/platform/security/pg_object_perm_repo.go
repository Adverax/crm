package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgObjectPermissionRepository implements ObjectPermissionRepository and AllObjectPermissions.
type PgObjectPermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPgObjectPermissionRepository creates a new PgObjectPermissionRepository.
func NewPgObjectPermissionRepository(pool *pgxpool.Pool) *PgObjectPermissionRepository {
	return &PgObjectPermissionRepository{pool: pool}
}

func (r *PgObjectPermissionRepository) Upsert(ctx context.Context, tx pgx.Tx, psID, objectID uuid.UUID, permissions int) (*ObjectPermission, error) {
	var op ObjectPermission
	err := tx.QueryRow(ctx, `
		INSERT INTO security.object_permissions (permission_set_id, object_id, permissions)
		VALUES ($1, $2, $3)
		ON CONFLICT (permission_set_id, object_id)
		DO UPDATE SET permissions = $3, updated_at = now()
		RETURNING id, permission_set_id, object_id, permissions, created_at, updated_at
	`, psID, objectID, permissions).Scan(
		&op.ID, &op.PermissionSetID, &op.ObjectID,
		&op.Permissions, &op.CreatedAt, &op.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgObjectPermRepo.Upsert: %w", err)
	}
	return &op, nil
}

func (r *PgObjectPermissionRepository) GetByPSAndObject(ctx context.Context, psID, objectID uuid.UUID) (*ObjectPermission, error) {
	var op ObjectPermission
	err := r.pool.QueryRow(ctx, `
		SELECT id, permission_set_id, object_id, permissions, created_at, updated_at
		FROM security.object_permissions
		WHERE permission_set_id = $1 AND object_id = $2
	`, psID, objectID).Scan(
		&op.ID, &op.PermissionSetID, &op.ObjectID,
		&op.Permissions, &op.CreatedAt, &op.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgObjectPermRepo.GetByPSAndObject: %w", err)
	}
	return &op, nil
}

func (r *PgObjectPermissionRepository) ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]ObjectPermission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, object_id, permissions, created_at, updated_at
		FROM security.object_permissions
		WHERE permission_set_id = $1
		ORDER BY created_at
	`, psID)
	if err != nil {
		return nil, fmt.Errorf("pgObjectPermRepo.ListByPermissionSetID: %w", err)
	}
	defer rows.Close()

	perms := make([]ObjectPermission, 0)
	for rows.Next() {
		var op ObjectPermission
		if err := rows.Scan(
			&op.ID, &op.PermissionSetID, &op.ObjectID,
			&op.Permissions, &op.CreatedAt, &op.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgObjectPermRepo.ListByPermissionSetID: scan: %w", err)
		}
		perms = append(perms, op)
	}
	return perms, rows.Err()
}

func (r *PgObjectPermissionRepository) ListByPermissionSetIDs(ctx context.Context, psIDs []uuid.UUID) ([]ObjectPermission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, object_id, permissions, created_at, updated_at
		FROM security.object_permissions
		WHERE permission_set_id = ANY($1)
	`, psIDs)
	if err != nil {
		return nil, fmt.Errorf("pgObjectPermRepo.ListByPermissionSetIDs: %w", err)
	}
	defer rows.Close()

	perms := make([]ObjectPermission, 0)
	for rows.Next() {
		var op ObjectPermission
		if err := rows.Scan(
			&op.ID, &op.PermissionSetID, &op.ObjectID,
			&op.Permissions, &op.CreatedAt, &op.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgObjectPermRepo.ListByPermissionSetIDs: scan: %w", err)
		}
		perms = append(perms, op)
	}
	return perms, rows.Err()
}

func (r *PgObjectPermissionRepository) Delete(ctx context.Context, tx pgx.Tx, psID, objectID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM security.object_permissions
		WHERE permission_set_id = $1 AND object_id = $2
	`, psID, objectID)
	if err != nil {
		return fmt.Errorf("pgObjectPermRepo.Delete: %w", err)
	}
	return nil
}
