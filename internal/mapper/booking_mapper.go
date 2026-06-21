package mapper

import (
	"booking-service/internal/dto"
	"booking-service/internal/model"
)

func ToCreateBookingResponse(booking *model.Booking) dto.CreateBookingResponse {
	items := make([]dto.BookingItemResponse, 0, len(booking.Items))
	for _, item := range booking.Items {
		items = append(items, dto.BookingItemResponse{
			TicketCategoryName: item.TicketCategoryName,
			Quantity:           item.Quantity,
			UnitPrice:          item.UnitPrice,
			SubtotalPrice:      item.SubtotalPrice,
		})
	}

	return dto.CreateBookingResponse{
		ID:           booking.ID,
		BookingCode:  booking.BookingCode,
		Status:       booking.Status,
		GuestName:    booking.GuestName,
		EventName:    booking.EventName,
		TotalTicket:  booking.TotalTicket,
		TotalPrice:   booking.TotalPrice,
		ExpiresAt:    booking.ExpiresAt,
		BookingItems: items,
	}
}
