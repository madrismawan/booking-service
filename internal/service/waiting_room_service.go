package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type WaitingRoomService struct {
	txManager             *repository.TransactionManager
	repo                  *repository.WaitingRoomRepository
	outboxRepo            *repository.OutboxEventRepository
	ticketCategoryService *TicketCategoryService
	ticketStockService    *TicketStockService
}

func NewWaitingRoomService(
	txManager *repository.TransactionManager,
	repo *repository.WaitingRoomRepository,
	outboxRepo *repository.OutboxEventRepository,
	ticketCategoryService *TicketCategoryService,
	ticketStockService *TicketStockService,
) *WaitingRoomService {
	return &WaitingRoomService{
		txManager:             txManager,
		repo:                  repo,
		outboxRepo:            outboxRepo,
		ticketCategoryService: ticketCategoryService,
		ticketStockService:    ticketStockService,
	}
}

func (s *WaitingRoomService) WithTx(tx *gorm.DB) *WaitingRoomService {
	return &WaitingRoomService{
		txManager:             s.txManager,
		repo:                  s.repo.WithTx(tx),
		outboxRepo:            s.outboxRepo.WithTx(tx),
		ticketCategoryService: s.ticketCategoryService.WithTx(tx),
		ticketStockService:    s.ticketStockService.WithTx(tx),
	}
}

func (s *WaitingRoomService) JoinQueue(ticketCategoryID int64) (*model.WaitingRoom, error) {
	queueToken, err := generateToken("queue")
	if err != nil {
		return nil, err
	}

	var waitingRoom model.WaitingRoom
	err = s.txManager.Transaction(func(tx *gorm.DB) error {
		service := s.WithTx(tx)

		category, err := service.ticketCategoryService.FindByID(ticketCategoryID)
		if err != nil {
			return err
		}

		waitingRoom = model.WaitingRoom{
			EventID:          category.EventID,
			EventName:        category.Event.Name,
			TicketCategoryID: category.ID,
			QueueToken:       queueToken,
			Status:           model.WaitingRoomStatusWaiting,
		}
		if err := service.repo.Create(&waitingRoom); err != nil {
			return err
		}

		payload, err := json.Marshal(rabbitmq.WaitingRoomJoinedPayload{
			TicketCategoryID: waitingRoom.TicketCategoryID,
			QueueToken:       waitingRoom.QueueToken,
			CreatedAt:        waitingRoom.CreatedAt,
		})
		if err != nil {
			return err
		}

		return service.outboxRepo.Create(&model.OutboxEvent{
			AggregateType: "waiting_room",
			AggregateID:   waitingRoom.ID,
			EventType:     rabbitmq.WaitingRoomJoinedEventType,
			Payload:       payload,
			Status:        model.OutboxStatusPending,
			NextAttemptAt: time.Now(),
		})
	})
	if err != nil {
		return nil, err
	}

	return &waitingRoom, nil
}

func (s *WaitingRoomService) GetStatus(queueToken string) (*model.WaitingRoom, error) {
	waitingRoom, err := s.repo.FindByQueueToken(queueToken)
	if err != nil {
		return nil, err
	}

	if waitingRoom.Status == model.WaitingRoomStatusReady &&
		waitingRoom.ExpiredAt != nil &&
		time.Now().After(*waitingRoom.ExpiredAt) {
		waitingRoom.Status = model.WaitingRoomStatusExpired
		if err := s.repo.Save(waitingRoom); err != nil {
			return nil, err
		}
	}

	return waitingRoom, nil
}

func (s *WaitingRoomService) MarkReady(queueToken string, ttl time.Duration) (*model.WaitingRoom, bool, error) {
	var waitingRoom model.WaitingRoom
	processed := false

	err := s.txManager.Transaction(func(tx *gorm.DB) error {
		service := s.WithTx(tx)

		record, err := service.repo.FindByQueueToken(queueToken)
		if err != nil {
			return err
		}

		if record.Status != model.WaitingRoomStatusWaiting {
			waitingRoom = *record
			return nil
		}

		stock, err := service.ticketStockService.FindByTicketCategoryID(record.TicketCategoryID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				reason := "ticket stock not found"
				record.Status = model.WaitingRoomStatusFailed
				record.FailedReason = &reason
				if err := service.repo.Save(record); err != nil {
					return err
				}
				waitingRoom = *record
				processed = true
				return nil
			}
			return err
		}

		if stock.AvailableQuantity <= 0 {
			reason := "ticket stock is sold out"
			record.Status = model.WaitingRoomStatusFailed
			record.FailedReason = &reason
			if err := service.repo.Save(record); err != nil {
				return err
			}
			waitingRoom = *record
			processed = true
			return nil
		}

		activeCheckoutCount, err := service.repo.CountActiveCheckoutByTicketCategoryID(record.TicketCategoryID)
		if err != nil {
			return err
		}
		if activeCheckoutCount >= int64(stock.AvailableQuantity) {
			reason := "checkout queue exceeds available ticket stock"
			record.Status = model.WaitingRoomStatusFailed
			record.FailedReason = &reason
			if err := service.repo.Save(record); err != nil {
				return err
			}
			waitingRoom = *record
			processed = true
			return nil
		}

		checkoutToken, err := generateToken("checkout")
		if err != nil {
			return err
		}
		expiredAt := time.Now().Add(ttl)

		record.CheckoutToken = &checkoutToken
		record.ExpiredAt = &expiredAt
		record.Status = model.WaitingRoomStatusReady

		if err := service.repo.Save(record); err != nil {
			return err
		}

		waitingRoom = *record
		processed = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}

	return &waitingRoom, processed, nil
}

func generateToken(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	if prefix == "" {
		return token, nil
	}
	return prefix + "_" + token, nil
}
