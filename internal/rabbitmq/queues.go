package rabbitmq

import (
	"fmt"
	"strings"
)

const (
	WaitingRoomQueue                = "waiting_room.queue"
	AccountingPaymentSucceededQueue = "accounting.payment_succeeded.queue"
	AccountingPaymentRetry5sQueue   = "accounting.payment_succeeded.retry.5s.queue"
	AccountingPaymentRetry10sQueue  = "accounting.payment_succeeded.retry.10s.queue"
	AccountingPaymentSucceededDLQ   = "accounting.payment_succeeded.dlq"
	TicketStockChangedQueue         = "ticket_stock.changed.queue"
)

func WaitingRoomEventQueue(eventSlug string) string {
	slug := strings.TrimSpace(eventSlug)
	if slug == "" {
		return WaitingRoomQueue
	}
	return fmt.Sprintf("%s.%s", WaitingRoomQueue, slug)
}
