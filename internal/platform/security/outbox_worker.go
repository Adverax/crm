package security

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OutboxWorker processes security outbox events via LISTEN/NOTIFY.
type OutboxWorker struct {
	connConfig  pgx.ConnConfig
	outboxRepo  OutboxRepository
	computer    EffectiveComputer
	rlsComputer RLSEffectiveComputer
	logger      *slog.Logger
}

// NewOutboxWorker creates a new OutboxWorker.
func NewOutboxWorker(
	connConfig pgx.ConnConfig,
	outboxRepo OutboxRepository,
	computer EffectiveComputer,
	rlsComputer RLSEffectiveComputer,
	logger *slog.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		connConfig:  connConfig,
		outboxRepo:  outboxRepo,
		computer:    computer,
		rlsComputer: rlsComputer,
		logger:      logger,
	}
}

// Run starts the outbox worker loop. Blocks until ctx is cancelled.
func (w *OutboxWorker) Run(ctx context.Context) error {
	for {
		if err := w.runLoop(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			w.logger.Error("outbox worker loop failed, reconnecting", "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func (w *OutboxWorker) runLoop(ctx context.Context) error {
	conn, err := pgx.ConnectConfig(ctx, &w.connConfig)
	if err != nil {
		return fmt.Errorf("outboxWorker.runLoop: connect: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "LISTEN security_outbox")
	if err != nil {
		return fmt.Errorf("outboxWorker.runLoop: LISTEN: %w", err)
	}

	w.logger.Info("outbox worker listening for notifications")

	// Process any unprocessed events on startup
	if err := w.processUnprocessed(ctx); err != nil {
		w.logger.Error("outbox worker: initial sweep failed", "error", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processUnprocessed(ctx); err != nil {
				w.logger.Error("outbox worker: fallback sweep failed", "error", err)
			}
		default:
		}

		waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := conn.WaitForNotification(waitCtx)
		cancel()

		if err != nil {
			if waitCtx.Err() != nil && ctx.Err() == nil {
				// Timeout on WaitForNotification — normal, continue loop
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("outboxWorker.runLoop: wait notification: %w", err)
		}

		if err := w.processUnprocessed(ctx); err != nil {
			w.logger.Error("outbox worker: process after notify failed", "error", err)
		}
	}
}

func (w *OutboxWorker) processUnprocessed(ctx context.Context) error {
	events, err := w.outboxRepo.ListUnprocessed(ctx, 100)
	if err != nil {
		return fmt.Errorf("outboxWorker.processUnprocessed: list: %w", err)
	}

	for _, event := range events {
		if err := w.dispatch(ctx, event); err != nil {
			w.logger.Error("outbox worker: dispatch failed",
				"event_id", event.ID,
				"event_type", event.EventType,
				"entity_id", event.EntityID,
				"error", err,
			)
			continue
		}

		if err := w.outboxRepo.MarkProcessed(ctx, event.ID); err != nil {
			w.logger.Error("outbox worker: mark processed failed",
				"event_id", event.ID, "error", err,
			)
		}
	}

	return nil
}

func (w *OutboxWorker) dispatch(ctx context.Context, event OutboxEvent) error {
	switch event.EventType {
	case "user_changed":
		if err := w.computer.RecomputeForUser(ctx, event.EntityID); err != nil {
			return err
		}
		if w.rlsComputer != nil {
			return w.rlsComputer.RecomputeVisibleOwnersForUser(ctx, event.EntityID)
		}
		return nil
	case "permission_set_changed":
		return w.computer.RecomputeForPermissionSet(ctx, event.EntityID)
	case "role_changed":
		if w.rlsComputer != nil {
			if err := w.rlsComputer.RecomputeRoleHierarchy(ctx); err != nil {
				return err
			}
			return w.rlsComputer.RecomputeVisibleOwnersAll(ctx)
		}
		return nil
	case "group_changed":
		if w.rlsComputer != nil {
			return w.rlsComputer.RecomputeGroupMembersForGroup(ctx, event.EntityID)
		}
		return nil
	case "object_changed":
		if w.rlsComputer != nil {
			return w.rlsComputer.RecomputeObjectHierarchy(ctx)
		}
		return nil
	case "territory_changed":
		// Territory cache recomputation is handled by enterprise edition.
		// In community edition, this is a no-op since there are no territories.
		w.logger.Info("outbox worker: territory_changed event (enterprise feature)",
			"entity_id", event.EntityID,
		)
		return nil
	default:
		w.logger.Warn("outbox worker: unknown event type",
			"event_type", event.EventType,
			"event_id", event.ID,
		)
		return nil
	}
}

// ParseConnConfig parses DSN into pgx.ConnConfig for the dedicated LISTEN connection.
func ParseConnConfig(dsn string) (*pgx.ConnConfig, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parseConnConfig: %w", err)
	}
	return config, nil
}

// WellKnownUserID returns the system admin user UUID for dev auth default.
func WellKnownUserID() uuid.UUID {
	return SystemAdminUserID
}
