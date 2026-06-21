package repository

import (
	"booking-service/internal/model"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) WithTx(tx *gorm.DB) *BookingRepository {
	return &BookingRepository{db: tx}
}

func (r *BookingRepository) Create(booking *model.Booking) error {
	return r.db.Create(booking).Error
}
