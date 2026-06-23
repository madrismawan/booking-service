package model

import (
	"encoding/json"
	"time"
)

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusSent       = "sent"
)

type OutboxEvent struct {
	ID            int64 `gorm:"primaryKey"`
	AggregateType string
	AggregateID   int64
	EventType     string
	Payload       json.RawMessage `gorm:"type:jsonb"`
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	ProcessedAt   *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
