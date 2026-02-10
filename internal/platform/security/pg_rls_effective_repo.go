package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRLSEffectiveCacheRepository implements RLSEffectiveCacheRepository using pgx.
type PgRLSEffectiveCacheRepository struct {
	pool *pgxpool.Pool
}

// NewPgRLSEffectiveCacheRepository creates a new PgRLSEffectiveCacheRepository.
func NewPgRLSEffectiveCacheRepository(pool *pgxpool.Pool) *PgRLSEffectiveCacheRepository {
	return &PgRLSEffectiveCacheRepository{pool: pool}
}

func (r *PgRLSEffectiveCacheRepository) ReplaceRoleHierarchy(ctx context.Context, tx pgx.Tx, entries []EffectiveRoleHierarchy) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_role_hierarchy`)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceRoleHierarchy: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_role_hierarchy (ancestor_role_id, descendant_role_id, depth)
			VALUES ($1, $2, $3)
		`, e.AncestorRoleID, e.DescendantRoleID, e.Depth)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceRoleHierarchy: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) ReplaceVisibleOwners(ctx context.Context, tx pgx.Tx, userID uuid.UUID, entries []EffectiveVisibleOwner) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_visible_owner WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceVisibleOwners: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_visible_owner (user_id, visible_owner_id)
			VALUES ($1, $2)
		`, e.UserID, e.VisibleOwnerID)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceVisibleOwners: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) ReplaceVisibleOwnersAll(ctx context.Context, tx pgx.Tx, entries []EffectiveVisibleOwner) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_visible_owner`)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceVisibleOwnersAll: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_visible_owner (user_id, visible_owner_id)
			VALUES ($1, $2)
		`, e.UserID, e.VisibleOwnerID)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceVisibleOwnersAll: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) ReplaceGroupMembers(ctx context.Context, tx pgx.Tx, groupID uuid.UUID, entries []EffectiveGroupMember) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_group_members WHERE group_id = $1`, groupID)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceGroupMembers: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_group_members (group_id, user_id)
			VALUES ($1, $2)
		`, e.GroupID, e.UserID)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceGroupMembers: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) ReplaceGroupMembersAll(ctx context.Context, tx pgx.Tx, entries []EffectiveGroupMember) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_group_members`)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceGroupMembersAll: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_group_members (group_id, user_id)
			VALUES ($1, $2)
		`, e.GroupID, e.UserID)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceGroupMembersAll: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) ReplaceObjectHierarchy(ctx context.Context, tx pgx.Tx, entries []EffectiveObjectHierarchy) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.effective_object_hierarchy`)
	if err != nil {
		return fmt.Errorf("pgRLSEffectiveRepo.ReplaceObjectHierarchy: delete: %w", err)
	}

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO security.effective_object_hierarchy (ancestor_object_id, descendant_object_id, depth)
			VALUES ($1, $2, $3)
		`, e.AncestorObjectID, e.DescendantObjectID, e.Depth)
		if err != nil {
			return fmt.Errorf("pgRLSEffectiveRepo.ReplaceObjectHierarchy: insert: %w", err)
		}
	}
	return nil
}

func (r *PgRLSEffectiveCacheRepository) GetVisibleOwners(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT visible_owner_id FROM security.effective_visible_owner WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgRLSEffectiveRepo.GetVisibleOwners: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgRLSEffectiveRepo.GetVisibleOwners: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgRLSEffectiveCacheRepository) GetGroupMemberships(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT group_id FROM security.effective_group_members WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgRLSEffectiveRepo.GetGroupMemberships: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgRLSEffectiveRepo.GetGroupMemberships: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgRLSEffectiveCacheRepository) GetRoleDescendants(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT descendant_role_id FROM security.effective_role_hierarchy
		WHERE ancestor_role_id = $1 AND depth > 0
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("pgRLSEffectiveRepo.GetRoleDescendants: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgRLSEffectiveRepo.GetRoleDescendants: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgRLSEffectiveCacheRepository) ListAllRoles(ctx context.Context) ([]EffectiveRoleHierarchy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ancestor_role_id, descendant_role_id, depth
		FROM security.effective_role_hierarchy
	`)
	if err != nil {
		return nil, fmt.Errorf("pgRLSEffectiveRepo.ListAllRoles: %w", err)
	}
	defer rows.Close()

	var entries []EffectiveRoleHierarchy
	for rows.Next() {
		var e EffectiveRoleHierarchy
		if err := rows.Scan(&e.AncestorRoleID, &e.DescendantRoleID, &e.Depth); err != nil {
			return nil, fmt.Errorf("pgRLSEffectiveRepo.ListAllRoles: scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
