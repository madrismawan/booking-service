package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) WithTx(tx *gorm.DB) *EventRepository {
	return &EventRepository{db: tx}
}

func (r *EventRepository) FindBySlug(slug string) (*model.Event, error) {
	var event model.Event
	err := r.db.Where("slug = ? AND deleted_at IS NULL", slug).First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &event, err
}
