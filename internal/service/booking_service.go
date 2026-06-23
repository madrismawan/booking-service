package service

import (
	"errors"
	"fmt"
	"time"

	"booking-service/internal/dto"
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type BookingService struct {
	txManager             *repository.TransactionManager
	repo                  *repository.BookingRepository
	ticketCategoryService *TicketCategoryService
	ticketStockService    *TicketStockService
	guestService          *GuestService
	bookingItemService    *BookingItemService
	waitingRoomRepo       *repository.WaitingRoomRepository
}

func NewBookingService(
	txManager *repository.TransactionManager,
	repo *repository.BookingRepository,
	ticketCategoryService *TicketCategoryService,
	ticketStockService *TicketStockService,
	guestService *GuestService,
	bookingItemService *BookingItemService,
	waitingRoomRepo *repository.WaitingRoomRepository,
) *BookingService {
	return &BookingService{
		txManager:             txManager,
		repo:                  repo,
		ticketCategoryService: ticketCategoryService,
		ticketStockService:    ticketStockService,
		guestService:          guestService,
		bookingItemService:    bookingItemService,
		waitingRoomRepo:       waitingRoomRepo,
	}
}

func (s *BookingService) CreateBooking(req dto.CreateBookingRequest) (*model.Booking, error) {
	var booking model.Booking
	var bookingItem model.BookingItem

	err := s.txManager.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		ticketCategoryService := s.ticketCategoryService.WithTx(tx)
		ticketStockService := s.ticketStockService.WithTx(tx)
		guestService := s.guestService.WithTx(tx)
		bookingItemService := s.bookingItemService.WithTx(tx)
		waitingRoomRepo := s.waitingRoomRepo.WithTx(tx)

		category, err := ticketCategoryService.FindByID(req.TicketCategoryID)
		if err != nil {
			return err
		}
		event := category.Event

		var waitingRoom *model.WaitingRoom
		waitingRoom, err = waitingRoomRepo.FindByCheckoutTokenForUpdate(req.CheckoutToken)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return repository.ErrInvalidCheckout
			}
			return err
		}
		if waitingRoom.TicketCategoryID != category.ID || waitingRoom.Status != model.WaitingRoomStatusReady {
			return repository.ErrInvalidCheckout
		}
		if waitingRoom.ExpiredAt == nil || time.Now().After(*waitingRoom.ExpiredAt) {
			waitingRoom.Status = model.WaitingRoomStatusExpired
			_ = waitingRoomRepo.Save(waitingRoom)
			return repository.ErrExpiredCheckout
		}
		waitingRoom.Status = model.WaitingRoomStatusCheckoutStarted
		if err := waitingRoomRepo.Save(waitingRoom); err != nil {
			return err
		}

		if _, err := ticketStockService.ReserveForUpdate(category.ID, req.Quantity); err != nil {
			return err
		}

		guest := model.Guest{
			Name:    "Guest",
			Email:   fmt.Sprintf("guest-%d@example.com", time.Now().UnixNano()),
			Phone:   "-",
			Address: "-",
		}
		if err := guestService.CreateGuest(&guest); err != nil {
			return err
		}

		subtotal := category.Price * int64(req.Quantity)
		booking = model.Booking{
			BookingCode:       fmt.Sprintf("BKG-%d", time.Now().UnixNano()),
			GuestID:           guest.ID,
			GuestName:         guest.Name,
			GuestEmail:        guest.Email,
			GuestPhone:        guest.Phone,
			GuestAddress:      guest.Address,
			EventID:           event.ID,
			EventSlug:         event.Slug,
			EventName:         event.Name,
			EventVenueName:    event.VenueName,
			EventVenueAddress: event.VenueAddress,
			EventStartsAt:     event.StartsAt,
			EventEndsAt:       event.EndsAt,
			Status:            "pending_payment",
			TotalTicket:       req.Quantity,
			TotalPrice:        subtotal,
			ExpiresAt:         time.Now().Add(15 * time.Minute),
		}
		if err := repo.Create(&booking); err != nil {
			return err
		}

		bookingItem = model.BookingItem{
			BookingID:                 booking.ID,
			TicketCategoryID:          category.ID,
			TicketCategoryName:        category.Name,
			TicketCategoryDescription: category.Description,
			Quantity:                  req.Quantity,
			UnitPrice:                 category.Price,
			SubtotalPrice:             subtotal,
		}
		if err := bookingItemService.Create(&bookingItem); err != nil {
			return err
		}
		booking.Items = []model.BookingItem{bookingItem}

		waitingRoom.BookingID = &booking.ID
		waitingRoom.Status = model.WaitingRoomStatusCompleted
		if err := waitingRoomRepo.Save(waitingRoom); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &booking, nil
}
