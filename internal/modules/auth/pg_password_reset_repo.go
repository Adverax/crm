package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgPasswordResetTokenRepository implements PasswordResetTokenRepository using PostgreSQL.
type PgPasswordResetTokenRepository struct {
	pool *pgxpool.Pool
}

// NewPgPasswordResetTokenRepository creates a new PgPasswordResetTokenRepository.
func NewPgPasswordResetTokenRepository(pool *pgxpool.Pool) *PgPasswordResetTokenRepository {
	return &PgPasswordResetTokenRepository{pool: pool}
}

func (r *PgPasswordResetTokenRepository) Create(ctx context.Context, token *PasswordResetToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO iam.password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("pgPasswordResetRepo.Create: %w", err)
	}
	return nil
}

func (r *PgPasswordResetTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var t PasswordResetToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM iam.password_reset_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgPasswordResetRepo.GetByTokenHash: %w", err)
	}
	return &t, nil
}

func (r *PgPasswordResetTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE iam.password_reset_tokens SET used_at = now() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("pgPasswordResetRepo.MarkUsed: %w", err)
	}
	return nil
}

func (r *PgPasswordResetTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM iam.password_reset_tokens WHERE expires_at <= now()
	`)
	if err != nil {
		return 0, fmt.Errorf("pgPasswordResetRepo.DeleteExpired: %w", err)
	}
	return tag.RowsAffected(), nil
}
