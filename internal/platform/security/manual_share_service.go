package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type manualShareServiceImpl struct {
	pool *pgxpool.Pool
}

// NewManualShareService creates a new ManualShareService.
func NewManualShareService(pool *pgxpool.Pool) ManualShareService {
	return &manualShareServiceImpl{pool: pool}
}

func (s *manualShareServiceImpl) ShareRecord(ctx context.Context, tableName string, input ShareRecordInput) (*RecordShare, error) {
	if input.AccessLevel != "read" && input.AccessLevel != "read_write" {
		return nil, fmt.Errorf("manualShareService.ShareRecord: %w",
			apperror.Validation("access_level must be 'read' or 'read_write'"))
	}

	shareTable := quoteShareIdent(tableName + "__share")

	query := fmt.Sprintf(`
		INSERT INTO %s (record_id, group_id, access_level, reason)
		VALUES ($1, $2, $3, 'manual')
		ON CONFLICT ON CONSTRAINT uq_%s_record_group_reason
		DO UPDATE SET access_level = EXCLUDED.access_level
		RETURNING id, record_id, group_id, access_level, reason, created_at
	`, shareTable, sanitizeShareIndex(tableName+"__share"))

	var share RecordShare
	err := s.pool.QueryRow(ctx, query,
		input.RecordID, input.GroupID, input.AccessLevel,
	).Scan(
		&share.ID, &share.RecordID, &share.GroupID,
		&share.AccessLevel, &share.Reason, &share.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("manualShareService.ShareRecord: %w", err)
	}

	return &share, nil
}

func (s *manualShareServiceImpl) RevokeShare(ctx context.Context, tableName string, input RevokeShareInput) error {
	shareTable := quoteShareIdent(tableName + "__share")

	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE record_id = $1 AND group_id = $2 AND reason = 'manual'
	`, shareTable)

	tag, err := s.pool.Exec(ctx, query, input.RecordID, input.GroupID)
	if err != nil {
		return fmt.Errorf("manualShareService.RevokeShare: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("manualShareService.RevokeShare: %w",
			apperror.NotFound("RecordShare", fmt.Sprintf("record=%s group=%s", input.RecordID, input.GroupID)))
	}

	return nil
}

func (s *manualShareServiceImpl) ListShares(ctx context.Context, tableName string, recordID uuid.UUID) ([]RecordShare, error) {
	shareTable := quoteShareIdent(tableName + "__share")

	query := fmt.Sprintf(`
		SELECT id, record_id, group_id, access_level, reason, created_at
		FROM %s
		WHERE record_id = $1
		ORDER BY created_at
	`, shareTable)

	rows, err := s.pool.Query(ctx, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("manualShareService.ListShares: %w", err)
	}
	defer rows.Close()

	var shares []RecordShare
	for rows.Next() {
		var share RecordShare
		if err := rows.Scan(
			&share.ID, &share.RecordID, &share.GroupID,
			&share.AccessLevel, &share.Reason, &share.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("manualShareService.ListShares: scan: %w", err)
		}
		shares = append(shares, share)
	}

	return shares, rows.Err()
}

func quoteShareIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func sanitizeShareIndex(name string) string {
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
