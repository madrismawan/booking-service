package model

import "time"

type Event struct {
	ID           int64 `gorm:"primaryKey"`
	Slug         string
	Name         string
	Description  string
	VenueName    string
	VenueAddress string
	StartsAt     time.Time
	EndsAt       *time.Time
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
