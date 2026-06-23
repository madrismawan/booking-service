package model

import "time"

type TicketStock struct {
	ID                int64 `gorm:"primaryKey"`
	TicketCategoryID  int64
	TotalQuantity     int
	AvailableQuantity int
	ReservedQuantity  int
	SoldQuantity      int
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
