package test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
)

func TestDoc5StockVersionOutboxAndETagStaySynchronized(t *testing.T) {
	fixture := newFixture(t, 10)

	updated, err := fixture.App.TicketStockService.ReserveForUpdate(fixture.Category.ID, 2)
	if err != nil {
		t.Fatalf("reserve stock: %v", err)
	}
	if updated.Version != 2 || updated.AvailableQuantity != 8 || updated.ReservedQuantity != 2 {
		t.Fatalf(
			"unexpected updated stock: version=%d available=%d reserved=%d",
			updated.Version,
			updated.AvailableQuantity,
			updated.ReservedQuantity,
		)
	}

	var outboxCount int64
	if err := fixture.DB.Model(&model.OutboxEvent{}).
		Where("aggregate_type = ? AND aggregate_id = ? AND event_type = ?",
			"ticket_stock",
			updated.ID,
			rabbitmq.TicketStockChangedEventType,
		).
		Count(&outboxCount).Error; err != nil {
		t.Fatalf("count stock outbox events: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("expected one stock outbox event, got %d", outboxCount)
	}

	path := "/api/v1/ticket-categories/" + strconv.FormatInt(fixture.Category.ID, 10) + "/stock"
	firstRequest := httptest.NewRequest(http.MethodGet, path, nil)
	firstResponse := httptest.NewRecorder()
	fixture.Router.ServeHTTP(firstResponse, firstRequest)
	if firstResponse.Code != http.StatusOK {
		t.Fatalf("expected stock snapshot HTTP 200, got %d", firstResponse.Code)
	}

	expectedETag := `"ticket-stock-` + strconv.FormatInt(fixture.Category.ID, 10) + `-v2"`
	if etag := firstResponse.Header().Get("ETag"); etag != expectedETag {
		t.Fatalf("expected ETag %s, got %s", expectedETag, etag)
	}

	cachedRequest := httptest.NewRequest(http.MethodGet, path, nil)
	cachedRequest.Header.Set("If-None-Match", expectedETag)
	cachedResponse := httptest.NewRecorder()
	fixture.Router.ServeHTTP(cachedResponse, cachedRequest)
	if cachedResponse.Code != http.StatusNotModified {
		t.Fatalf("expected cached snapshot HTTP 304, got %d", cachedResponse.Code)
	}

	if rabbitmq.ShouldRefreshTicketStock(2, 2) {
		t.Fatal("consumer must ignore duplicate stock version")
	}
	if !rabbitmq.ShouldRefreshTicketStock(2, 3) {
		t.Fatal("consumer must refresh for a newer stock version")
	}
}
