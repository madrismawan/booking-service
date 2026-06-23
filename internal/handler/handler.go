package handler

import (
	"booking-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingService     *service.BookingService
	waitingRoomService *service.WaitingRoomService
}

func New(bookingService *service.BookingService, waitingRoomService *service.WaitingRoomService) *Handler {
	return &Handler{
		bookingService:     bookingService,
		waitingRoomService: waitingRoomService,
	}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/ticket-categories/:ticket_category_id/queue/join", h.joinQueue)
	api.GET("/queue/:queue_token/status", h.getQueueStatus)
	api.POST("/booking", h.createBooking)
}
