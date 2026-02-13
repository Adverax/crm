package auth

import (
	"context"
	"log/slog"
)

// ConsoleEmailSender logs password reset emails to stdout for development.
type ConsoleEmailSender struct{}

// NewConsoleEmailSender creates a new ConsoleEmailSender.
func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{}
}

func (s *ConsoleEmailSender) SendPasswordReset(_ context.Context, email, resetURL string) error {
	slog.Info("password reset email",
		"to", email,
		"reset_url", resetURL,
	)
	return nil
}
