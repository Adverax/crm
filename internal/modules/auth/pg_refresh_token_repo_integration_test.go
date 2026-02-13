//go:build integration

package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/testutil"
)

func setupTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	psID := uuid.New()
	profileID := uuid.New()
	userID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO iam.permission_sets (id, api_name, label, ps_type) VALUES ($1, $2, $3, $4)`,
		psID, "test_ps_"+psID.String()[:8], "Test PS", "grant",
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO iam.profiles (id, api_name, label, base_permission_set_id) VALUES ($1, $2, $3, $4)`,
		profileID, "test_profile_"+profileID.String()[:8], "Test Profile", psID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO iam.users (id, username, email, first_name, last_name, profile_id) VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, "testuser_"+userID.String()[:8], "test_"+userID.String()[:8]+"@example.com", "Test", "User", profileID,
	)
	require.NoError(t, err)

	return userID
}

func TestPgRefreshTokenRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.refresh_tokens", "iam.users", "iam.profiles", "iam.permission_sets")

	repo := NewPgRefreshTokenRepository(pool)
	ctx := context.Background()

	t.Run("Create and GetByTokenHash", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		token := &RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "hash_create_get_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := repo.Create(ctx, token)
		require.NoError(t, err)

		fetched, err := repo.GetByTokenHash(ctx, token.TokenHash)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, token.ID, fetched.ID)
		assert.Equal(t, token.UserID, fetched.UserID)
		assert.Equal(t, token.TokenHash, fetched.TokenHash)
		assert.WithinDuration(t, token.ExpiresAt, fetched.ExpiresAt, time.Second)
		assert.WithinDuration(t, token.CreatedAt, fetched.CreatedAt, time.Second)
		assert.WithinDuration(t, token.UpdatedAt, fetched.UpdatedAt, time.Second)
	})

	t.Run("GetByTokenHash returns nil for non-existent hash", func(t *testing.T) {
		fetched, err := repo.GetByTokenHash(ctx, "nonexistent_hash_"+uuid.NewString()[:8])
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("DeleteByTokenHash removes the token", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		token := &RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "hash_delete_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := repo.Create(ctx, token)
		require.NoError(t, err)

		err = repo.DeleteByTokenHash(ctx, token.TokenHash)
		require.NoError(t, err)

		fetched, err := repo.GetByTokenHash(ctx, token.TokenHash)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("DeleteAllByUserID removes all tokens for user", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		for i := 0; i < 3; i++ {
			token := &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash_delall_" + uuid.NewString()[:8],
				ExpiresAt: now.Add(7 * 24 * time.Hour),
				CreatedAt: now,
				UpdatedAt: now,
			}
			err := repo.Create(ctx, token)
			require.NoError(t, err)
		}

		err := repo.DeleteAllByUserID(ctx, userID)
		require.NoError(t, err)

		// Verify all are gone by trying to find any token for this user
		var count int
		err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.refresh_tokens WHERE user_id = $1`, userID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("DeleteExpired removes only expired tokens", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		expiredToken := &RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "hash_expired_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(-1 * time.Hour),
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		}
		err := repo.Create(ctx, expiredToken)
		require.NoError(t, err)

		validToken := &RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "hash_valid_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = repo.Create(ctx, validToken)
		require.NoError(t, err)

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Expired token is gone
		fetched, err := repo.GetByTokenHash(ctx, expiredToken.TokenHash)
		require.NoError(t, err)
		assert.Nil(t, fetched)

		// Valid token still exists
		fetched, err = repo.GetByTokenHash(ctx, validToken.TokenHash)
		require.NoError(t, err)
		assert.NotNil(t, fetched)
	})
}
