package repository

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInsufficientStock = errors.New("insufficient ticket stock")
	ErrQueuePublish      = errors.New("failed to publish queue message")
	ErrInvalidCheckout   = errors.New("invalid checkout token")
	ErrExpiredCheckout   = errors.New("checkout token expired")
)
