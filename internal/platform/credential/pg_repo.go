package credential

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for credentials.
type Repository interface {
	Create(ctx context.Context, cred *Credential, encryptedAuth, nonce []byte) error
	GetByID(ctx context.Context, id uuid.UUID) (*Credential, []byte, []byte, error)
	GetByCode(ctx context.Context, code string) (*Credential, []byte, []byte, error)
	ListAll(ctx context.Context) ([]Credential, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateCredentialInput, encryptedAuth, nonce []byte) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error

	// Token cache for OAuth2
	GetToken(ctx context.Context, credentialID uuid.UUID) (encryptedToken, nonce []byte, tokenType string, expiresAt time.Time, err error)
	UpsertToken(ctx context.Context, credentialID uuid.UUID, encryptedToken, nonce []byte, tokenType string, expiresAt time.Time) error
	DeleteToken(ctx context.Context, credentialID uuid.UUID) error

	// Usage log
	LogUsage(ctx context.Context, entry *UsageLogEntry) error
	GetUsageLog(ctx context.Context, credentialID uuid.UUID, limit int) ([]UsageLogEntry, error)
}

// PgRepository is the PostgreSQL implementation of Repository.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PgRepository.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) Create(ctx context.Context, cred *Credential, encryptedAuth, nonce []byte) error {
	query := `
		INSERT INTO metadata.credentials (id, code, name, description, type, base_url, auth_data_encrypted, auth_data_nonce, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		cred.ID, cred.Code, cred.Name, cred.Description,
		string(cred.Type), cred.BaseURL,
		encryptedAuth, nonce,
		cred.IsActive, cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.Create: %w", err)
	}
	return nil
}

func (r *PgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Credential, []byte, []byte, error) {
	query := `
		SELECT id, code, name, description, type, base_url, auth_data_encrypted, auth_data_nonce, is_active, created_at, updated_at
		FROM metadata.credentials
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return r.scanCredentialRow(row)
}

func (r *PgRepository) GetByCode(ctx context.Context, code string) (*Credential, []byte, []byte, error) {
	query := `
		SELECT id, code, name, description, type, base_url, auth_data_encrypted, auth_data_nonce, is_active, created_at, updated_at
		FROM metadata.credentials
		WHERE code = $1
	`
	row := r.pool.QueryRow(ctx, query, code)
	return r.scanCredentialRow(row)
}

func (r *PgRepository) scanCredentialRow(row pgx.Row) (*Credential, []byte, []byte, error) {
	var cred Credential
	var encAuth, nonce []byte
	var credType string

	err := row.Scan(
		&cred.ID, &cred.Code, &cred.Name, &cred.Description,
		&credType, &cred.BaseURL,
		&encAuth, &nonce,
		&cred.IsActive, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("pgCredentialRepo.scan: %w", err)
	}

	cred.Type = CredentialType(credType)
	return &cred, encAuth, nonce, nil
}

func (r *PgRepository) ListAll(ctx context.Context) ([]Credential, error) {
	query := `
		SELECT id, code, name, description, type, base_url, is_active, created_at, updated_at
		FROM metadata.credentials
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pgCredentialRepo.ListAll: %w", err)
	}
	defer rows.Close()

	var credentials []Credential
	for rows.Next() {
		var cred Credential
		var credType string
		if err := rows.Scan(
			&cred.ID, &cred.Code, &cred.Name, &cred.Description,
			&credType, &cred.BaseURL,
			&cred.IsActive, &cred.CreatedAt, &cred.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgCredentialRepo.ListAll scan: %w", err)
		}
		cred.Type = CredentialType(credType)
		credentials = append(credentials, cred)
	}

	return credentials, nil
}

func (r *PgRepository) Update(ctx context.Context, id uuid.UUID, input UpdateCredentialInput, encryptedAuth, nonce []byte) error {
	if encryptedAuth != nil {
		query := `
			UPDATE metadata.credentials
			SET name = $1, description = $2, base_url = $3, auth_data_encrypted = $4, auth_data_nonce = $5, updated_at = now()
			WHERE id = $6
		`
		_, err := r.pool.Exec(ctx, query, input.Name, input.Description, input.BaseURL, encryptedAuth, nonce, id)
		if err != nil {
			return fmt.Errorf("pgCredentialRepo.Update (with auth): %w", err)
		}
	} else {
		query := `
			UPDATE metadata.credentials
			SET name = $1, description = $2, base_url = $3, updated_at = now()
			WHERE id = $4
		`
		_, err := r.pool.Exec(ctx, query, input.Name, input.Description, input.BaseURL, id)
		if err != nil {
			return fmt.Errorf("pgCredentialRepo.Update: %w", err)
		}
	}
	return nil
}

func (r *PgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM metadata.credentials WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	query := `UPDATE metadata.credentials SET is_active = $1, updated_at = now() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.SetActive: %w", err)
	}
	return nil
}

func (r *PgRepository) GetToken(ctx context.Context, credentialID uuid.UUID) ([]byte, []byte, string, time.Time, error) {
	query := `
		SELECT access_token_encrypted, access_token_nonce, token_type, expires_at
		FROM metadata.credential_tokens
		WHERE credential_id = $1
	`
	var encToken, nonce []byte
	var tokenType string
	var expiresAt time.Time

	err := r.pool.QueryRow(ctx, query, credentialID).Scan(&encToken, &nonce, &tokenType, &expiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, "", time.Time{}, nil
		}
		return nil, nil, "", time.Time{}, fmt.Errorf("pgCredentialRepo.GetToken: %w", err)
	}

	return encToken, nonce, tokenType, expiresAt, nil
}

func (r *PgRepository) UpsertToken(ctx context.Context, credentialID uuid.UUID, encryptedToken, nonce []byte, tokenType string, expiresAt time.Time) error {
	query := `
		INSERT INTO metadata.credential_tokens (credential_id, access_token_encrypted, access_token_nonce, token_type, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		ON CONFLICT (credential_id) DO UPDATE SET
			access_token_encrypted = EXCLUDED.access_token_encrypted,
			access_token_nonce = EXCLUDED.access_token_nonce,
			token_type = EXCLUDED.token_type,
			expires_at = EXCLUDED.expires_at,
			updated_at = now()
	`
	_, err := r.pool.Exec(ctx, query, credentialID, encryptedToken, nonce, tokenType, expiresAt)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.UpsertToken: %w", err)
	}
	return nil
}

func (r *PgRepository) DeleteToken(ctx context.Context, credentialID uuid.UUID) error {
	query := `DELETE FROM metadata.credential_tokens WHERE credential_id = $1`
	_, err := r.pool.Exec(ctx, query, credentialID)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.DeleteToken: %w", err)
	}
	return nil
}

func (r *PgRepository) LogUsage(ctx context.Context, entry *UsageLogEntry) error {
	query := `
		INSERT INTO metadata.credential_usage_log (id, credential_id, procedure_code, request_url, response_status, success, error_message, duration_ms, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.CredentialID, entry.ProcedureCode,
		entry.RequestURL, entry.ResponseStatus,
		entry.Success, entry.ErrorMessage,
		entry.DurationMs, entry.UserID, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("pgCredentialRepo.LogUsage: %w", err)
	}
	return nil
}

func (r *PgRepository) GetUsageLog(ctx context.Context, credentialID uuid.UUID, limit int) ([]UsageLogEntry, error) {
	query := `
		SELECT id, credential_id, procedure_code, request_url, response_status, success, error_message, duration_ms, user_id, created_at
		FROM metadata.credential_usage_log
		WHERE credential_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, credentialID, limit)
	if err != nil {
		return nil, fmt.Errorf("pgCredentialRepo.GetUsageLog: %w", err)
	}
	defer rows.Close()

	var entries []UsageLogEntry
	for rows.Next() {
		var e UsageLogEntry
		if err := rows.Scan(
			&e.ID, &e.CredentialID, &e.ProcedureCode,
			&e.RequestURL, &e.ResponseStatus,
			&e.Success, &e.ErrorMessage,
			&e.DurationMs, &e.UserID, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgCredentialRepo.GetUsageLog scan: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, nil
}
