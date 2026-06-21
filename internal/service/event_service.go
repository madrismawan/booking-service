package service

import (
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) WithTx(tx *gorm.DB) *EventService {
	return &EventService{repo: s.repo.WithTx(tx)}
}

func (s *EventService) FindBySlug(slug string) (*model.Event, error) {
	return s.repo.FindBySlug(slug)
}
