package procedure

import (
	"context"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// NotificationCommandExecutor is a stub for notification.* commands.
type NotificationCommandExecutor struct{}

// NewNotificationCommandExecutor creates a stub NotificationCommandExecutor.
func NewNotificationCommandExecutor() *NotificationCommandExecutor {
	return &NotificationCommandExecutor{}
}

func (e *NotificationCommandExecutor) Category() string { return "notification" }

func (e *NotificationCommandExecutor) Execute(_ context.Context, cmd metadata.CommandDef, _ *ExecutionContext) (any, error) {
	return nil, apperror.BadRequest("notification.* commands are not yet implemented (command: " + cmd.Type + ")")
}
