package mapper

import (
	"booking-service/internal/dto"
	"booking-service/internal/model"
)

func ToJoinQueueResponse(waitingRoom *model.WaitingRoom) dto.JoinQueueResponse {
	return dto.JoinQueueResponse{
		QueueToken:       waitingRoom.QueueToken,
		TicketCategoryID: waitingRoom.TicketCategoryID,
		Status:           waitingRoom.Status,
	}
}

func ToQueueStatusResponse(waitingRoom *model.WaitingRoom) dto.QueueStatusResponse {
	return dto.QueueStatusResponse{
		QueueToken:       waitingRoom.QueueToken,
		TicketCategoryID: waitingRoom.TicketCategoryID,
		Status:           waitingRoom.Status,
		BookingID:        waitingRoom.BookingID,
		CheckoutToken:    waitingRoom.CheckoutToken,
		ExpiredAt:        waitingRoom.ExpiredAt,
		FailedReason:     waitingRoom.FailedReason,
	}
}
