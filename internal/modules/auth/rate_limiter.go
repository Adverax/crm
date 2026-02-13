package auth

import (
	"sync"
	"time"
)

// RateLimiter provides in-memory sliding window rate limiting per key.
type RateLimiter struct {
	maxAttempts int
	window      time.Duration
	mu          sync.Mutex
	entries     map[string][]time.Time
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxAttempts: maxAttempts,
		window:      window,
		entries:     make(map[string][]time.Time),
	}
	go rl.cleanupLoop()
	return rl
}

// Allow checks if the key is allowed to make a request. Returns false if rate limited.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	attempts := rl.entries[key]
	var valid []time.Time
	for _, t := range attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxAttempts {
		rl.entries[key] = valid
		return false
	}

	rl.entries[key] = append(valid, now)
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for key, attempts := range rl.entries {
		var valid []time.Time
		for _, t := range attempts {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.entries, key)
		} else {
			rl.entries[key] = valid
		}
	}
}
