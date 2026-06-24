package dto

import "time"

type PaymentWebhookRequest struct {
	Provider      string    `json:"provider"`
	RefID         string    `json:"ref_id"`
	BookingID     int64     `json:"booking_id"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	Amount        int64     `json:"amount"`
	PaidAt        time.Time `json:"paid_at"`
}

func (r PaymentWebhookRequest) Valid() bool {
	return r.Provider != "" &&
		r.RefID != "" &&
		r.BookingID > 0 &&
		r.PaymentMethod != "" &&
		r.Status == "paid" &&
		r.Amount > 0 &&
		!r.PaidAt.IsZero()
}

type PaymentWebhookResponse struct {
	PaymentTransactionID int64  `json:"payment_transaction_id"`
	TransactionCode      string `json:"transaction_code"`
	BookingID            int64  `json:"booking_id"`
	BookingStatus        string `json:"booking_status"`
	PaymentStatus        string `json:"payment_status"`
	Duplicate            bool   `json:"duplicate"`
}
