package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"booking-service/internal/config"
	"booking-service/internal/rabbitmq"
)

const maxAccountingAttempts = 3

type AccountingWorker struct {
	consumer   *rabbitmq.Client
	publisher  *rabbitmq.Client
	httpClient *http.Client
	apiURL     string
	apiToken   string
	logger     *log.Logger
}

func NewAccountingWorker(
	consumer *rabbitmq.Client,
	publisher *rabbitmq.Client,
	cfg config.AccountingConfig,
	logger *log.Logger,
) *AccountingWorker {
	if logger == nil {
		logger = log.Default()
	}
	return &AccountingWorker{
		consumer:  consumer,
		publisher: publisher,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		apiURL:   cfg.APIURL,
		apiToken: cfg.APIToken,
		logger:   logger,
	}
}

func (w *AccountingWorker) Start(ctx context.Context) error {
	if err := w.declareQueues(); err != nil {
		return err
	}

	deliveries, err := w.consumer.Consume(
		rabbitmq.AccountingPaymentSucceededQueue,
		"booking-service-accounting-worker",
	)
	if err != nil {
		return err
	}

	w.logger.Printf(
		"accounting worker listening on queue %q",
		rabbitmq.AccountingPaymentSucceededQueue,
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("accounting delivery channel closed")
			}
			if err := w.handleDelivery(ctx, delivery.Body); err != nil {
				w.logger.Printf("handle accounting delivery: %v", err)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}

func (w *AccountingWorker) declareQueues() error {
	if _, err := w.consumer.DeclareQueue(rabbitmq.AccountingPaymentSucceededQueue); err != nil {
		return err
	}
	if _, err := w.consumer.DeclareQueue(rabbitmq.AccountingPaymentSucceededDLQ); err != nil {
		return err
	}
	if _, err := w.consumer.DeclareRetryQueue(
		rabbitmq.AccountingPaymentRetry5sQueue,
		rabbitmq.AccountingPaymentSucceededQueue,
		5*time.Second,
	); err != nil {
		return err
	}
	if _, err := w.consumer.DeclareRetryQueue(
		rabbitmq.AccountingPaymentRetry10sQueue,
		rabbitmq.AccountingPaymentSucceededQueue,
		10*time.Second,
	); err != nil {
		return err
	}
	return nil
}

func (w *AccountingWorker) handleDelivery(ctx context.Context, body []byte) error {
	var message rabbitmq.AccountingPaymentSucceededMessage
	if err := json.Unmarshal(body, &message); err != nil {
		return w.publisher.PublishBytesDeclared(
			ctx,
			rabbitmq.AccountingPaymentSucceededDLQ,
			"application/json",
			body,
		)
	}
	if message.Attempt < 1 {
		message.Attempt = 1
	}

	retryable, err := w.sendToAccounting(ctx, message)
	if err == nil {
		w.logger.Printf(
			"accounting payment delivered event_id=%d booking_id=%d",
			message.EventID,
			message.BookingID,
		)
		return nil
	}

	message.LastError = err.Error()
	if !retryable || message.Attempt >= maxAccountingAttempts {
		w.logger.Printf(
			"accounting payment moved to DLQ event_id=%d attempt=%d error=%v",
			message.EventID,
			message.Attempt,
			err,
		)
		return w.publisher.PublishJSONDeclared(
			ctx,
			rabbitmq.AccountingPaymentSucceededDLQ,
			message,
		)
	}

	message.Attempt++
	retryQueue := rabbitmq.AccountingPaymentRetry5sQueue
	if message.Attempt == maxAccountingAttempts {
		retryQueue = rabbitmq.AccountingPaymentRetry10sQueue
	}

	w.logger.Printf(
		"accounting payment scheduled for retry event_id=%d attempt=%d error=%v",
		message.EventID,
		message.Attempt,
		err,
	)
	return w.publisher.PublishJSONDeclared(ctx, retryQueue, message)
}

func (w *AccountingWorker) sendToAccounting(
	ctx context.Context,
	message rabbitmq.AccountingPaymentSucceededMessage,
) (bool, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		w.apiURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "accounting-payment-"+strconv.FormatInt(message.EventID, 10))
	if w.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiToken)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return true, fmt.Errorf("call accounting API: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return false, nil
	}

	responseErr := fmt.Errorf("accounting API returned status %d", resp.StatusCode)
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		return true, responseErr
	}
	return false, responseErr
}
