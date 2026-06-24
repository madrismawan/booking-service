package repository

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

type PaymentTransactionRepository struct {
	db *gorm.DB
}

func NewPaymentTransactionRepository(db *gorm.DB) *PaymentTransactionRepository {
	return &PaymentTransactionRepository{db: db}
}

func (r *PaymentTransactionRepository) WithTx(tx *gorm.DB) *PaymentTransactionRepository {
	return &PaymentTransactionRepository{db: tx}
}

func (r *PaymentTransactionRepository) Create(payment *model.PaymentTransaction) error {
	return r.db.Create(payment).Error
}

func (r *PaymentTransactionRepository) FindByRefID(
	provider string,
	refID string,
) (*model.PaymentTransaction, error) {
	var payment model.PaymentTransaction
	err := r.db.
		Where("provider = ? AND ref_id = ?", provider, refID).
		First(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &payment, err
}
