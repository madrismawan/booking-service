package test

import (
	"net/http"
	"testing"
	"time"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
)

func TestDoc3PaymentWebhookUpdatesBookingStockAndAccountingOutbox(t *testing.T) {
	fixture := newFixture(t, 5)
	booking := fixture.createPendingBooking(t, 2)

	status, response := sendPaymentWebhook(t, fixture.Router, map[string]any{
		"provider":       "midtrans",
		"ref_id":         "doc-3-payment-ref",
		"booking_id":     booking.ID,
		"payment_method": "virtual_account",
		"status":         "paid",
		"amount":         booking.TotalPrice,
		"paid_at":        time.Now().UTC().Format(time.RFC3339Nano),
	})
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if !response.Success || response.Data.Duplicate {
		t.Fatalf("expected successful non-duplicate response: %+v", response)
	}

	var storedBooking model.Booking
	if err := fixture.DB.First(&storedBooking, booking.ID).Error; err != nil {
		t.Fatalf("load paid booking: %v", err)
	}
	if storedBooking.Status != "paid" || storedBooking.PaidAt == nil {
		t.Fatalf("booking was not paid: status=%s paid_at=%v", storedBooking.Status, storedBooking.PaidAt)
	}

	var stock model.TicketStock
	if err := fixture.DB.First(&stock, fixture.Stock.ID).Error; err != nil {
		t.Fatalf("load sold stock: %v", err)
	}
	if stock.AvailableQuantity != 3 || stock.ReservedQuantity != 0 || stock.SoldQuantity != 2 {
		t.Fatalf(
			"unexpected stock after payment: available=%d reserved=%d sold=%d",
			stock.AvailableQuantity,
			stock.ReservedQuantity,
			stock.SoldQuantity,
		)
	}

	var paymentCount int64
	if err := fixture.DB.Model(&model.PaymentTransaction{}).
		Where("provider = ? AND ref_id = ?", "midtrans", "doc-3-payment-ref").
		Count(&paymentCount).Error; err != nil {
		t.Fatalf("count payment transaction: %v", err)
	}
	if paymentCount != 1 {
		t.Fatalf("expected one payment transaction, got %d", paymentCount)
	}

	var accountingOutboxCount int64
	if err := fixture.DB.Model(&model.OutboxEvent{}).
		Where(
			"aggregate_type = ? AND aggregate_id = ? AND event_type = ?",
			"payment_transaction",
			response.Data.PaymentTransactionID,
			rabbitmq.AccountingPaymentSucceededEventType,
		).
		Count(&accountingOutboxCount).Error; err != nil {
		t.Fatalf("count accounting outbox: %v", err)
	}
	if accountingOutboxCount != 1 {
		t.Fatalf("expected one accounting outbox event, got %d", accountingOutboxCount)
	}
}
