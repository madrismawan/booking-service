package handler

import (
	"booking-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingService *service.BookingService
}

func New(bookingService *service.BookingService) *Handler {
	return &Handler{bookingService: bookingService}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/booking/:slug", h.createBooking)
}
