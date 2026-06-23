package repository

import (
	"context"
	"errors"
	"time"

	"booking-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxEventRepository struct {
	db *gorm.DB
}

func NewOutboxEventRepository(db *gorm.DB) *OutboxEventRepository {
	return &OutboxEventRepository{db: db}
}

func (r *OutboxEventRepository) WithTx(tx *gorm.DB) *OutboxEventRepository {
	return &OutboxEventRepository{db: tx}
}

func (r *OutboxEventRepository) Create(event *model.OutboxEvent) error {
	return r.db.Create(event).Error
}

func (r *OutboxEventRepository) ClaimNext(
	ctx context.Context,
	processingTimeout time.Duration,
) (*model.OutboxEvent, error) {
	var claimed model.OutboxEvent

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		staleBefore := time.Now().Add(-processingTimeout)
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where(
				"(status = ? AND next_attempt_at <= ?) OR (status = ? AND updated_at <= ?)",
				model.OutboxStatusPending,
				time.Now(),
				model.OutboxStatusProcessing,
				staleBefore,
			).
			Order("id ASC").
			First(&claimed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		claimed.Status = model.OutboxStatusProcessing
		claimed.Attempts++
		claimed.LastError = nil
		return tx.Model(&claimed).Updates(map[string]any{
			"status":     claimed.Status,
			"attempts":   claimed.Attempts,
			"last_error": nil,
			"updated_at": time.Now(),
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return &claimed, nil
}

func (r *OutboxEventRepository) MarkSent(ctx context.Context, id int64, processedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":          model.OutboxStatusSent,
			"processed_at":    processedAt,
			"last_error":      nil,
			"next_attempt_at": processedAt,
			"updated_at":      processedAt,
		}).Error
}

func (r *OutboxEventRepository) MarkRetry(
	ctx context.Context,
	id int64,
	lastError string,
	nextAttemptAt time.Time,
) error {
	return r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":          model.OutboxStatusPending,
			"last_error":      lastError,
			"next_attempt_at": nextAttemptAt,
			"updated_at":      time.Now(),
		}).Error
}
