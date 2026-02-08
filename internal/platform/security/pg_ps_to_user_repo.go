package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgPermissionSetToUserRepository implements PermissionSetToUserRepository using pgx.
type PgPermissionSetToUserRepository struct {
	pool *pgxpool.Pool
}

// NewPgPermissionSetToUserRepository creates a new PgPermissionSetToUserRepository.
func NewPgPermissionSetToUserRepository(pool *pgxpool.Pool) *PgPermissionSetToUserRepository {
	return &PgPermissionSetToUserRepository{pool: pool}
}

func (r *PgPermissionSetToUserRepository) Assign(ctx context.Context, tx pgx.Tx, psID, userID uuid.UUID) (*PermissionSetToUser, error) {
	var psu PermissionSetToUser
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.permission_set_to_users (permission_set_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (permission_set_id, user_id) DO NOTHING
		RETURNING id, permission_set_id, user_id, created_at
	`, psID, userID).Scan(
		&psu.ID, &psu.PermissionSetID, &psu.UserID, &psu.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgPSToUserRepo.Assign: %w", err)
	}
	return &psu, nil
}

func (r *PgPermissionSetToUserRepository) Revoke(ctx context.Context, tx pgx.Tx, psID, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM iam.permission_set_to_users
		WHERE permission_set_id = $1 AND user_id = $2
	`, psID, userID)
	if err != nil {
		return fmt.Errorf("pgPSToUserRepo.Revoke: %w", err)
	}
	return nil
}

func (r *PgPermissionSetToUserRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]PermissionSetToUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, user_id, created_at
		FROM iam.permission_set_to_users
		WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgPSToUserRepo.ListByUserID: %w", err)
	}
	defer rows.Close()

	var assignments []PermissionSetToUser
	for rows.Next() {
		var psu PermissionSetToUser
		if err := rows.Scan(&psu.ID, &psu.PermissionSetID, &psu.UserID, &psu.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgPSToUserRepo.ListByUserID: scan: %w", err)
		}
		assignments = append(assignments, psu)
	}
	return assignments, rows.Err()
}

func (r *PgPermissionSetToUserRepository) ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]PermissionSetToUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, permission_set_id, user_id, created_at
		FROM iam.permission_set_to_users
		WHERE permission_set_id = $1
		ORDER BY created_at
	`, psID)
	if err != nil {
		return nil, fmt.Errorf("pgPSToUserRepo.ListByPermissionSetID: %w", err)
	}
	defer rows.Close()

	var assignments []PermissionSetToUser
	for rows.Next() {
		var psu PermissionSetToUser
		if err := rows.Scan(&psu.ID, &psu.PermissionSetID, &psu.UserID, &psu.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgPSToUserRepo.ListByPermissionSetID: scan: %w", err)
		}
		assignments = append(assignments, psu)
	}
	return assignments, rows.Err()
}
