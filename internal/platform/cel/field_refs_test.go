package cel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRecordFieldRefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		want       []string
	}{
		{
			name:       "single field reference",
			expression: "record.name != ''",
			want:       []string{"name"},
		},
		{
			name:       "multiple field references",
			expression: "record.end_date > record.start_date",
			want:       []string{"end_date", "start_date"},
		},
		{
			name:       "duplicate field references deduplicated",
			expression: "record.name != '' && size(record.name) > 3",
			want:       []string{"name"},
		},
		{
			name:       "no record references",
			expression: "1 + 2 > 0",
			want:       nil,
		},
		{
			name:       "old references not captured",
			expression: "old.status != record.status",
			want:       []string{"status"},
		},
		{
			name:       "user references not captured",
			expression: "user.role == 'admin' && record.active == true",
			want:       []string{"active"},
		},
		{
			name:       "nested expressions",
			expression: "record.amount > 0 && (record.status == 'open' || record.status == 'pending')",
			want:       []string{"amount", "status"},
		},
		{
			name:       "parse error returns nil",
			expression: "this is not valid CEL %%%",
			want:       nil,
		},
		{
			name:       "empty expression returns nil",
			expression: "",
			want:       nil,
		},
		{
			name:       "function calls with record fields",
			expression: "size(record.description) > 0 && record.name.startsWith('A')",
			want:       []string{"description", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractRecordFieldRefs(tt.expression)
			assert.Equal(t, tt.want, got)
		})
	}
}
