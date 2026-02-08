package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgEffectivePermissionRepository implements EffectivePermissionRepository using pgx.
type PgEffectivePermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPgEffectivePermissionRepository creates a new PgEffectivePermissionRepository.
func NewPgEffectivePermissionRepository(pool *pgxpool.Pool) *PgEffectivePermissionRepository {
	return &PgEffectivePermissionRepository{pool: pool}
}

func (r *PgEffectivePermissionRepository) GetOLS(ctx context.Context, userID, objectID uuid.UUID) (*EffectiveOLS, error) {
	var e EffectiveOLS
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, object_id, permissions, computed_at
		FROM security.effective_ols
		WHERE user_id = $1 AND object_id = $2
	`, userID, objectID).Scan(
		&e.UserID, &e.ObjectID, &e.Permissions, &e.ComputedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgEffectiveRepo.GetOLS: %w", err)
	}
	return &e, nil
}

func (r *PgEffectivePermissionRepository) GetFLS(ctx context.Context, userID, fieldID uuid.UUID) (*EffectiveFLS, error) {
	var e EffectiveFLS
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, field_id, permissions, computed_at
		FROM security.effective_fls
		WHERE user_id = $1 AND field_id = $2
	`, userID, fieldID).Scan(
		&e.UserID, &e.FieldID, &e.Permissions, &e.ComputedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgEffectiveRepo.GetFLS: %w", err)
	}
	return &e, nil
}

func (r *PgEffectivePermissionRepository) GetFieldList(ctx context.Context, userID, objectID uuid.UUID, mask int) (*EffectiveFieldList, error) {
	var e EffectiveFieldList
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, object_id, mask, field_names, computed_at
		FROM security.effective_field_lists
		WHERE user_id = $1 AND object_id = $2 AND mask = $3
	`, userID, objectID, mask).Scan(
		&e.UserID, &e.ObjectID, &e.Mask, &e.FieldNames, &e.ComputedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgEffectiveRepo.GetFieldList: %w", err)
	}
	return &e, nil
}

func (r *PgEffectivePermissionRepository) UpsertOLS(ctx context.Context, tx pgx.Tx, userID, objectID uuid.UUID, permissions int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO security.effective_ols (user_id, object_id, permissions, computed_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, object_id)
		DO UPDATE SET permissions = $3, computed_at = now()
	`, userID, objectID, permissions)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.UpsertOLS: %w", err)
	}
	return nil
}

func (r *PgEffectivePermissionRepository) UpsertFLS(ctx context.Context, tx pgx.Tx, userID, fieldID uuid.UUID, permissions int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO security.effective_fls (user_id, field_id, permissions, computed_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, field_id)
		DO UPDATE SET permissions = $3, computed_at = now()
	`, userID, fieldID, permissions)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.UpsertFLS: %w", err)
	}
	return nil
}

func (r *PgEffectivePermissionRepository) UpsertFieldList(ctx context.Context, tx pgx.Tx, userID, objectID uuid.UUID, mask int, fieldNames []string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO security.effective_field_lists (user_id, object_id, mask, field_names, computed_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (user_id, object_id, mask)
		DO UPDATE SET field_names = $4, computed_at = now()
	`, userID, objectID, mask, fieldNames)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.UpsertFieldList: %w", err)
	}
	return nil
}

func (r *PgEffectivePermissionRepository) DeleteByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_field_lists WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.DeleteByUserID: field_lists: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM security.effective_fls WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.DeleteByUserID: fls: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM security.effective_ols WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.DeleteByUserID: ols: %w", err)
	}
	return nil
}
