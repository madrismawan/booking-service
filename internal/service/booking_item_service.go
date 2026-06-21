package service

import (
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type BookingItemService struct {
	repo *repository.BookingItemRepository
}

func NewBookingItemService(repo *repository.BookingItemRepository) *BookingItemService {
	return &BookingItemService{repo: repo}
}

func (s *BookingItemService) WithTx(tx *gorm.DB) *BookingItemService {
	return &BookingItemService{repo: s.repo.WithTx(tx)}
}

func (s *BookingItemService) Create(item *model.BookingItem) error {
	return s.repo.Create(item)
}
