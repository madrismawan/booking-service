package app

import (
	"context"

	"booking-service/internal/repository"
	"booking-service/internal/service"

	"gorm.io/gorm"
)

type Publisher interface {
	PublishJSON(ctx context.Context, queueName string, payload any) error
}

type Container struct {
	BookingService     *service.BookingService
	WaitingRoomService *service.WaitingRoomService
	TicketStockService *service.TicketStockService
	OutboxEventRepo    *repository.OutboxEventRepository
}

func NewContainer(db *gorm.DB, publisher Publisher) *Container {
	txManager := repository.NewTransactionManager(db)

	ticketCategoryRepo := repository.NewTicketCategoryRepository(db)
	ticketStockRepo := repository.NewTicketStockRepository(db)
	outboxEventRepo := repository.NewOutboxEventRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingItemRepo := repository.NewBookingItemRepository(db)
	guestRepo := repository.NewGuestRepository(db)
	waitingRoomRepo := repository.NewWaitingRoomRepository(db)

	ticketCategoryService := service.NewTicketCategoryService(ticketCategoryRepo)
	ticketStockService := service.NewTicketStockService(ticketStockRepo, outboxEventRepo)
	guestService := service.NewGuestService(guestRepo)
	bookingItemService := service.NewBookingItemService(bookingItemRepo)

	bookingService := service.NewBookingService(
		txManager,
		bookingRepo,
		ticketCategoryService,
		ticketStockService,
		guestService,
		bookingItemService,
		waitingRoomRepo,
	)

	waitingRoomService := service.NewWaitingRoomService(
		txManager,
		waitingRoomRepo,
		ticketCategoryService,
		ticketStockService,
		publisher,
	)

	return &Container{
		BookingService:     bookingService,
		WaitingRoomService: waitingRoomService,
		TicketStockService: ticketStockService,
		OutboxEventRepo:    outboxEventRepo,
	}
}
