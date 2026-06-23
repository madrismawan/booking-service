package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	ticketCategoryService *TicketCategoryService
	ticketStockService    *TicketStockService
	publisher             waitingRoomPublisher
}

type waitingRoomPublisher interface {
	PublishJSON(ctx context.Context, queueName string, payload any) error
}

func NewWaitingRoomService(
	txManager *repository.TransactionManager,
	repo *repository.WaitingRoomRepository,
	ticketCategoryService *TicketCategoryService,
	ticketStockService *TicketStockService,
	publisher waitingRoomPublisher,
) *WaitingRoomService {
	return &WaitingRoomService{
		txManager:             txManager,
		repo:                  repo,
		ticketCategoryService: ticketCategoryService,
		ticketStockService:    ticketStockService,
		publisher:             publisher,
	}
}

func (s *WaitingRoomService) WithTx(tx *gorm.DB) *WaitingRoomService {
	return &WaitingRoomService{
		txManager:             s.txManager,
		repo:                  s.repo.WithTx(tx),
		ticketCategoryService: s.ticketCategoryService.WithTx(tx),
		ticketStockService:    s.ticketStockService.WithTx(tx),
		publisher:             s.publisher,
	}
}

func (s *WaitingRoomService) JoinQueue(ticketCategoryID int64) (*model.WaitingRoom, error) {
	category, err := s.ticketCategoryService.FindByID(ticketCategoryID)
	if err != nil {
		return nil, err
	}

	queueToken, err := generateToken("queue")
	if err != nil {
		return nil, err
	}

	waitingRoom := model.WaitingRoom{
		EventID:          category.EventID,
		EventName:        category.Event.Name,
		TicketCategoryID: category.ID,
		QueueToken:       queueToken,
		Status:           model.WaitingRoomStatusWaiting,
	}
	if err := s.repo.Create(&waitingRoom); err != nil {
		return nil, err
	}

	if s.publisher == nil {
		waitingRoom.Status = model.WaitingRoomStatusFailed
		_ = s.repo.Save(&waitingRoom)
		return &waitingRoom, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.publisher.PublishJSON(ctx, rabbitmq.WaitingRoomQueue, rabbitmq.WaitingRoomMessage{
		TicketCategoryID: waitingRoom.TicketCategoryID,
		QueueToken:       waitingRoom.QueueToken,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		waitingRoom.Status = model.WaitingRoomStatusFailed
		_ = s.repo.Save(&waitingRoom)
		return &waitingRoom, nil
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
