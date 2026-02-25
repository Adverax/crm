package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgNavigationRepository is a PostgreSQL implementation of NavigationRepository.
type PgNavigationRepository struct {
	pool *pgxpool.Pool
}

// NewPgNavigationRepository creates a new PgNavigationRepository.
func NewPgNavigationRepository(pool *pgxpool.Pool) *PgNavigationRepository {
	return &PgNavigationRepository{pool: pool}
}

func (r *PgNavigationRepository) Create(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.Create: marshal config: %w", err)
	}

	nav := &ProfileNavigation{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.profile_navigation (profile_id, config)
		VALUES ($1, $2)
		RETURNING id, profile_id, config, created_at, updated_at`,
		input.ProfileID, configJSON,
	).Scan(&nav.ID, &nav.ProfileID, &configRaw, &nav.CreatedAt, &nav.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.Create: %w", err)
	}

	if err := json.Unmarshal(configRaw, &nav.Config); err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.Create: unmarshal config: %w", err)
	}
	return nav, nil
}

func (r *PgNavigationRepository) GetByID(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error) {
	nav := &ProfileNavigation{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_navigation
		WHERE id = $1`, id,
	).Scan(&nav.ID, &nav.ProfileID, &configRaw, &nav.CreatedAt, &nav.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgNavigationRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &nav.Config); err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.GetByID: unmarshal config: %w", err)
	}
	return nav, nil
}

func (r *PgNavigationRepository) GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error) {
	nav := &ProfileNavigation{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_navigation
		WHERE profile_id = $1`, profileID,
	).Scan(&nav.ID, &nav.ProfileID, &configRaw, &nav.CreatedAt, &nav.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgNavigationRepo.GetByProfileID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &nav.Config); err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.GetByProfileID: unmarshal config: %w", err)
	}
	return nav, nil
}

func (r *PgNavigationRepository) ListAll(ctx context.Context) ([]ProfileNavigation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_navigation
		ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanNavigations(rows)
}

func (r *PgNavigationRepository) Update(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.Update: marshal config: %w", err)
	}

	nav := &ProfileNavigation{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.profile_navigation SET
			config = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, profile_id, config, created_at, updated_at`,
		id, configJSON,
	).Scan(&nav.ID, &nav.ProfileID, &configRaw, &nav.CreatedAt, &nav.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgNavigationRepo.Update: %w", err)
	}

	if err := json.Unmarshal(configRaw, &nav.Config); err != nil {
		return nil, fmt.Errorf("pgNavigationRepo.Update: unmarshal config: %w", err)
	}
	return nav, nil
}

func (r *PgNavigationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.profile_navigation WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgNavigationRepo.Delete: %w", err)
	}
	return nil
}

func scanNavigations(rows pgx.Rows) ([]ProfileNavigation, error) {
	var navs []ProfileNavigation
	for rows.Next() {
		var nav ProfileNavigation
		var configRaw []byte
		if err := rows.Scan(
			&nav.ID, &nav.ProfileID, &configRaw, &nav.CreatedAt, &nav.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanNavigations: %w", err)
		}
		if err := json.Unmarshal(configRaw, &nav.Config); err != nil {
			return nil, fmt.Errorf("scanNavigations: unmarshal config: %w", err)
		}
		navs = append(navs, nav)
	}
	return navs, rows.Err()
}
