package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	"github.com/proxima-research/proxima.crm.platform/pkg/keyset"
)

// CursorHandler handles cursor encoding/decoding for SOQL pagination.
// It wraps pkg/keyset to provide SOQL-specific cursor functionality.
type CursorHandler struct {
	cursor *keyset.Cursor
}

// NewCursorHandler creates a new cursor handler with the given secret.
func NewCursorHandler(secret engine.SecretProvider) *CursorHandler {
	// Adapt engine.SecretProvider to keyset.SecretProvider
	return &CursorHandler{
		cursor: keyset.NewCursor(
			secretProviderAdapter{secret},
			keyset.WithTieBreaker(engine.DefaultTieBreaker),
		),
	}
}

// secretProviderAdapter adapts engine.SecretProvider to keyset.SecretProvider.
type secretProviderAdapter struct {
	provider engine.SecretProvider
}

func (a secretProviderAdapter) Secret() []byte {
	return a.provider.Secret()
}

// PaginationContext holds pagination context for query execution.
type PaginationContext struct {
	// PageSize is the number of records to fetch.
	PageSize int

	// Cursor is the decoded cursor payload (nil for first page).
	Cursor *keyset.CursorPayload

	// UserID is the user ID for FID generation.
	UserID int64

	// SortKeys are the normalized sort keys.
	SortKeys engine.SortKeys
}

// DecodeCursor decodes and validates a cursor for the given query context.
func (h *CursorHandler) DecodeCursor(
	cursorStr string,
	pagination *engine.PaginationInfo,
	userID int64,
) (*PaginationContext, error) {
	ctx := &PaginationContext{
		PageSize: pagination.PageSize,
		UserID:   userID,
		SortKeys: pagination.SortKeys,
	}

	if cursorStr == "" {
		return ctx, nil
	}

	// Build FID for validation
	fid := keyset.BuildFID(pagination.Object, userID, nil)

	// Decode cursor
	payload, err := h.cursor.Decode(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	// Validate context - convert engine.SortKeys to keyset.SortKeys
	keysetSortKeys := toKeysetSortKeys(pagination.SortKeys)
	if err := h.cursor.ValidateContext(payload, fid, keysetSortKeys); err != nil {
		return nil, fmt.Errorf("cursor context mismatch: %w", err)
	}

	ctx.Cursor = payload
	return ctx, nil
}

// BuildNextCursor builds the next cursor from the last record.
func (h *CursorHandler) BuildNextCursor(
	lastRecord *Record,
	pagination *engine.PaginationInfo,
	userID int64,
) (string, error) {
	if lastRecord == nil || len(lastRecord.Fields) == 0 {
		return "", nil
	}

	// Extract values for sort key fields
	lastRow := make(map[string]interface{})
	for _, sk := range pagination.SortKeys {
		// Get value from record fields
		// The field name in the record might be the SOQL name, not SQL column
		value := h.extractSortKeyValue(lastRecord, sk.Field, pagination)
		if value != nil {
			lastRow[sk.Field] = h.formatValueForCursor(value)
		}
	}

	if len(lastRow) == 0 {
		return "", nil
	}

	// Convert engine.SortKeys to keyset.SortKeys
	keysetSortKeys := toKeysetSortKeys(pagination.SortKeys)

	return h.cursor.Next(
		lastRow,
		pagination.Object,
		userID,
		nil, // No filter for SOQL (filter is part of query)
		keysetSortKeys,
	)
}

// toKeysetSortKeys converts engine.SortKeys to keyset.SortKeys.
func toKeysetSortKeys(keys engine.SortKeys) keyset.SortKeys {
	result := make(keyset.SortKeys, len(keys))
	for i, k := range keys {
		result[i] = keyset.SortKey{
			Field: k.Field,
			Dir:   string(k.Dir),
		}
	}
	return result
}

// extractSortKeyValue extracts the value for a sort key from the record.
func (h *CursorHandler) extractSortKeyValue(rec *Record, sqlColumn string, pagination *engine.PaginationInfo) any {
	// First try to find by SQL column name directly
	if val, ok := rec.Fields[sqlColumn]; ok {
		return val
	}

	// Try to find by SOQL name mapping
	for i, sk := range pagination.SortKeys {
		if sk.Field == sqlColumn && i < len(pagination.SortKeySOQL) {
			soqlName := pagination.SortKeySOQL[i]
			if val, ok := rec.Fields[soqlName]; ok {
				return val
			}
		}
	}

	return nil
}

// formatValueForCursor formats a value for inclusion in cursor.
func (h *CursorHandler) formatValueForCursor(value any) any {
	switch v := value.(type) {
	case time.Time:
		return v.Format(time.RFC3339Nano)
	default:
		return value
	}
}

// BuildKeysetWhereClause builds a keyset WHERE clause from cursor.
func (h *CursorHandler) BuildKeysetWhereClause(
	cursor *keyset.CursorPayload,
	pagination *engine.PaginationInfo,
	mainAlias string,
) (string, []any, int, error) {
	if cursor == nil || cursor.LastRow == nil {
		return "", nil, 0, nil
	}

	// Build keyset comparison
	// For same-direction keys: (col1, col2) < (val1, val2)
	// For mixed-direction: expanded OR/AND form
	return h.buildKeysetComparison(cursor.LastRow, pagination, mainAlias)
}

// buildKeysetComparison builds the SQL WHERE clause for keyset pagination.
func (h *CursorHandler) buildKeysetComparison(
	lastRow map[string]interface{},
	pagination *engine.PaginationInfo,
	mainAlias string,
) (string, []any, int, error) {
	if len(pagination.SortKeys) == 0 {
		return "", nil, 0, nil
	}

	// Check if all directions are the same
	allSameDir := true
	firstDir := strings.ToLower(string(pagination.SortKeys[0].Dir))
	for _, sk := range pagination.SortKeys[1:] {
		if strings.ToLower(string(sk.Dir)) != firstDir {
			allSameDir = false
			break
		}
	}

	if allSameDir {
		return h.buildTupleComparison(lastRow, pagination, mainAlias, firstDir)
	}

	return h.buildExpandedComparison(lastRow, pagination, mainAlias)
}

// buildTupleComparison builds efficient tuple comparison for same-direction keys.
// Example: (created_at, record_id) < ($1, $2)
func (h *CursorHandler) buildTupleComparison(
	lastRow map[string]interface{},
	pagination *engine.PaginationInfo,
	mainAlias string,
	direction string,
) (string, []any, int, error) {
	var columns []string
	var params []any
	paramNum := 1

	for _, sk := range pagination.SortKeys {
		col := fmt.Sprintf("%s.%s", mainAlias, sk.Field)
		columns = append(columns, col)

		val, ok := lastRow[sk.Field]
		if !ok {
			return "", nil, 0, fmt.Errorf("missing sort key value: %s", sk.Field)
		}
		params = append(params, val)
		paramNum++
	}

	op := "<"
	if strings.ToLower(direction) == string(engine.SortAsc) {
		op = ">"
	}

	// Build parameter placeholders
	placeholders := make([]string, len(params))
	for i := range params {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	whereClause := fmt.Sprintf("(%s) %s (%s)",
		strings.Join(columns, ", "),
		op,
		strings.Join(placeholders, ", "),
	)

	return whereClause, params, len(params), nil
}

// buildExpandedComparison builds expanded OR/AND form for mixed-direction keys.
// Example: (col1 < $1 OR (col1 = $1 AND col2 > $2))
func (h *CursorHandler) buildExpandedComparison(
	lastRow map[string]interface{},
	pagination *engine.PaginationInfo,
	mainAlias string,
) (string, []any, int, error) {
	var orClauses []string
	var params []any
	paramNum := 1

	// Collect all values first
	values := make([]any, len(pagination.SortKeys))
	for i, sk := range pagination.SortKeys {
		val, ok := lastRow[sk.Field]
		if !ok {
			return "", nil, 0, fmt.Errorf("missing sort key value: %s", sk.Field)
		}
		values[i] = val
	}

	for i, sk := range pagination.SortKeys {
		var andParts []string

		// Add equality conditions for all previous columns
		for j := 0; j < i; j++ {
			col := fmt.Sprintf("%s.%s", mainAlias, pagination.SortKeys[j].Field)
			andParts = append(andParts, fmt.Sprintf("%s = $%d", col, paramNum))
			params = append(params, values[j])
			paramNum++
		}

		// Add comparison for current column
		col := fmt.Sprintf("%s.%s", mainAlias, sk.Field)
		op := "<"
		if strings.ToLower(string(sk.Dir)) == string(engine.SortAsc) {
			op = ">"
		}
		andParts = append(andParts, fmt.Sprintf("%s %s $%d", col, op, paramNum))
		params = append(params, values[i])
		paramNum++

		if len(andParts) == 1 {
			orClauses = append(orClauses, andParts[0])
		} else {
			orClauses = append(orClauses, "("+strings.Join(andParts, " AND ")+")")
		}
	}

	if len(orClauses) == 1 {
		return orClauses[0], params, len(params), nil
	}

	return "(" + strings.Join(orClauses, " OR ") + ")", params, len(params), nil
}
