package repository

import (
	"booking-service/internal/model"

	"gorm.io/gorm"
)

type BookingItemRepository struct {
	db *gorm.DB
}

func NewBookingItemRepository(db *gorm.DB) *BookingItemRepository {
	return &BookingItemRepository{db: db}
}

func (r *BookingItemRepository) WithTx(tx *gorm.DB) *BookingItemRepository {
	return &BookingItemRepository{db: tx}
}

func (r *BookingItemRepository) Create(item *model.BookingItem) error {
	return r.db.Create(item).Error
}

func (r *BookingItemRepository) FindByBookingID(bookingID int64) ([]model.BookingItem, error) {
	var items []model.BookingItem
	err := r.db.Where("booking_id = ?", bookingID).Order("id ASC").Find(&items).Error
	return items, err
}
