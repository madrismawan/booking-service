package test

import (
	"sync"
	"testing"

	"booking-service/internal/model"
	"booking-service/internal/rabbitmq"
)

func TestDoc2HighTrafficQueuePersistsEveryRequestAndOutboxEvent(t *testing.T) {
	const requestCount = 50
	fixture := newFixture(t, requestCount)

	start := make(chan struct{})
	errors := make(chan error, requestCount)
	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := fixture.App.WaitingRoomService.JoinQueue(fixture.Category.ID)
			errors <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("join queue under high traffic: %v", err)
		}
	}

	var waitingRoomCount int64
	if err := fixture.DB.Model(&model.WaitingRoom{}).
		Where("ticket_category_id = ?", fixture.Category.ID).
		Count(&waitingRoomCount).Error; err != nil {
		t.Fatalf("count waiting rooms: %v", err)
	}
	if waitingRoomCount != requestCount {
		t.Fatalf("expected %d waiting rooms, got %d", requestCount, waitingRoomCount)
	}

	var outboxCount int64
	if err := fixture.DB.Model(&model.OutboxEvent{}).
		Where("aggregate_type = ? AND event_type = ?", "waiting_room", rabbitmq.WaitingRoomJoinedEventType).
		Count(&outboxCount).Error; err != nil {
		t.Fatalf("count waiting-room outbox events: %v", err)
	}
	if outboxCount != requestCount {
		t.Fatalf("expected %d outbox events, got %d", requestCount, outboxCount)
	}
}
