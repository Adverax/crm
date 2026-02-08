package security

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgOutboxRepository implements OutboxRepository using pgx.
type PgOutboxRepository struct {
	pool *pgxpool.Pool
}

// NewPgOutboxRepository creates a new PgOutboxRepository.
func NewPgOutboxRepository(pool *pgxpool.Pool) *PgOutboxRepository {
	return &PgOutboxRepository{pool: pool}
}

func (r *PgOutboxRepository) Insert(ctx context.Context, tx pgx.Tx, event OutboxEvent) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO security.security_outbox (event_type, entity_type, entity_id, payload)
		VALUES ($1, $2, $3, $4)
	`, event.EventType, event.EntityType, event.EntityID, event.Payload)
	if err != nil {
		return fmt.Errorf("pgOutboxRepo.Insert: %w", err)
	}
	return nil
}

func (r *PgOutboxRepository) ListUnprocessed(ctx context.Context, limit int) ([]OutboxEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, event_type, entity_type, entity_id, payload, created_at, processed_at
		FROM security.security_outbox
		WHERE processed_at IS NULL
		ORDER BY created_at
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("pgOutboxRepo.ListUnprocessed: %w", err)
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(
			&e.ID, &e.EventType, &e.EntityType, &e.EntityID,
			&e.Payload, &e.CreatedAt, &e.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("pgOutboxRepo.ListUnprocessed: scan: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *PgOutboxRepository) MarkProcessed(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE security.security_outbox
		SET processed_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("pgOutboxRepo.MarkProcessed: %w", err)
	}
	return nil
}
