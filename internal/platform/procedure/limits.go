package procedure

import "time"

// Execution limits (ADR-0024).
const (
	MaxExecutionTimeout = 30 * time.Second
	MaxCommands         = 50
	MaxCallDepth        = 3
	MaxNestingDepth     = 5
	MaxInputSize        = 1 << 20 // 1MB
	MaxHTTPCalls        = 10
	MaxNotifications    = 10
	MaxRetryAttempts    = 5
	MaxRetryDelayMs     = 60000 // 60s
	MinRetryDelayMs     = 100   // 100ms
)
