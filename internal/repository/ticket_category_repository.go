package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

type TicketCategoryRepository struct {
	db *gorm.DB
}

func NewTicketCategoryRepository(db *gorm.DB) *TicketCategoryRepository {
	return &TicketCategoryRepository{db: db}
}

func (r *TicketCategoryRepository) WithTx(tx *gorm.DB) *TicketCategoryRepository {
	return &TicketCategoryRepository{db: tx}
}

func (r *TicketCategoryRepository) FindByID(id int64) (*model.TicketCategory, error) {
	var category model.TicketCategory
	err := r.db.Preload("Event").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &category, err
}

func (r *TicketCategoryRepository) FindFirstByEventID(eventID int64) (*model.TicketCategory, error) {
	var category model.TicketCategory
	err := r.db.Where("event_id = ? AND deleted_at IS NULL", eventID).Order("id ASC").First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &category, err
}
