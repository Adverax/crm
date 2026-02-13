package auth

import "context"

// EmailSender sends transactional emails.
type EmailSender interface {
	SendPasswordReset(ctx context.Context, email, resetURL string) error
}
