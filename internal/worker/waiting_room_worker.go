package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"booking-service/internal/rabbitmq"
	"booking-service/internal/service"
)

const checkoutTokenTTL = 5 * time.Minute

type WaitingRoomWorker struct {
	mq                 *rabbitmq.Client
	waitingRoomService *service.WaitingRoomService
	logger             *log.Logger
}

func NewWaitingRoomWorker(
	mq *rabbitmq.Client,
	waitingRoomService *service.WaitingRoomService,
	logger *log.Logger,
) *WaitingRoomWorker {
	if logger == nil {
		logger = log.Default()
	}
	return &WaitingRoomWorker{
		mq:                 mq,
		waitingRoomService: waitingRoomService,
		logger:             logger,
	}
}

func (w *WaitingRoomWorker) Start(ctx context.Context) error {
	deliveries, err := w.mq.Consume(rabbitmq.WaitingRoomQueue, "booking-service-waiting-room-worker")
	if err != nil {
		return err
	}

	w.logger.Printf("waiting room worker listening on queue %q", rabbitmq.WaitingRoomQueue)

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("waiting room delivery channel closed")
			}
			action := w.handleDelivery(delivery.Body)
			switch action {
			case deliveryAck:
				_ = delivery.Ack(false)
			case deliveryReject:
				_ = delivery.Nack(false, false)
			case deliveryRequeue:
				_ = delivery.Nack(false, true)
			}
		}
	}
}

type deliveryAction string

const (
	deliveryAck     deliveryAction = "ack"
	deliveryReject  deliveryAction = "reject"
	deliveryRequeue deliveryAction = "requeue"
)

func (w *WaitingRoomWorker) handleDelivery(body []byte) deliveryAction {
	var message rabbitmq.WaitingRoomMessage
	if err := json.Unmarshal(body, &message); err != nil {
		w.logger.Printf("invalid waiting room message: %v", err)
		return deliveryReject
	}

	waitingRoom, processed, err := w.waitingRoomService.MarkReady(message.QueueToken, checkoutTokenTTL)
	if err != nil {
		w.logger.Printf("mark waiting room ready queue_token=%s: %v", message.QueueToken, err)
		return deliveryRequeue
	}
	if !processed {
		w.logger.Printf(
			"skip waiting room queue_token=%s status=%s",
			waitingRoom.QueueToken,
			waitingRoom.Status,
		)
		return deliveryAck
	}

	if waitingRoom.Status == "failed" {
		failedReason := ""
		if waitingRoom.FailedReason != nil {
			failedReason = *waitingRoom.FailedReason
		}
		w.logger.Printf(
			"waiting room failed: queue_token=%s ticket_category_id=%d reason=%s",
			waitingRoom.QueueToken,
			waitingRoom.TicketCategoryID,
			failedReason,
		)
		return deliveryAck
	}

	w.logger.Printf(
		"waiting room ready: queue_token=%s ticket_category_id=%d expires_at=%s",
		waitingRoom.QueueToken,
		waitingRoom.TicketCategoryID,
		waitingRoom.ExpiredAt.Format(time.RFC3339),
	)
	return deliveryAck
}
