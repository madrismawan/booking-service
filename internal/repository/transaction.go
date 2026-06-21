package repository

import "gorm.io/gorm"

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (m *TransactionManager) Transaction(fn func(tx *gorm.DB) error) error {
	return m.db.Transaction(fn)
}
