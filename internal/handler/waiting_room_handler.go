package handler

import (
	"net/http"
	"strconv"

	"booking-service/internal/dto"
	"booking-service/internal/mapper"

	"github.com/gin-gonic/gin"
)

func (h *Handler) joinQueue(ctx *gin.Context) {
	ticketCategoryID, err := strconv.ParseInt(ctx.Param("ticket_category_id"), 10, 64)
	if err != nil || ticketCategoryID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid ticket category id",
			Errors:  nil,
		})
		return
	}

	waitingRoom, err := h.waitingRoomService.JoinQueue(ticketCategoryID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, dto.Response{
		Success: true,
		Message: "you are in queue",
		Data:    mapper.ToJoinQueueResponse(waitingRoom),
	})
}

func (h *Handler) getQueueStatus(ctx *gin.Context) {
	waitingRoom, err := h.waitingRoomService.GetStatus(ctx.Param("queue_token"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	message := "still waiting"
	switch waitingRoom.Status {
	case "ready":
		message = "ready for checkout"
	case "checkout_started":
		message = "order is being processed"
	case "expired":
		message = "checkout token expired"
	case "completed":
		message = "booking completed"
	case "failed":
		message = "queue failed"
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: message,
		Data:    mapper.ToQueueStatusResponse(waitingRoom),
	})
}
