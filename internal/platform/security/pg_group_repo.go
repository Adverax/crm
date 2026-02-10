package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgGroupRepository implements GroupRepository using pgx.
type PgGroupRepository struct {
	pool *pgxpool.Pool
}

// NewPgGroupRepository creates a new PgGroupRepository.
func NewPgGroupRepository(pool *pgxpool.Pool) *PgGroupRepository {
	return &PgGroupRepository{pool: pool}
}

func (r *PgGroupRepository) Create(ctx context.Context, tx pgx.Tx, input CreateGroupInput) (*Group, error) {
	var g Group
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.groups (api_name, label, group_type, related_role_id, related_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
	`,
		input.APIName, input.Label, input.GroupType, input.RelatedRoleID, input.RelatedUserID,
	).Scan(
		&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.Create: %w", err)
	}
	return &g, nil
}

func (r *PgGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	var g Group
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
		FROM iam.groups WHERE id = $1
	`, id).Scan(
		&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.GetByID: %w", err)
	}
	return &g, nil
}

func (r *PgGroupRepository) GetByAPIName(ctx context.Context, apiName string) (*Group, error) {
	var g Group
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
		FROM iam.groups WHERE api_name = $1
	`, apiName).Scan(
		&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.GetByAPIName: %w", err)
	}
	return &g, nil
}

func (r *PgGroupRepository) GetByRelatedRoleID(ctx context.Context, roleID uuid.UUID, groupType GroupType) (*Group, error) {
	var g Group
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
		FROM iam.groups WHERE related_role_id = $1 AND group_type = $2
	`, roleID, groupType).Scan(
		&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.GetByRelatedRoleID: %w", err)
	}
	return &g, nil
}

func (r *PgGroupRepository) GetByRelatedUserID(ctx context.Context, userID uuid.UUID) (*Group, error) {
	var g Group
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
		FROM iam.groups WHERE related_user_id = $1 AND group_type = 'personal'
	`, userID).Scan(
		&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.GetByRelatedUserID: %w", err)
	}
	return &g, nil
}

func (r *PgGroupRepository) List(ctx context.Context, limit, offset int32) ([]Group, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, group_type, related_role_id, related_user_id, created_at, updated_at
		FROM iam.groups
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgGroupRepo.List: %w", err)
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var g Group
		if err := rows.Scan(
			&g.ID, &g.APIName, &g.Label, &g.GroupType, &g.RelatedRoleID, &g.RelatedUserID,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgGroupRepo.List: scan: %w", err)
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *PgGroupRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgGroupRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgGroupRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.groups`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgGroupRepo.Count: %w", err)
	}
	return count, nil
}
