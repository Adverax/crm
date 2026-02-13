//go:build integration

package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/testutil"
)

func TestPgPasswordResetRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.password_reset_tokens", "iam.users", "iam.profiles", "iam.permission_sets")

	repo := NewPgPasswordResetTokenRepository(pool)
	ctx := context.Background()

	t.Run("Create and GetByTokenHash", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		token := &PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "reset_hash_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
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
		assert.Nil(t, fetched.UsedAt)
		assert.WithinDuration(t, token.CreatedAt, fetched.CreatedAt, time.Second)
	})

	t.Run("GetByTokenHash returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByTokenHash(ctx, "nonexistent_reset_hash_"+uuid.NewString()[:8])
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("MarkUsed sets used_at", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		token := &PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "reset_mark_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
		}

		err := repo.Create(ctx, token)
		require.NoError(t, err)

		err = repo.MarkUsed(ctx, token.ID)
		require.NoError(t, err)

		fetched, err := repo.GetByTokenHash(ctx, token.TokenHash)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.NotNil(t, fetched.UsedAt)
		assert.WithinDuration(t, time.Now().UTC(), *fetched.UsedAt, 5*time.Second)
	})

	t.Run("DeleteExpired removes only expired tokens", func(t *testing.T) {
		userID := setupTestUser(t, pool)
		now := time.Now().UTC().Truncate(time.Microsecond)

		expiredToken := &PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "reset_expired_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(-1 * time.Hour),
			CreatedAt: now.Add(-2 * time.Hour),
		}
		err := repo.Create(ctx, expiredToken)
		require.NoError(t, err)

		validToken := &PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "reset_valid_" + uuid.NewString()[:8],
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
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
