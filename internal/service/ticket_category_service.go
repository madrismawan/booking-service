package service

import (
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type TicketCategoryService struct {
	repo *repository.TicketCategoryRepository
}

func NewTicketCategoryService(repo *repository.TicketCategoryRepository) *TicketCategoryService {
	return &TicketCategoryService{repo: repo}
}

func (s *TicketCategoryService) WithTx(tx *gorm.DB) *TicketCategoryService {
	return &TicketCategoryService{repo: s.repo.WithTx(tx)}
}

func (s *TicketCategoryService) FindFirstByEventID(eventID int64) (*model.TicketCategory, error) {
	return s.repo.FindFirstByEventID(eventID)
}
