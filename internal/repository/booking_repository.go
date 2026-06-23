package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *BookingRepository) FindByIDForUpdate(id int64) (*model.Booking, error) {
	var booking model.Booking
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&booking).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &booking, err
}

func (r *BookingRepository) Save(booking *model.Booking) error {
	return r.db.Save(booking).Error
}
