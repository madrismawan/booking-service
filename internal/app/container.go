package app

import (
	"booking-service/internal/repository"
	"booking-service/internal/service"

	"gorm.io/gorm"
)

type Container struct {
	BookingService     *service.BookingService
	WaitingRoomService *service.WaitingRoomService
	TicketStockService *service.TicketStockService
	PaymentService     *service.PaymentService
	OutboxEventRepo    *repository.OutboxEventRepository
}

func NewContainer(db *gorm.DB, webhookSecret string) *Container {
	txManager := repository.NewTransactionManager(db)

	ticketCategoryRepo := repository.NewTicketCategoryRepository(db)
	ticketStockRepo := repository.NewTicketStockRepository(db)
	outboxEventRepo := repository.NewOutboxEventRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingItemRepo := repository.NewBookingItemRepository(db)
	paymentRepo := repository.NewPaymentTransactionRepository(db)
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
		outboxEventRepo,
		ticketCategoryService,
		ticketStockService,
	)
	paymentService := service.NewPaymentService(
		txManager,
		paymentRepo,
		bookingRepo,
		bookingItemRepo,
		ticketStockService,
		outboxEventRepo,
		webhookSecret,
	)

	return &Container{
		BookingService:     bookingService,
		WaitingRoomService: waitingRoomService,
		TicketStockService: ticketStockService,
		PaymentService:     paymentService,
		OutboxEventRepo:    outboxEventRepo,
	}
}
