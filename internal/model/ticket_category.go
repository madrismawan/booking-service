package model

import "time"

type TicketCategory struct {
	ID            int64 `gorm:"primaryKey"`
	EventID       int64
	Name          string
	Description   string
	Price         int64
	SaleStartsAt  *time.Time
	SaleEndsAt    *time.Time
	MaxPerBooking int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Event         Event `gorm:"foreignKey:EventID"`
}
