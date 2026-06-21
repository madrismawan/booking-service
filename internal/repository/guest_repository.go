package repository

import (
	"booking-service/internal/model"

	"gorm.io/gorm"
)

type GuestRepository struct {
	db *gorm.DB
}

func NewGuestRepository(db *gorm.DB) *GuestRepository {
	return &GuestRepository{db: db}
}

func (r *GuestRepository) WithTx(tx *gorm.DB) *GuestRepository {
	return &GuestRepository{db: tx}
}

func (r *GuestRepository) Create(guest *model.Guest) error {
	return r.db.Create(guest).Error
}
