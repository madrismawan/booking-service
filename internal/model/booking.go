package model

import "time"

type Booking struct {
	ID                int64 `gorm:"primaryKey"`
	BookingCode       string
	GuestID           int64
	GuestName         string
	GuestEmail        string
	GuestPhone        string
	GuestAddress      string
	EventID           int64
	EventSlug         string
	EventName         string
	EventVenueName    string
	EventVenueAddress string
	EventStartsAt     time.Time
	EventEndsAt       *time.Time
	Status            string
	TotalTicket       int
	TotalPrice        int64
	ExpiresAt         time.Time
	PaidAt            *time.Time
	CancelledAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	Items             []BookingItem `gorm:"foreignKey:BookingID"`
}
