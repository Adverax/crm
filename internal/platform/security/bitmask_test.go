package security

import "testing"

func TestHasOLS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		permissions int
		flag        int
		want        bool
	}{
		{name: "read when has read", permissions: OLSRead, flag: OLSRead, want: true},
		{name: "create when has all", permissions: OLSAll, flag: OLSCreate, want: true},
		{name: "delete when has none", permissions: 0, flag: OLSDelete, want: false},
		{name: "update when has read only", permissions: OLSRead, flag: OLSUpdate, want: false},
		{name: "read when has read+create", permissions: OLSRead | OLSCreate, flag: OLSRead, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HasOLS(tt.permissions, tt.flag)
			if got != tt.want {
				t.Errorf("HasOLS(%d, %d) = %v, want %v", tt.permissions, tt.flag, got, tt.want)
			}
		})
	}
}

func TestHasFLS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		permissions int
		flag        int
		want        bool
	}{
		{name: "read when has read", permissions: FLSRead, flag: FLSRead, want: true},
		{name: "write when has all", permissions: FLSAll, flag: FLSWrite, want: true},
		{name: "write when has read only", permissions: FLSRead, flag: FLSWrite, want: false},
		{name: "read when has none", permissions: 0, flag: FLSRead, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HasFLS(tt.permissions, tt.flag)
			if got != tt.want {
				t.Errorf("HasFLS(%d, %d) = %v, want %v", tt.permissions, tt.flag, got, tt.want)
			}
		})
	}
}

func TestComputeEffective(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		grants []int
		denies []int
		want   int
	}{
		{
			name:   "single grant no deny",
			grants: []int{OLSRead | OLSCreate},
			denies: nil,
			want:   OLSRead | OLSCreate,
		},
		{
			name:   "multiple grants OR together",
			grants: []int{OLSRead, OLSCreate},
			denies: nil,
			want:   OLSRead | OLSCreate,
		},
		{
			name:   "deny removes specific permissions",
			grants: []int{OLSAll},
			denies: []int{OLSDelete},
			want:   OLSRead | OLSCreate | OLSUpdate,
		},
		{
			name:   "deny all removes everything",
			grants: []int{OLSAll},
			denies: []int{OLSAll},
			want:   0,
		},
		{
			name:   "no grants results in zero",
			grants: nil,
			denies: []int{OLSRead},
			want:   0,
		},
		{
			name:   "multiple denies OR together",
			grants: []int{OLSAll},
			denies: []int{OLSDelete, OLSCreate},
			want:   OLSRead | OLSUpdate,
		},
		{
			name:   "FLS grant deny combo",
			grants: []int{FLSAll},
			denies: []int{FLSWrite},
			want:   FLSRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeEffective(tt.grants, tt.denies)
			if got != tt.want {
				t.Errorf("ComputeEffective(%v, %v) = %d, want %d", tt.grants, tt.denies, got, tt.want)
			}
		})
	}
}
