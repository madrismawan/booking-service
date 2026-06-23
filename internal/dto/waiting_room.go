package dto

import "time"

type JoinQueueResponse struct {
	QueueToken       string `json:"queue_token"`
	TicketCategoryID int64  `json:"ticket_category_id"`
	Status           string `json:"status"`
}

type QueueStatusResponse struct {
	QueueToken       string     `json:"queue_token"`
	TicketCategoryID int64      `json:"ticket_category_id"`
	Status           string     `json:"status"`
	BookingID        *int64     `json:"booking_id,omitempty"`
	CheckoutToken    *string    `json:"checkout_token,omitempty"`
	ExpiredAt        *time.Time `json:"expired_at,omitempty"`
	FailedReason     *string    `json:"failed_reason,omitempty"`
}
