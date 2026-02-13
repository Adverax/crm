package testutil

import "github.com/google/uuid"

// Stable UUIDs for reproducible tests. Never use random UUIDs in test fixtures.
var (
	AdminUserID    = uuid.MustParse("00000000-0000-4000-a000-000000000001")
	AdminProfileID = uuid.MustParse("00000000-0000-4000-a000-000000000002")
	AdminRoleID    = uuid.MustParse("00000000-0000-4000-a000-000000000003")
	TestUserID     = uuid.MustParse("00000000-0000-4000-a000-000000000011")
	TestProfileID  = uuid.MustParse("00000000-0000-4000-a000-000000000012")
	TestRoleID     = uuid.MustParse("00000000-0000-4000-a000-000000000013")
	TestObjectID   = uuid.MustParse("00000000-0000-4000-a000-000000000021")
	TestFieldID    = uuid.MustParse("00000000-0000-4000-a000-000000000022")
)
