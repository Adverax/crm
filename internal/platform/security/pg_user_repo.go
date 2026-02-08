package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgUserRepository implements UserRepository using pgx.
type PgUserRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserRepository creates a new PgUserRepository.
func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) Create(ctx context.Context, tx pgx.Tx, input CreateUserInput) (*User, error) {
	var u User
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.users (username, email, first_name, last_name, profile_id, role_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, username, email, first_name, last_name,
			profile_id, role_id, is_active, created_at, updated_at
	`, input.Username, input.Email, input.FirstName, input.LastName,
		input.ProfileID, input.RoleID).Scan(
		&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
		&u.ProfileID, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgUserRepo.Create: %w", err)
	}
	return &u, nil
}

func (r *PgUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, first_name, last_name,
			profile_id, role_id, is_active, created_at, updated_at
		FROM iam.users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
		&u.ProfileID, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserRepo.GetByID: %w", err)
	}
	return &u, nil
}

func (r *PgUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, first_name, last_name,
			profile_id, role_id, is_active, created_at, updated_at
		FROM iam.users WHERE username = $1
	`, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
		&u.ProfileID, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserRepo.GetByUsername: %w", err)
	}
	return &u, nil
}

func (r *PgUserRepository) List(ctx context.Context, limit, offset int32) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, email, first_name, last_name,
			profile_id, role_id, is_active, created_at, updated_at
		FROM iam.users
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgUserRepo.List: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
			&u.ProfileID, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgUserRepo.List: scan: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PgUserRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateUserInput) (*User, error) {
	var u User
	err := tx.QueryRow(ctx, `
		UPDATE iam.users SET
			email = $2, first_name = $3, last_name = $4,
			profile_id = $5, role_id = $6, is_active = $7, updated_at = now()
		WHERE id = $1
		RETURNING id, username, email, first_name, last_name,
			profile_id, role_id, is_active, created_at, updated_at
	`, id, input.Email, input.FirstName, input.LastName,
		input.ProfileID, input.RoleID, input.IsActive).Scan(
		&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName,
		&u.ProfileID, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgUserRepo.Update: %w", err)
	}
	return &u, nil
}

func (r *PgUserRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgUserRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgUserRepo.Count: %w", err)
	}
	return count, nil
}
