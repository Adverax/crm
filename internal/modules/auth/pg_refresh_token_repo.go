package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRefreshTokenRepository implements RefreshTokenRepository using PostgreSQL.
type PgRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

// NewPgRefreshTokenRepository creates a new PgRefreshTokenRepository.
func NewPgRefreshTokenRepository(pool *pgxpool.Pool) *PgRefreshTokenRepository {
	return &PgRefreshTokenRepository{pool: pool}
}

func (r *PgRefreshTokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO iam.refresh_tokens (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.UpdatedAt)
	if err != nil {
		return fmt.Errorf("pgRefreshTokenRepo.Create: %w", err)
	}
	return nil
}

func (r *PgRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var t RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, updated_at
		FROM iam.refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgRefreshTokenRepo.GetByTokenHash: %w", err)
	}
	return &t, nil
}

func (r *PgRefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM iam.refresh_tokens WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("pgRefreshTokenRepo.DeleteByTokenHash: %w", err)
	}
	return nil
}

func (r *PgRefreshTokenRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM iam.refresh_tokens WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("pgRefreshTokenRepo.DeleteAllByUserID: %w", err)
	}
	return nil
}

func (r *PgRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM iam.refresh_tokens WHERE expires_at <= now()
	`)
	if err != nil {
		return 0, fmt.Errorf("pgRefreshTokenRepo.DeleteExpired: %w", err)
	}
	return tag.RowsAffected(), nil
}
