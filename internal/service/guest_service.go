package service

import (
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type GuestService struct {
	repo *repository.GuestRepository
}

func NewGuestService(repo *repository.GuestRepository) *GuestService {
	return &GuestService{repo: repo}
}

func (s *GuestService) WithTx(tx *gorm.DB) *GuestService {
	return &GuestService{repo: s.repo.WithTx(tx)}
}

func (s *GuestService) CreateGuest(guest *model.Guest) error {
	return s.repo.Create(guest)
}
