package test

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
)

func TestDoc4ConcurrentDuplicateWebhookIsProcessedOnce(t *testing.T) {
	fixture := newFixture(t, 2)
	booking := fixture.createPendingBooking(t, 1)
	payload := map[string]any{
		"provider":       "doku",
		"ref_id":         "doc-4-duplicate-ref",
		"booking_id":     booking.ID,
		"payment_method": "ewallet",
		"status":         "paid",
		"amount":         booking.TotalPrice,
		"paid_at":        time.Now().UTC().Format(time.RFC3339Nano),
	}

	type result struct {
		status   int
		response paymentWebhookResponse
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			status, response := sendPaymentWebhook(t, fixture.Router, payload)
			results <- result{status: status, response: response}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var duplicateCount, initialCount int
	var paymentID int64
	for result := range results {
		if result.status != http.StatusOK {
			t.Fatalf("expected both requests to return HTTP 200, got %d", result.status)
		}
		if result.response.Data.Duplicate {
			duplicateCount++
		} else {
			initialCount++
		}
		if paymentID == 0 {
			paymentID = result.response.Data.PaymentTransactionID
		} else if paymentID != result.response.Data.PaymentTransactionID {
			t.Fatalf("duplicate returned a different payment ID")
		}
	}
	if initialCount != 1 || duplicateCount != 1 {
		t.Fatalf("expected one initial and one duplicate response, got initial=%d duplicate=%d", initialCount, duplicateCount)
	}

	var paymentCount int64
	if err := fixture.DB.Model(&model.PaymentTransaction{}).
		Where("provider = ? AND ref_id = ?", "doku", "doc-4-duplicate-ref").
		Count(&paymentCount).Error; err != nil {
		t.Fatalf("count duplicate payments: %v", err)
	}
	if paymentCount != 1 {
		t.Fatalf("expected one stored payment, got %d", paymentCount)
	}

	var accountingOutboxCount int64
	if err := fixture.DB.Model(&model.OutboxEvent{}).
		Where("aggregate_type = ? AND aggregate_id = ? AND event_type = ?",
			"payment_transaction",
			paymentID,
			rabbitmq.AccountingPaymentSucceededEventType,
		).
		Count(&accountingOutboxCount).Error; err != nil {
		t.Fatalf("count duplicate accounting outbox events: %v", err)
	}
	if accountingOutboxCount != 1 {
		t.Fatalf("expected one accounting outbox event, got %d", accountingOutboxCount)
	}

	var stock model.TicketStock
	if err := fixture.DB.First(&stock, fixture.Stock.ID).Error; err != nil {
		t.Fatalf("load stock after duplicate webhook: %v", err)
	}
	if stock.ReservedQuantity != 0 || stock.SoldQuantity != 1 {
		t.Fatalf("stock was updated more than once: reserved=%d sold=%d", stock.ReservedQuantity, stock.SoldQuantity)
	}
}
