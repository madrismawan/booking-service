package service

import (
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type TicketStockService struct {
	repo *repository.TicketStockRepository
}

func NewTicketStockService(repo *repository.TicketStockRepository) *TicketStockService {
	return &TicketStockService{repo: repo}
}

func (s *TicketStockService) WithTx(tx *gorm.DB) *TicketStockService {
	return &TicketStockService{repo: s.repo.WithTx(tx)}
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
	if err := s.repo.Save(stock); err != nil {
		return nil, err
	}

	return stock, nil
}
