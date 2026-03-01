package dml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statements []string
		want       []DMLTargetInfo
	}{
		{
			name:       "INSERT extracts object and fields",
			statements: []string{"INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech')"},
			want: []DMLTargetInfo{
				{Object: "Account", Fields: []string{"Name", "Industry"}, Operation: "insert"},
			},
		},
		{
			name:       "UPDATE extracts object and SET fields",
			statements: []string{"UPDATE Contact SET Status = 'Active', Email = 'test@test.com' WHERE Id = '123'"},
			want: []DMLTargetInfo{
				{Object: "Contact", Fields: []string{"Status", "Email"}, Operation: "update"},
			},
		},
		{
			name:       "DELETE extracts object with no fields",
			statements: []string{"DELETE FROM Task WHERE Status = 'Done'"},
			want: []DMLTargetInfo{
				{Object: "Task", Fields: nil, Operation: "delete"},
			},
		},
		{
			name:       "UPSERT extracts object and fields",
			statements: []string{"UPSERT Account (external_id, Name, Industry) VALUES ('ext1', 'Acme', 'Tech') ON external_id"},
			want: []DMLTargetInfo{
				{Object: "Account", Fields: []string{"external_id", "Name", "Industry"}, Operation: "upsert"},
			},
		},
		{
			name: "multiple statements",
			statements: []string{
				"INSERT INTO Account (Name) VALUES ('Acme')",
				"UPDATE Contact SET Status = 'Active'",
			},
			want: []DMLTargetInfo{
				{Object: "Account", Fields: []string{"Name"}, Operation: "insert"},
				{Object: "Contact", Fields: []string{"Status"}, Operation: "update"},
			},
		},
		{
			name:       "invalid statement is skipped",
			statements: []string{"NOT A VALID DML", "INSERT INTO Account (Name) VALUES ('Acme')"},
			want: []DMLTargetInfo{
				{Object: "Account", Fields: []string{"Name"}, Operation: "insert"},
			},
		},
		{
			name:       "empty input",
			statements: nil,
			want:       nil,
		},
		{
			name:       "all invalid",
			statements: []string{"garbage", "more garbage"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractTargets(tt.statements)
			assert.Equal(t, tt.want, got)
		})
	}
}
