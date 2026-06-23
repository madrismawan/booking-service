package service

import (
	"encoding/json"
	"fmt"
	"time"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type TicketStockService struct {
	repo       *repository.TicketStockRepository
	outboxRepo *repository.OutboxEventRepository
}

func NewTicketStockService(
	repo *repository.TicketStockRepository,
	outboxRepo *repository.OutboxEventRepository,
) *TicketStockService {
	return &TicketStockService{repo: repo, outboxRepo: outboxRepo}
}

func (s *TicketStockService) WithTx(tx *gorm.DB) *TicketStockService {
	return &TicketStockService{
		repo:       s.repo.WithTx(tx),
		outboxRepo: s.outboxRepo.WithTx(tx),
	}
}

func (s *TicketStockService) FindByTicketCategoryID(ticketCategoryID int64) (*model.TicketStock, error) {
	return s.repo.FindByTicketCategoryID(ticketCategoryID)
}

func (s *TicketStockService) FindByTicketCategoryIDForUpdate(ticketCategoryID int64) (*model.TicketStock, error) {
	return s.repo.FindByTicketCategoryIDForUpdate(ticketCategoryID)
}

func (s *TicketStockService) ReserveForUpdate(ticketCategoryID int64, quantity int) (*model.TicketStock, error) {
	stock, err := s.repo.FindByTicketCategoryIDForUpdate(ticketCategoryID)
	if err != nil {
		return nil, err
	}
	if stock.AvailableQuantity < quantity {
		return nil, repository.ErrInsufficientStock
	}

	stock.AvailableQuantity -= quantity
	stock.ReservedQuantity += quantity
	stock.Version++
	stock.UpdatedAt = time.Now()
	if err := s.repo.Save(stock); err != nil {
		return nil, err
	}

	if err := s.createStockChangedEvent(stock); err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *TicketStockService) MarkSoldForUpdate(
	ticketCategoryID int64,
	quantity int,
) (*model.TicketStock, error) {
	stock, err := s.repo.FindByTicketCategoryIDForUpdate(ticketCategoryID)
	if err != nil {
		return nil, err
	}
	if stock.ReservedQuantity < quantity {
		return nil, repository.ErrPaymentConflict
	}

	stock.ReservedQuantity -= quantity
	stock.SoldQuantity += quantity
	stock.Version++
	stock.UpdatedAt = time.Now()
	if err := s.repo.Save(stock); err != nil {
		return nil, err
	}

	if err := s.createStockChangedEvent(stock); err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *TicketStockService) createStockChangedEvent(stock *model.TicketStock) error {
	payload, err := json.Marshal(rabbitmq.TicketStockChangedPayload{
		EventType:        rabbitmq.TicketStockChangedEventType,
		SchemaVersion:    rabbitmq.TicketStockChangedSchemaVersion,
		TicketCategoryID: stock.TicketCategoryID,
		StockVersion:     stock.Version,
		SnapshotURL:      fmt.Sprintf("/api/v1/ticket-categories/%d/stock", stock.TicketCategoryID),
		ChangedAt:        stock.UpdatedAt,
	})
	if err != nil {
		return err
	}

	now := time.Now()
	if err := s.outboxRepo.Create(&model.OutboxEvent{
		AggregateType: "ticket_stock",
		AggregateID:   stock.ID,
		EventType:     rabbitmq.TicketStockChangedEventType,
		Payload:       payload,
		Status:        model.OutboxStatusPending,
		NextAttemptAt: now,
	}); err != nil {
		return err
	}
	return nil
}
