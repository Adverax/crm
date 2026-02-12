//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgObjectDefinitionLookup implements ObjectDefinitionLookup using pgx.
type PgObjectDefinitionLookup struct {
	pool *pgxpool.Pool
}

// NewPgObjectDefinitionLookup creates a new PgObjectDefinitionLookup.
func NewPgObjectDefinitionLookup(pool *pgxpool.Pool) *PgObjectDefinitionLookup {
	return &PgObjectDefinitionLookup{pool: pool}
}

// GetTableName returns the table_name for a given object definition ID.
func (l *PgObjectDefinitionLookup) GetTableName(ctx context.Context, objectID uuid.UUID) (string, error) {
	var tableName string
	err := l.pool.QueryRow(ctx,
		`SELECT table_name FROM metadata.object_definitions WHERE id = $1`,
		objectID,
	).Scan(&tableName)
	if err != nil {
		return "", fmt.Errorf("pgObjectDefinitionLookup.GetTableName: %w", err)
	}
	return tableName, nil
}
