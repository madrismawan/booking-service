package model

import "time"

type BookingItem struct {
	ID                        int64 `gorm:"primaryKey"`
	BookingID                 int64
	TicketCategoryID          int64
	TicketCategoryName        string
	TicketCategoryDescription string
	Quantity                  int
	UnitPrice                 int64
	SubtotalPrice             int64
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
