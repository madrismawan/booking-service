package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/repository"
)

const (
	defaultOutboxPollInterval      = time.Second
	defaultOutboxProcessingTimeout = time.Minute
	defaultOutboxPublishTimeout    = 5 * time.Second
	maxOutboxRetryDelay            = 5 * time.Minute
)

type outboxEventStore interface {
	ClaimNext(context.Context, time.Duration) (*model.OutboxEvent, error)
	MarkSent(context.Context, int64, time.Time) error
	MarkRetry(context.Context, int64, string, time.Time) error
}

type outboxPublisher interface {
	PublishJSON(context.Context, string, any) error
}

type OutboxWorker struct {
	store             outboxEventStore
	publisher         outboxPublisher
	logger            *log.Logger
	pollInterval      time.Duration
	processingTimeout time.Duration
	publishTimeout    time.Duration
}

func NewOutboxWorker(
	store outboxEventStore,
	publisher outboxPublisher,
	logger *log.Logger,
) *OutboxWorker {
	if logger == nil {
		logger = log.Default()
	}
	return &OutboxWorker{
		store:             store,
		publisher:         publisher,
		logger:            logger,
		pollInterval:      defaultOutboxPollInterval,
		processingTimeout: defaultOutboxProcessingTimeout,
		publishTimeout:    defaultOutboxPublishTimeout,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) error {
	w.logger.Printf("outbox worker publishing to queue %q", rabbitmq.TicketStockChangedQueue)

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			processed, err := w.ProcessNext(ctx)
			if err != nil {
				w.logger.Printf("process outbox event: %v", err)
			}

			delay := w.pollInterval
			if processed {
				delay = 0
			}
			timer.Reset(delay)
		}
	}
}

func (w *OutboxWorker) ProcessNext(ctx context.Context) (bool, error) {
	event, err := w.store.ClaimNext(ctx, w.processingTimeout)
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := w.publish(ctx, event); err != nil {
		nextAttemptAt := time.Now().Add(outboxRetryDelay(event.Attempts))
		if markErr := w.store.MarkRetry(ctx, event.ID, err.Error(), nextAttemptAt); markErr != nil {
			return true, fmt.Errorf("publish: %v; mark retry: %w", err, markErr)
		}
		return true, err
	}

	if err := w.store.MarkSent(ctx, event.ID, time.Now()); err != nil {
		return true, fmt.Errorf("mark outbox event sent: %w", err)
	}

	return true, nil
}

func (w *OutboxWorker) publish(ctx context.Context, event *model.OutboxEvent) error {
	switch event.EventType {
	case rabbitmq.TicketStockChangedEventType:
		var payload rabbitmq.TicketStockChangedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode ticket stock event payload: %w", err)
		}
		if payload.EventType != rabbitmq.TicketStockChangedEventType {
			return fmt.Errorf("invalid ticket stock event type %q", payload.EventType)
		}

		message := rabbitmq.TicketStockChangedMessage{
			EventID:          event.ID,
			EventType:        payload.EventType,
			SchemaVersion:    payload.SchemaVersion,
			TicketCategoryID: payload.TicketCategoryID,
			StockVersion:     payload.StockVersion,
			SnapshotURL:      payload.SnapshotURL,
			ChangedAt:        payload.ChangedAt,
		}

		publishCtx, cancel := context.WithTimeout(ctx, w.publishTimeout)
		defer cancel()
		return w.publisher.PublishJSON(publishCtx, rabbitmq.TicketStockChangedQueue, message)
	default:
		return fmt.Errorf("unsupported outbox event type %q", event.EventType)
	}
}

func outboxRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	delay := time.Duration(attempt*attempt) * 5 * time.Second
	if delay > maxOutboxRetryDelay {
		return maxOutboxRetryDelay
	}
	return delay
}
