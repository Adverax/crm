package auth

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		maxAttempts int
		window      time.Duration
		calls       int
		wantAllowed int
	}{
		{name: "allows under limit", maxAttempts: 3, window: time.Minute, calls: 2, wantAllowed: 2},
		{name: "blocks at limit", maxAttempts: 3, window: time.Minute, calls: 5, wantAllowed: 3},
		{name: "single attempt allowed", maxAttempts: 1, window: time.Minute, calls: 3, wantAllowed: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rl := &RateLimiter{
				maxAttempts: tt.maxAttempts,
				window:      tt.window,
				entries:     make(map[string][]time.Time),
			}

			allowed := 0
			for i := 0; i < tt.calls; i++ {
				if rl.Allow("test-key") {
					allowed++
				}
			}

			if allowed != tt.wantAllowed {
				t.Errorf("Allow() allowed %d, want %d", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	t.Parallel()

	rl := &RateLimiter{
		maxAttempts: 2,
		window:      50 * time.Millisecond,
		entries:     make(map[string][]time.Time),
	}

	rl.Allow("key")
	rl.Allow("key")

	if rl.Allow("key") {
		t.Error("should be blocked at limit")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("key") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()

	rl := &RateLimiter{
		maxAttempts: 5,
		window:      50 * time.Millisecond,
		entries:     make(map[string][]time.Time),
	}

	rl.Allow("key1")
	rl.Allow("key2")

	time.Sleep(60 * time.Millisecond)
	rl.cleanup()

	rl.mu.Lock()
	count := len(rl.entries)
	rl.mu.Unlock()

	if count != 0 {
		t.Errorf("cleanup() left %d entries, want 0", count)
	}
}
