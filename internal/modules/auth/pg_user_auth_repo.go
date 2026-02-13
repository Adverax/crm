package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgUserAuthRepository implements UserAuthRepository using PostgreSQL.
type PgUserAuthRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserAuthRepository creates a new PgUserAuthRepository.
func NewPgUserAuthRepository(pool *pgxpool.Pool) *PgUserAuthRepository {
	return &PgUserAuthRepository{pool: pool}
}

func (r *PgUserAuthRepository) GetByUsername(ctx context.Context, username string) (*UserWithPassword, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, username, email, first_name, last_name, profile_id, role_id,
		       is_active, password_hash
		FROM iam.users WHERE username = $1
	`, username))
}

func (r *PgUserAuthRepository) GetByID(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, username, email, first_name, last_name, profile_id, role_id,
		       is_active, password_hash
		FROM iam.users WHERE id = $1
	`, id))
}

func (r *PgUserAuthRepository) GetByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, username, email, first_name, last_name, profile_id, role_id,
		       is_active, password_hash
		FROM iam.users WHERE email = $1
	`, email))
}

func (r *PgUserAuthRepository) SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE iam.users SET password_hash = $1, updated_at = now() WHERE id = $2
	`, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("pgUserAuthRepo.SetPassword: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pgUserAuthRepo.SetPassword: user not found")
	}
	return nil
}

func (r *PgUserAuthRepository) scanUser(row pgx.Row) (*UserWithPassword, error) {
	var u UserWithPassword
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
		&u.ProfileID, &u.RoleID, &u.IsActive, &u.PasswordHash,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserAuthRepo.scanUser: %w", err)
	}
	return &u, nil
}
