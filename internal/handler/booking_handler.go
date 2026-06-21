package handler

import (
	"net/http"

	"booking-service/internal/dto"
	"booking-service/internal/mapper"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createBooking(ctx *gin.Context) {
	var req dto.CreateBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid request payload",
			Errors:  nil,
		})
		return
	}

	booking, err := h.bookingService.CreateBooking(ctx.Param("slug"), req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Message: "booking created",
		Data:    mapper.ToCreateBookingResponse(booking),
	})
}
