package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgDashboardRepository is a PostgreSQL implementation of DashboardRepository.
type PgDashboardRepository struct {
	pool *pgxpool.Pool
}

// NewPgDashboardRepository creates a new PgDashboardRepository.
func NewPgDashboardRepository(pool *pgxpool.Pool) *PgDashboardRepository {
	return &PgDashboardRepository{pool: pool}
}

func (r *PgDashboardRepository) Create(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.Create: marshal config: %w", err)
	}

	dash := &ProfileDashboard{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.profile_dashboards (profile_id, config)
		VALUES ($1, $2)
		RETURNING id, profile_id, config, created_at, updated_at`,
		input.ProfileID, configJSON,
	).Scan(&dash.ID, &dash.ProfileID, &configRaw, &dash.CreatedAt, &dash.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.Create: %w", err)
	}

	if err := json.Unmarshal(configRaw, &dash.Config); err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.Create: unmarshal config: %w", err)
	}
	return dash, nil
}

func (r *PgDashboardRepository) GetByID(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error) {
	dash := &ProfileDashboard{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_dashboards
		WHERE id = $1`, id,
	).Scan(&dash.ID, &dash.ProfileID, &configRaw, &dash.CreatedAt, &dash.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgDashboardRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &dash.Config); err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.GetByID: unmarshal config: %w", err)
	}
	return dash, nil
}

func (r *PgDashboardRepository) GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error) {
	dash := &ProfileDashboard{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_dashboards
		WHERE profile_id = $1`, profileID,
	).Scan(&dash.ID, &dash.ProfileID, &configRaw, &dash.CreatedAt, &dash.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgDashboardRepo.GetByProfileID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &dash.Config); err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.GetByProfileID: unmarshal config: %w", err)
	}
	return dash, nil
}

func (r *PgDashboardRepository) ListAll(ctx context.Context) ([]ProfileDashboard, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, profile_id, config, created_at, updated_at
		FROM metadata.profile_dashboards
		ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanDashboards(rows)
}

func (r *PgDashboardRepository) Update(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.Update: marshal config: %w", err)
	}

	dash := &ProfileDashboard{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.profile_dashboards SET
			config = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, profile_id, config, created_at, updated_at`,
		id, configJSON,
	).Scan(&dash.ID, &dash.ProfileID, &configRaw, &dash.CreatedAt, &dash.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgDashboardRepo.Update: %w", err)
	}

	if err := json.Unmarshal(configRaw, &dash.Config); err != nil {
		return nil, fmt.Errorf("pgDashboardRepo.Update: unmarshal config: %w", err)
	}
	return dash, nil
}

func (r *PgDashboardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.profile_dashboards WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgDashboardRepo.Delete: %w", err)
	}
	return nil
}

func scanDashboards(rows pgx.Rows) ([]ProfileDashboard, error) {
	var dashes []ProfileDashboard
	for rows.Next() {
		var dash ProfileDashboard
		var configRaw []byte
		if err := rows.Scan(
			&dash.ID, &dash.ProfileID, &configRaw, &dash.CreatedAt, &dash.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanDashboards: %w", err)
		}
		if err := json.Unmarshal(configRaw, &dash.Config); err != nil {
			return nil, fmt.Errorf("scanDashboards: unmarshal config: %w", err)
		}
		dashes = append(dashes, dash)
	}
	return dashes, rows.Err()
}
