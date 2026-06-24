package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"booking-service/internal/dto"
	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/repository"

	"gorm.io/gorm"
)

type PaymentWebhookResult struct {
	Payment   *model.PaymentTransaction
	Booking   *model.Booking
	Duplicate bool
}

type PaymentService struct {
	txManager       *repository.TransactionManager
	paymentRepo     *repository.PaymentTransactionRepository
	bookingRepo     *repository.BookingRepository
	bookingItemRepo *repository.BookingItemRepository
	ticketStock     *TicketStockService
	outboxRepo      *repository.OutboxEventRepository
	webhookSecret   string
}

func NewPaymentService(
	txManager *repository.TransactionManager,
	paymentRepo *repository.PaymentTransactionRepository,
	bookingRepo *repository.BookingRepository,
	bookingItemRepo *repository.BookingItemRepository,
	ticketStock *TicketStockService,
	outboxRepo *repository.OutboxEventRepository,
	webhookSecret string,
) *PaymentService {
	return &PaymentService{
		txManager:       txManager,
		paymentRepo:     paymentRepo,
		bookingRepo:     bookingRepo,
		bookingItemRepo: bookingItemRepo,
		ticketStock:     ticketStock,
		outboxRepo:      outboxRepo,
		webhookSecret:   webhookSecret,
	}
}

func (s *PaymentService) VerifySignature(body []byte, signature string) bool {
	signature = strings.TrimSpace(strings.TrimPrefix(signature, "sha256="))
	if signature == "" || s.webhookSecret == "" {
		return false
	}

	provided, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}

func (s *PaymentService) ProcessWebhook(
	req dto.PaymentWebhookRequest,
	rawPayload []byte,
) (*PaymentWebhookResult, error) {
	if !req.Valid() {
		return nil, repository.ErrInvalidPayment
	}

	var result PaymentWebhookResult
	err := s.txManager.Transaction(func(tx *gorm.DB) error {
		lockKey := req.Provider + "|ref|" + req.RefID
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
			lockKey,
		).Error; err != nil {
			return err
		}

		paymentRepo := s.paymentRepo.WithTx(tx)
		existing, err := paymentRepo.FindByRefID(req.Provider, req.RefID)
		if err == nil {
			booking, bookingErr := s.bookingRepo.WithTx(tx).FindByID(existing.BookingID)
			if bookingErr != nil {
				return bookingErr
			}
			result = PaymentWebhookResult{
				Payment:   existing,
				Booking:   booking,
				Duplicate: true,
			}
			return nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return err
		}

		bookingRepo := s.bookingRepo.WithTx(tx)
		booking, err := bookingRepo.FindByIDForUpdate(req.BookingID)
		if err != nil {
			return err
		}
		if booking.Status != "pending_payment" ||
			req.Amount != booking.TotalPrice ||
			req.PaidAt.After(booking.ExpiresAt) {
			return repository.ErrPaymentConflict
		}

		transactionCode, err := newPaymentTransactionCode()
		if err != nil {
			return err
		}

		payment := &model.PaymentTransaction{
			BookingID:       booking.ID,
			TransactionCode: transactionCode,
			Provider:        req.Provider,
			RefID:           req.RefID,
			PaymentMethod:   req.PaymentMethod,
			Status:          "paid",
			Amount:          req.Amount,
			Payload:         append(json.RawMessage(nil), rawPayload...),
			PaidAt:          &req.PaidAt,
		}
		if err := paymentRepo.Create(payment); err != nil {
			return err
		}

		booking.Status = "paid"
		booking.PaidAt = &req.PaidAt
		if err := bookingRepo.Save(booking); err != nil {
			return err
		}

		items, err := s.bookingItemRepo.WithTx(tx).FindByBookingID(booking.ID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return repository.ErrPaymentConflict
		}

		ticketStock := s.ticketStock.WithTx(tx)
		for _, item := range items {
			if _, err := ticketStock.MarkSoldForUpdate(item.TicketCategoryID, item.Quantity); err != nil {
				return err
			}
		}

		accountingPayload, err := json.Marshal(rabbitmq.AccountingPaymentSucceededPayload{
			PaymentTransactionID: payment.ID,
			BookingID:            booking.ID,
			BookingCode:          booking.BookingCode,
			Amount:               payment.Amount,
			PaidAt:               req.PaidAt,
		})
		if err != nil {
			return err
		}

		if err := s.outboxRepo.WithTx(tx).Create(&model.OutboxEvent{
			AggregateType: "payment_transaction",
			AggregateID:   payment.ID,
			EventType:     rabbitmq.AccountingPaymentSucceededEventType,
			Payload:       accountingPayload,
			Status:        model.OutboxStatusPending,
			NextAttemptAt: time.Now(),
		}); err != nil {
			return err
		}

		result = PaymentWebhookResult{
			Payment: payment,
			Booking: booking,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func newPaymentTransactionCode() (string, error) {
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate payment transaction code: %w", err)
	}
	return "PAY-" + strings.ToUpper(hex.EncodeToString(random[:])), nil
}
