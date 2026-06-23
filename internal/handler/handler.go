package handler

import (
	"booking-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingService     *service.BookingService
	waitingRoomService *service.WaitingRoomService
	ticketStockService *service.TicketStockService
	paymentService     *service.PaymentService
}

func New(
	bookingService *service.BookingService,
	waitingRoomService *service.WaitingRoomService,
	ticketStockService *service.TicketStockService,
	paymentService *service.PaymentService,
) *Handler {
	return &Handler{
		bookingService:     bookingService,
		waitingRoomService: waitingRoomService,
		ticketStockService: ticketStockService,
		paymentService:     paymentService,
	}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/ticket-categories/:ticket_category_id/queue/join", h.joinQueue)
	api.GET("/ticket-categories/:ticket_category_id/stock", h.getTicketStock)
	api.GET("/queue/:queue_token/status", h.getQueueStatus)
	api.POST("/booking", h.createBooking)
	api.POST("/payments/webhook", h.paymentWebhook)
}
