package model

import "time"

type Guest struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	Email     string
	Phone     string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
