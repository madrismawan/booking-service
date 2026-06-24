package test

import (
	"errors"
	"sync"
	"testing"

	"booking-service/internal/dto"
	"booking-service/internal/model"
	"booking-service/internal/repository"
)

func TestDoc1ConcurrentBookingDoesNotOversell(t *testing.T) {
	fixture := newFixture(t, 1)
	first := fixture.createReadyWaitingRoom(t, "race-1")
	second := fixture.createReadyWaitingRoom(t, "race-2")

	requests := []dto.CreateBookingRequest{
		{
			CheckoutToken:    *first.CheckoutToken,
			TicketCategoryID: fixture.Category.ID,
			Quantity:         1,
		},
		{
			CheckoutToken:    *second.CheckoutToken,
			TicketCategoryID: fixture.Category.ID,
			Quantity:         1,
		},
	}

	start := make(chan struct{})
	results := make(chan error, len(requests))
	var wg sync.WaitGroup
	for _, request := range requests {
		request := request
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := fixture.App.BookingService.CreateBooking(request)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var successCount, insufficientStockCount int
	for err := range results {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, repository.ErrInsufficientStock):
			insufficientStockCount++
		default:
			t.Fatalf("unexpected booking error: %v", err)
		}
	}

	if successCount != 1 || insufficientStockCount != 1 {
		t.Fatalf(
			"expected one success and one insufficient-stock result, got success=%d insufficient=%d",
			successCount,
			insufficientStockCount,
		)
	}

	var stock model.TicketStock
	if err := fixture.DB.First(&stock, fixture.Stock.ID).Error; err != nil {
		t.Fatalf("load final stock: %v", err)
	}
	if stock.AvailableQuantity != 0 || stock.ReservedQuantity != 1 || stock.SoldQuantity != 0 {
		t.Fatalf(
			"stock was oversold: available=%d reserved=%d sold=%d",
			stock.AvailableQuantity,
			stock.ReservedQuantity,
			stock.SoldQuantity,
		)
	}
}
