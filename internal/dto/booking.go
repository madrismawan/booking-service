package dto

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type CreateBookingRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CreateBookingResponse struct {
	ID           int64                 `json:"id"`
	BookingCode  string                `json:"booking_code"`
	Status       string                `json:"status"`
	GuestName    string                `json:"guest_name"`
	EventName    string                `json:"event_name"`
	TotalTicket  int                   `json:"total_ticket"`
	TotalPrice   int64                 `json:"total_price"`
	ExpiresAt    time.Time             `json:"expires_at"`
	BookingItems []BookingItemResponse `json:"booking_items"`
}

type BookingItemResponse struct {
	TicketCategoryName string `json:"ticket_category_name"`
	Quantity           int    `json:"quantity"`
	UnitPrice          int64  `json:"unit_price"`
	SubtotalPrice      int64  `json:"subtotal_price"`
}
