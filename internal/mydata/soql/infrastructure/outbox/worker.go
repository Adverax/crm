// Package outbox provides a worker for processing metadata change events from the outbox table.
package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proxima-research/proxima.crm.kernel/database/sql"
	"github.com/proxima-research/proxima.crm.kernel/log"
	soqlModel "github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
)

const (
	outboxChannelName = "metadata_outbox_events"
)

// EventResult represents the result of processing an outbox event.
type EventResult struct {
	EventID       int64
	EventType     string
	EntityType    string
	EntityID      int64
	ObjectApiName string
}

// Worker processes metadata change events from the outbox table
// and invalidates corresponding SOQL caches.
type Worker struct {
	db          sql.DB
	logger      log.Logger
	invalidator soqlModel.CacheInvalidator
	stopCh      chan struct{}
	doneCh      chan struct{}
	eventCh     chan EventResult
	cancelFunc  context.CancelFunc
}

// NewWorker creates a new metadata outbox worker.
func NewWorker(db sql.DB, logger log.Logger, invalidator soqlModel.CacheInvalidator) *Worker {
	return &Worker{
		db:          db,
		logger:      logger,
		invalidator: invalidator,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

// NewWorkerWithEventChannel creates a new worker with an event channel for testing.
func NewWorkerWithEventChannel(db sql.DB, logger log.Logger, invalidator soqlModel.CacheInvalidator, eventCh chan EventResult) *Worker {
	worker := NewWorker(db, logger, invalidator)
	worker.eventCh = eventCh
	return worker
}

// Start begins listening for metadata change events.
func (w *Worker) Start(ctx context.Context) error {
	pool := w.db.Pool()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}

	// Start LISTEN on the channel
	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s", outboxChannelName))
	if err != nil {
		conn.Release()
		return fmt.Errorf("listen %s failed: %w", outboxChannelName, err)
	}

	ctx, w.cancelFunc = context.WithCancel(ctx)
	go w.run(ctx, conn)
	return nil
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	close(w.stopCh)
	w.cancelFunc()
	<-w.doneCh
}

func (w *Worker) run(ctx context.Context, conn *pgxpool.Conn) {
	defer close(w.doneCh)
	defer conn.Release()
	defer func() {
		if w.eventCh != nil {
			close(w.eventCh)
		}
	}()

	// Process any pending events on startup
	w.drainQueue(ctx)

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.WithError(err).Error(ctx, "error waiting for notification")
				time.Sleep(1 * time.Second)
				continue
			}

			if notification != nil {
				w.drainQueue(ctx)
			}
		}
	}
}

func (w *Worker) drainQueue(ctx context.Context) {
	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			result, err := w.processNextEvent(ctx)
			if err != nil {
				w.logger.WithError(err).Error(ctx, "failed to process metadata outbox event")
				return
			}

			if result != nil && w.eventCh != nil {
				select {
				case w.eventCh <- *result:
				default:
				}
			}

			if result == nil {
				return
			}

			// Invalidate caches based on the event
			w.invalidateCache(ctx, result)
		}
	}
}

func (w *Worker) processNextEvent(ctx context.Context) (*EventResult, error) {
	query := `SELECT * FROM metadata.process_metadata_outbox_event()`

	var (
		eventID       *int64
		eventType     *string
		entityType    *string
		entityID      *int64
		objectApiName *string
	)

	err := w.db.QueryRow(ctx, query).Scan(
		&eventID,
		&eventType,
		&entityType,
		&entityID,
		&objectApiName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// All fields are nullable in the result type
	if eventID == nil || eventType == nil || entityType == nil || entityID == nil {
		return nil, nil
	}

	result := &EventResult{
		EventID:    *eventID,
		EventType:  *eventType,
		EntityType: *entityType,
		EntityID:   *entityID,
	}

	if objectApiName != nil {
		result.ObjectApiName = *objectApiName
	}

	return result, nil
}

func (w *Worker) invalidateCache(ctx context.Context, event *EventResult) {
	if w.invalidator == nil {
		return
	}

	if event.ObjectApiName == "" {
		// No object name - clear all caches as fallback
		w.invalidator.InvalidateAll(ctx)
		w.invalidator.ClearQueryCache(ctx)
		return
	}

	// Invalidate the object metadata cache
	w.invalidator.InvalidateObject(ctx, event.ObjectApiName)

	// Invalidate only queries that depend on this object (targeted invalidation)
	w.invalidator.InvalidateQueriesByObject(ctx, event.ObjectApiName)

	w.logger.Info(ctx, "invalidated caches for object", log.Fields{
		"event_id":        event.EventID,
		"event_type":      event.EventType,
		"entity_type":     event.EntityType,
		"entity_id":       event.EntityID,
		"object_api_name": event.ObjectApiName,
	})
}
