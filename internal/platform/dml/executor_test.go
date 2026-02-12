package dml

import (
	"testing"
)

func TestInjectDMLRLSClause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sql        string
		params     []any
		rlsClause  string
		rlsParams  []any
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "UPDATE with existing WHERE",
			sql:        "UPDATE public.obj_account SET name = $1 WHERE id = $2 RETURNING id",
			params:     []any{"NewName", "abc-123"},
			rlsClause:  "owner_id = $1",
			rlsParams:  []any{"user-1"},
			wantSQL:    "UPDATE public.obj_account SET name = $1 WHERE owner_id = $3 AND id = $2 RETURNING id",
			wantParams: []any{"NewName", "abc-123", "user-1"},
		},
		{
			name:       "DELETE with existing WHERE",
			sql:        "DELETE FROM public.obj_account WHERE id = $1 RETURNING id",
			params:     []any{"abc-123"},
			rlsClause:  "owner_id = $1",
			rlsParams:  []any{"user-1"},
			wantSQL:    "DELETE FROM public.obj_account WHERE owner_id = $2 AND id = $1 RETURNING id",
			wantParams: []any{"abc-123", "user-1"},
		},
		{
			name:       "DELETE without WHERE",
			sql:        "DELETE FROM public.obj_account RETURNING id",
			params:     nil,
			rlsClause:  "owner_id = $1",
			rlsParams:  []any{"user-1"},
			wantSQL:    "DELETE FROM public.obj_account WHERE owner_id = $1 RETURNING id",
			wantParams: []any{"user-1"},
		},
		{
			name:       "UPDATE without WHERE but with RETURNING",
			sql:        "UPDATE public.obj_account SET name = $1 RETURNING id",
			params:     []any{"NewName"},
			rlsClause:  "owner_id IN ($1, $2)",
			rlsParams:  []any{"user-1", "user-2"},
			wantSQL:    "UPDATE public.obj_account SET name = $1 WHERE owner_id IN ($2, $3) RETURNING id",
			wantParams: []any{"NewName", "user-1", "user-2"},
		},
		{
			name:       "multiple RLS params renumbered correctly",
			sql:        "UPDATE public.obj_contact SET email = $1 WHERE id = $2 RETURNING id",
			params:     []any{"a@b.com", "rec-1"},
			rlsClause:  "(owner_id = $1 OR id IN (SELECT record_id FROM obj_contact__share WHERE group_id = $2))",
			rlsParams:  []any{"user-1", "group-1"},
			wantSQL:    "UPDATE public.obj_contact SET email = $1 WHERE (owner_id = $3 OR id IN (SELECT record_id FROM obj_contact__share WHERE group_id = $4)) AND id = $2 RETURNING id",
			wantParams: []any{"a@b.com", "rec-1", "user-1", "group-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotSQL, gotParams := injectDMLRLSClause(tt.sql, tt.params, tt.rlsClause, tt.rlsParams)
			if gotSQL != tt.wantSQL {
				t.Errorf("SQL mismatch\n got: %s\nwant: %s", gotSQL, tt.wantSQL)
			}
			if len(gotParams) != len(tt.wantParams) {
				t.Errorf("params length mismatch: got %d, want %d", len(gotParams), len(tt.wantParams))
				return
			}
			for i := range gotParams {
				if gotParams[i] != tt.wantParams[i] {
					t.Errorf("params[%d] mismatch: got %v, want %v", i, gotParams[i], tt.wantParams[i])
				}
			}
		})
	}
}
