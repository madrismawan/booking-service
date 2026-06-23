package dto

import "time"

type TicketStockResponse struct {
	TicketCategoryID  int64     `json:"ticket_category_id"`
	TotalQuantity     int       `json:"total_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
	SoldQuantity      int       `json:"sold_quantity"`
	Version           int64     `json:"version"`
	UpdatedAt         time.Time `json:"updated_at"`
}
