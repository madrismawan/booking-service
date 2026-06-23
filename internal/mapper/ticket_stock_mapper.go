package mapper

import (
	"booking-service/internal/dto"
	"booking-service/internal/model"
)

func ToTicketStockResponse(stock *model.TicketStock) dto.TicketStockResponse {
	return dto.TicketStockResponse{
		TicketCategoryID:  stock.TicketCategoryID,
		TotalQuantity:     stock.TotalQuantity,
		AvailableQuantity: stock.AvailableQuantity,
		ReservedQuantity:  stock.ReservedQuantity,
		SoldQuantity:      stock.SoldQuantity,
		Version:           stock.Version,
		UpdatedAt:         stock.UpdatedAt,
	}
}
