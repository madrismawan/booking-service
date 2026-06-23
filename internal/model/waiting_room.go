package model

import "time"

const (
	WaitingRoomStatusWaiting         = "waiting"
	WaitingRoomStatusReady           = "ready"
	WaitingRoomStatusExpired         = "expired"
	WaitingRoomStatusCheckoutStarted = "checkout_started"
	WaitingRoomStatusCompleted       = "completed"
	WaitingRoomStatusFailed          = "failed"
)

type WaitingRoom struct {
	ID               int64 `gorm:"primaryKey"`
	EventID          int64
	EventName        string
	TicketCategoryID int64
	QueueToken       string
	CheckoutToken    *string
	BookingID        *int64
	Status           string
	FailedReason     *string
	ExpiredAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
