package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgProcedureRepository is a PostgreSQL implementation of ProcedureRepository.
type PgProcedureRepository struct {
	pool *pgxpool.Pool
}

// NewPgProcedureRepository creates a new PgProcedureRepository.
func NewPgProcedureRepository(pool *pgxpool.Pool) *PgProcedureRepository {
	return &PgProcedureRepository{pool: pool}
}

func (r *PgProcedureRepository) Create(ctx context.Context, input CreateProcedureInput) (*Procedure, error) {
	p := &Procedure{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO metadata.procedures (code, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, code, name, description,
			draft_version_id, published_version_id,
			created_at, updated_at`,
		input.Code, input.Name, input.Description,
	).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.DraftVersionID, &p.PublishedVersionID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.Create: %w", err)
	}
	return p, nil
}

func (r *PgProcedureRepository) GetByID(ctx context.Context, id uuid.UUID) (*Procedure, error) {
	p := &Procedure{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, description,
			draft_version_id, published_version_id,
			created_at, updated_at
		FROM metadata.procedures
		WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.DraftVersionID, &p.PublishedVersionID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgProcedureRepo.GetByID: %w", err)
	}
	return p, nil
}

func (r *PgProcedureRepository) GetByCode(ctx context.Context, code string) (*Procedure, error) {
	p := &Procedure{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, description,
			draft_version_id, published_version_id,
			created_at, updated_at
		FROM metadata.procedures
		WHERE code = $1`, code,
	).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.DraftVersionID, &p.PublishedVersionID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgProcedureRepo.GetByCode: %w", err)
	}
	return p, nil
}

func (r *PgProcedureRepository) ListAll(ctx context.Context) ([]Procedure, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, description,
			draft_version_id, published_version_id,
			created_at, updated_at
		FROM metadata.procedures
		ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanProcedures(rows)
}

func (r *PgProcedureRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.procedures WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgProcedureRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM metadata.procedures`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgProcedureRepo.Count: %w", err)
	}
	return count, nil
}

func (r *PgProcedureRepository) UpdateMetadata(ctx context.Context, id uuid.UUID, input UpdateProcedureMetadataInput) (*Procedure, error) {
	p := &Procedure{}
	err := r.pool.QueryRow(ctx, `
		UPDATE metadata.procedures SET
			name = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, code, name, description,
			draft_version_id, published_version_id,
			created_at, updated_at`,
		id, input.Name, input.Description,
	).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.DraftVersionID, &p.PublishedVersionID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgProcedureRepo.UpdateMetadata: %w", err)
	}
	return p, nil
}

func (r *PgProcedureRepository) SetDraftVersionID(ctx context.Context, id uuid.UUID, versionID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE metadata.procedures SET draft_version_id = $2, updated_at = now()
		WHERE id = $1`, id, versionID)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.SetDraftVersionID: %w", err)
	}
	return nil
}

func (r *PgProcedureRepository) SetPublishedVersionID(ctx context.Context, id uuid.UUID, versionID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE metadata.procedures SET published_version_id = $2, updated_at = now()
		WHERE id = $1`, id, versionID)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.SetPublishedVersionID: %w", err)
	}
	return nil
}

// --- Version methods ---

func (r *PgProcedureRepository) CreateVersion(ctx context.Context, procID uuid.UUID, version int, def ProcedureDefinition, changeSummary string, createdBy *uuid.UUID) (*ProcedureVersion, error) {
	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.CreateVersion: marshal definition: %w", err)
	}

	v := &ProcedureVersion{}
	var defRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.procedure_versions
			(procedure_id, version, definition, change_summary, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at`,
		procID, version, defJSON, changeSummary, createdBy,
	).Scan(
		&v.ID, &v.ProcedureID, &v.Version, &defRaw,
		&v.Status, &v.ChangeSummary, &v.CreatedBy,
		&v.CreatedAt, &v.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.CreateVersion: %w", err)
	}
	if err := json.Unmarshal(defRaw, &v.Definition); err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.CreateVersion: unmarshal definition: %w", err)
	}
	return v, nil
}

func (r *PgProcedureRepository) GetVersionByID(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error) {
	return r.scanSingleVersion(ctx, `
		SELECT id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at
		FROM metadata.procedure_versions
		WHERE id = $1`, id)
}

func (r *PgProcedureRepository) GetDraftVersion(ctx context.Context, procID uuid.UUID) (*ProcedureVersion, error) {
	return r.scanSingleVersion(ctx, `
		SELECT id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at
		FROM metadata.procedure_versions
		WHERE procedure_id = $1 AND status = 'draft'`, procID)
}

func (r *PgProcedureRepository) GetPublishedVersion(ctx context.Context, procID uuid.UUID) (*ProcedureVersion, error) {
	return r.scanSingleVersion(ctx, `
		SELECT id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at
		FROM metadata.procedure_versions
		WHERE procedure_id = $1 AND status = 'published'`, procID)
}

func (r *PgProcedureRepository) UpdateDraft(ctx context.Context, versionID uuid.UUID, def ProcedureDefinition, changeSummary string) (*ProcedureVersion, error) {
	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.UpdateDraft: marshal definition: %w", err)
	}

	v := &ProcedureVersion{}
	var defRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.procedure_versions SET
			definition = $2, change_summary = $3
		WHERE id = $1 AND status = 'draft'
		RETURNING id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at`,
		versionID, defJSON, changeSummary,
	).Scan(
		&v.ID, &v.ProcedureID, &v.Version, &defRaw,
		&v.Status, &v.ChangeSummary, &v.CreatedBy,
		&v.CreatedAt, &v.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgProcedureRepo.UpdateDraft: %w", err)
	}
	if err := json.Unmarshal(defRaw, &v.Definition); err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.UpdateDraft: unmarshal definition: %w", err)
	}
	return v, nil
}

func (r *PgProcedureRepository) DeleteVersion(ctx context.Context, versionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.procedure_versions WHERE id = $1`, versionID)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.DeleteVersion: %w", err)
	}
	return nil
}

func (r *PgProcedureRepository) ListVersions(ctx context.Context, procID uuid.UUID) ([]ProcedureVersion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at
		FROM metadata.procedure_versions
		WHERE procedure_id = $1
		ORDER BY version DESC`, procID)
	if err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.ListVersions: %w", err)
	}
	defer rows.Close()

	return scanProcedureVersions(rows)
}

func (r *PgProcedureRepository) UpdateVersionStatus(ctx context.Context, versionID uuid.UUID, status VersionStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE metadata.procedure_versions SET status = $2
		WHERE id = $1`, versionID, string(status))
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.UpdateVersionStatus: %w", err)
	}
	return nil
}

func (r *PgProcedureRepository) SetVersionPublishedAt(ctx context.Context, versionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE metadata.procedure_versions SET published_at = now()
		WHERE id = $1`, versionID)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.SetVersionPublishedAt: %w", err)
	}
	return nil
}

func (r *PgProcedureRepository) GetMaxVersion(ctx context.Context, procID uuid.UUID) (int, error) {
	var maxVersion *int
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(version) FROM metadata.procedure_versions
		WHERE procedure_id = $1`, procID).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("pgProcedureRepo.GetMaxVersion: %w", err)
	}
	if maxVersion == nil {
		return 0, nil
	}
	return *maxVersion, nil
}

func (r *PgProcedureRepository) GetPreviousPublished(ctx context.Context, procID uuid.UUID, beforeVersion int) (*ProcedureVersion, error) {
	return r.scanSingleVersion(ctx, `
		SELECT id, procedure_id, version, definition, status,
			change_summary, created_by, created_at, published_at
		FROM metadata.procedure_versions
		WHERE procedure_id = $1 AND status = 'superseded' AND version < $2
		ORDER BY version DESC
		LIMIT 1`, procID, beforeVersion)
}

func (r *PgProcedureRepository) CountSuperseded(ctx context.Context, procID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM metadata.procedure_versions
		WHERE procedure_id = $1 AND status = 'superseded'`, procID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgProcedureRepo.CountSuperseded: %w", err)
	}
	return count, nil
}

func (r *PgProcedureRepository) DeleteOldestSuperseded(ctx context.Context, procID uuid.UUID, keepCount int) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM metadata.procedure_versions
		WHERE id IN (
			SELECT id FROM metadata.procedure_versions
			WHERE procedure_id = $1 AND status = 'superseded'
			ORDER BY version ASC
			LIMIT (
				SELECT GREATEST(COUNT(*) - $2, 0)
				FROM metadata.procedure_versions
				WHERE procedure_id = $1 AND status = 'superseded'
			)
		)`, procID, keepCount)
	if err != nil {
		return fmt.Errorf("pgProcedureRepo.DeleteOldestSuperseded: %w", err)
	}
	return nil
}

// --- helpers ---

func (r *PgProcedureRepository) scanSingleVersion(ctx context.Context, query string, args ...any) (*ProcedureVersion, error) {
	v := &ProcedureVersion{}
	var defRaw []byte
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&v.ID, &v.ProcedureID, &v.Version, &defRaw,
		&v.Status, &v.ChangeSummary, &v.CreatedBy,
		&v.CreatedAt, &v.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgProcedureRepo.scanSingleVersion: %w", err)
	}
	if err := json.Unmarshal(defRaw, &v.Definition); err != nil {
		return nil, fmt.Errorf("pgProcedureRepo.scanSingleVersion: unmarshal definition: %w", err)
	}
	return v, nil
}

func scanProcedures(rows pgx.Rows) ([]Procedure, error) {
	var procedures []Procedure
	for rows.Next() {
		var p Procedure
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.Description,
			&p.DraftVersionID, &p.PublishedVersionID,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanProcedures: %w", err)
		}
		procedures = append(procedures, p)
	}
	return procedures, rows.Err()
}

func scanProcedureVersions(rows pgx.Rows) ([]ProcedureVersion, error) {
	var versions []ProcedureVersion
	for rows.Next() {
		var v ProcedureVersion
		var defRaw []byte
		if err := rows.Scan(
			&v.ID, &v.ProcedureID, &v.Version, &defRaw,
			&v.Status, &v.ChangeSummary, &v.CreatedBy,
			&v.CreatedAt, &v.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("scanProcedureVersions: %w", err)
		}
		if err := json.Unmarshal(defRaw, &v.Definition); err != nil {
			return nil, fmt.Errorf("scanProcedureVersions: unmarshal definition: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
