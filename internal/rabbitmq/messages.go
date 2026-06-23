package rabbitmq

import "time"

type WaitingRoomMessage struct {
	TicketCategoryID int64     `json:"ticket_category_id"`
	QueueToken       string    `json:"queue_token"`
	CreatedAt        time.Time `json:"created_at"`
}

type AccountingPaymentSucceededMessage struct {
	PaymentTransactionID int64     `json:"payment_transaction_id"`
	BookingID            int64     `json:"booking_id"`
	BookingCode          string    `json:"booking_code"`
	Amount               int64     `json:"amount"`
	PaidAt               time.Time `json:"paid_at"`
}
