package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WaitingRoomRepository struct {
	db *gorm.DB
}

func NewWaitingRoomRepository(db *gorm.DB) *WaitingRoomRepository {
	return &WaitingRoomRepository{db: db}
}

func (r *WaitingRoomRepository) WithTx(tx *gorm.DB) *WaitingRoomRepository {
	return &WaitingRoomRepository{db: tx}
}

func (r *WaitingRoomRepository) Create(waitingRoom *model.WaitingRoom) error {
	return r.db.Create(waitingRoom).Error
}

func (r *WaitingRoomRepository) FindByQueueToken(queueToken string) (*model.WaitingRoom, error) {
	var waitingRoom model.WaitingRoom
	err := r.db.Where("queue_token = ?", queueToken).First(&waitingRoom).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &waitingRoom, err
}

func (r *WaitingRoomRepository) FindByQueueTokenForUpdate(queueToken string) (*model.WaitingRoom, error) {
	var waitingRoom model.WaitingRoom
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("queue_token = ?", queueToken).
		First(&waitingRoom).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &waitingRoom, err
}

func (r *WaitingRoomRepository) FindByCheckoutTokenForUpdate(checkoutToken string) (*model.WaitingRoom, error) {
	var waitingRoom model.WaitingRoom
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("checkout_token = ?", checkoutToken).
		First(&waitingRoom).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &waitingRoom, err
}

func (r *WaitingRoomRepository) CountActiveCheckoutByTicketCategoryID(ticketCategoryID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.WaitingRoom{}).
		Where("ticket_category_id = ?", ticketCategoryID).
		Where("status IN ?", []string{
			model.WaitingRoomStatusReady,
			model.WaitingRoomStatusCheckoutStarted,
		}).
		Where("expired_at IS NOT NULL AND expired_at > NOW()").
		Count(&count).Error
	return count, err
}

func (r *WaitingRoomRepository) Save(waitingRoom *model.WaitingRoom) error {
	return r.db.Save(waitingRoom).Error
}
