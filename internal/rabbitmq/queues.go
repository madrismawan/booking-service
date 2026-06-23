package rabbitmq

import (
	"fmt"
	"strings"
)

const (
	WaitingRoomQueue                = "waiting_room.queue"
	AccountingPaymentSucceededQueue = "accounting.payment_succeeded.queue"
	TicketStockChangedQueue         = "ticket_stock.changed.queue"
)

func WaitingRoomEventQueue(eventSlug string) string {
	slug := strings.TrimSpace(eventSlug)
	if slug == "" {
		return WaitingRoomQueue
	}
	return fmt.Sprintf("%s.%s", WaitingRoomQueue, slug)
}
