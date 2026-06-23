package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TicketStockRepository struct {
	db *gorm.DB
}

func NewTicketStockRepository(db *gorm.DB) *TicketStockRepository {
	return &TicketStockRepository{db: db}
}

func (r *TicketStockRepository) WithTx(tx *gorm.DB) *TicketStockRepository {
	return &TicketStockRepository{db: tx}
}

func (r *TicketStockRepository) FindByTicketCategoryID(ticketCategoryID int64) (*model.TicketStock, error) {
	var stock model.TicketStock
	err := r.db.Where("ticket_category_id = ?", ticketCategoryID).First(&stock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &stock, err
}

func (r *TicketStockRepository) FindByTicketCategoryIDForUpdate(ticketCategoryID int64) (*model.TicketStock, error) {
	var stock model.TicketStock
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("ticket_category_id = ?", ticketCategoryID).
		First(&stock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &stock, err
}

func (r *TicketStockRepository) Save(stock *model.TicketStock) error {
	return r.db.Save(stock).Error
}
