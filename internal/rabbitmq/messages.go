package rabbitmq

import "time"

type WaitingRoomMessage struct {
	TicketCategoryID int64     `json:"ticket_category_id"`
	QueueToken       string    `json:"queue_token"`
	CreatedAt        time.Time `json:"created_at"`
}

const (
	WaitingRoomJoinedEventType     = "waiting_room.joined"
	WaitingRoomJoinedSchemaVersion = 1
)

type WaitingRoomJoinedPayload struct {
	TicketCategoryID int64     `json:"ticket_category_id"`
	QueueToken       string    `json:"queue_token"`
	CreatedAt        time.Time `json:"created_at"`
}

const (
	AccountingPaymentSucceededEventType     = "accounting.payment_succeeded"
	AccountingPaymentSucceededSchemaVersion = 1
)

type AccountingPaymentSucceededPayload struct {
	PaymentTransactionID int64     `json:"payment_transaction_id"`
	BookingID            int64     `json:"booking_id"`
	BookingCode          string    `json:"booking_code"`
	Amount               int64     `json:"amount"`
	PaidAt               time.Time `json:"paid_at"`
}

type AccountingPaymentSucceededMessage struct {
	EventID              int64     `json:"event_id"`
	EventType            string    `json:"event_type"`
	SchemaVersion        int       `json:"schema_version"`
	PaymentTransactionID int64     `json:"payment_transaction_id"`
	BookingID            int64     `json:"booking_id"`
	BookingCode          string    `json:"booking_code"`
	Amount               int64     `json:"amount"`
	PaidAt               time.Time `json:"paid_at"`
	Attempt              int       `json:"attempt"`
	LastError            string    `json:"last_error,omitempty"`
}

const (
	TicketStockChangedEventType     = "ticket_stock.changed"
	TicketStockChangedSchemaVersion = 1
)

type TicketStockChangedPayload struct {
	EventType        string    `json:"event_type"`
	SchemaVersion    int       `json:"schema_version"`
	TicketCategoryID int64     `json:"ticket_category_id"`
	StockVersion     int64     `json:"stock_version"`
	SnapshotURL      string    `json:"snapshot_url"`
	ChangedAt        time.Time `json:"changed_at"`
}

type TicketStockChangedMessage struct {
	EventID          int64     `json:"event_id"`
	EventType        string    `json:"event_type"`
	SchemaVersion    int       `json:"schema_version"`
	TicketCategoryID int64     `json:"ticket_category_id"`
	StockVersion     int64     `json:"stock_version"`
	SnapshotURL      string    `json:"snapshot_url"`
	ChangedAt        time.Time `json:"changed_at"`
}

func ShouldRefreshTicketStock(localVersion, eventVersion int64) bool {
	return eventVersion > localVersion
}
