package model

import (
	"encoding/json"
	"time"
)

type PaymentTransaction struct {
	ID              int64 `gorm:"primaryKey"`
	BookingID       int64
	TransactionCode string `gorm:"not null;uniqueIndex"`
	Provider        string
	RefID           string
	PaymentMethod   string
	Status          string
	Amount          int64
	Payload         json.RawMessage `gorm:"type:jsonb"`
	PaidAt          *time.Time
	ExpiredAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
