package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgFunctionRepository is a PostgreSQL implementation of FunctionRepository.
type PgFunctionRepository struct {
	pool *pgxpool.Pool
}

// NewPgFunctionRepository creates a new PgFunctionRepository.
func NewPgFunctionRepository(pool *pgxpool.Pool) *PgFunctionRepository {
	return &PgFunctionRepository{pool: pool}
}

func (r *PgFunctionRepository) Create(ctx context.Context, input CreateFunctionInput) (*Function, error) {
	paramsJSON, err := json.Marshal(input.Params)
	if err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.Create: marshal params: %w", err)
	}

	fn := &Function{}
	var paramsRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.functions
			(name, description, params, return_type, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, params, return_type, body,
			created_at, updated_at`,
		input.Name, input.Description, paramsJSON,
		input.ReturnType, input.Body,
	).Scan(
		&fn.ID, &fn.Name, &fn.Description, &paramsRaw,
		&fn.ReturnType, &fn.Body, &fn.CreatedAt, &fn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.Create: %w", err)
	}

	if err := json.Unmarshal(paramsRaw, &fn.Params); err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.Create: unmarshal params: %w", err)
	}
	return fn, nil
}

func (r *PgFunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Function, error) {
	fn := &Function{}
	var paramsRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, params, return_type, body,
			created_at, updated_at
		FROM metadata.functions
		WHERE id = $1`, id,
	).Scan(
		&fn.ID, &fn.Name, &fn.Description, &paramsRaw,
		&fn.ReturnType, &fn.Body, &fn.CreatedAt, &fn.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgFunctionRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(paramsRaw, &fn.Params); err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.GetByID: unmarshal params: %w", err)
	}
	return fn, nil
}

func (r *PgFunctionRepository) GetByName(ctx context.Context, name string) (*Function, error) {
	fn := &Function{}
	var paramsRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, params, return_type, body,
			created_at, updated_at
		FROM metadata.functions
		WHERE name = $1`, name,
	).Scan(
		&fn.ID, &fn.Name, &fn.Description, &paramsRaw,
		&fn.ReturnType, &fn.Body, &fn.CreatedAt, &fn.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgFunctionRepo.GetByName: %w", err)
	}

	if err := json.Unmarshal(paramsRaw, &fn.Params); err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.GetByName: unmarshal params: %w", err)
	}
	return fn, nil
}

func (r *PgFunctionRepository) ListAll(ctx context.Context) ([]Function, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, params, return_type, body,
			created_at, updated_at
		FROM metadata.functions
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanFunctions(rows)
}

func (r *PgFunctionRepository) Update(ctx context.Context, id uuid.UUID, input UpdateFunctionInput) (*Function, error) {
	paramsJSON, err := json.Marshal(input.Params)
	if err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.Update: marshal params: %w", err)
	}

	fn := &Function{}
	var paramsRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.functions SET
			description = $2, params = $3, return_type = $4,
			body = $5, updated_at = now()
		WHERE id = $1
		RETURNING id, name, description, params, return_type, body,
			created_at, updated_at`,
		id, input.Description, paramsJSON,
		input.ReturnType, input.Body,
	).Scan(
		&fn.ID, &fn.Name, &fn.Description, &paramsRaw,
		&fn.ReturnType, &fn.Body, &fn.CreatedAt, &fn.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgFunctionRepo.Update: %w", err)
	}

	if err := json.Unmarshal(paramsRaw, &fn.Params); err != nil {
		return nil, fmt.Errorf("pgFunctionRepo.Update: unmarshal params: %w", err)
	}
	return fn, nil
}

func (r *PgFunctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.functions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgFunctionRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgFunctionRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM metadata.functions`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgFunctionRepo.Count: %w", err)
	}
	return count, nil
}

func scanFunctions(rows pgx.Rows) ([]Function, error) {
	var functions []Function
	for rows.Next() {
		var fn Function
		var paramsRaw []byte
		if err := rows.Scan(
			&fn.ID, &fn.Name, &fn.Description, &paramsRaw,
			&fn.ReturnType, &fn.Body, &fn.CreatedAt, &fn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanFunctions: %w", err)
		}
		if err := json.Unmarshal(paramsRaw, &fn.Params); err != nil {
			return nil, fmt.Errorf("scanFunctions: unmarshal params: %w", err)
		}
		functions = append(functions, fn)
	}
	return functions, rows.Err()
}
