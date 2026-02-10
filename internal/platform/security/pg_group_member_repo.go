package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgGroupMemberRepository implements GroupMemberRepository using pgx.
type PgGroupMemberRepository struct {
	pool *pgxpool.Pool
}

// NewPgGroupMemberRepository creates a new PgGroupMemberRepository.
func NewPgGroupMemberRepository(pool *pgxpool.Pool) *PgGroupMemberRepository {
	return &PgGroupMemberRepository{pool: pool}
}

func (r *PgGroupMemberRepository) Add(ctx context.Context, tx pgx.Tx, input AddGroupMemberInput) (*GroupMember, error) {
	var m GroupMember
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.group_members (group_id, member_user_id, member_group_id)
		VALUES ($1, $2, $3)
		RETURNING id, group_id, member_user_id, member_group_id, created_at
	`,
		input.GroupID, input.MemberUserID, input.MemberGroupID,
	).Scan(
		&m.ID, &m.GroupID, &m.MemberUserID, &m.MemberGroupID, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgGroupMemberRepo.Add: %w", err)
	}
	return &m, nil
}

func (r *PgGroupMemberRepository) Remove(ctx context.Context, tx pgx.Tx, groupID uuid.UUID, memberUserID *uuid.UUID, memberGroupID *uuid.UUID) error {
	if memberUserID != nil {
		_, err := tx.Exec(ctx, `
			DELETE FROM iam.group_members WHERE group_id = $1 AND member_user_id = $2
		`, groupID, *memberUserID)
		if err != nil {
			return fmt.Errorf("pgGroupMemberRepo.Remove: %w", err)
		}
		return nil
	}
	if memberGroupID != nil {
		_, err := tx.Exec(ctx, `
			DELETE FROM iam.group_members WHERE group_id = $1 AND member_group_id = $2
		`, groupID, *memberGroupID)
		if err != nil {
			return fmt.Errorf("pgGroupMemberRepo.Remove: %w", err)
		}
		return nil
	}
	return fmt.Errorf("pgGroupMemberRepo.Remove: either member_user_id or member_group_id must be set")
}

func (r *PgGroupMemberRepository) ListByGroupID(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, group_id, member_user_id, member_group_id, created_at
		FROM iam.group_members WHERE group_id = $1
		ORDER BY created_at
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("pgGroupMemberRepo.ListByGroupID: %w", err)
	}
	defer rows.Close()

	members := make([]GroupMember, 0)
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.MemberUserID, &m.MemberGroupID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgGroupMemberRepo.ListByGroupID: scan: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *PgGroupMemberRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]GroupMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, group_id, member_user_id, member_group_id, created_at
		FROM iam.group_members WHERE member_user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgGroupMemberRepo.ListByUserID: %w", err)
	}
	defer rows.Close()

	members := make([]GroupMember, 0)
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.MemberUserID, &m.MemberGroupID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("pgGroupMemberRepo.ListByUserID: scan: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *PgGroupMemberRepository) DeleteByGroupID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.group_members WHERE group_id = $1`, groupID)
	if err != nil {
		return fmt.Errorf("pgGroupMemberRepo.DeleteByGroupID: %w", err)
	}
	return nil
}
