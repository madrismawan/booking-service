package seeder

import (
	"booking-service/internal/model"

	"gorm.io/gorm"
)

func seedTicketStocks(tx *gorm.DB, categories []model.TicketCategory) error {
	for _, category := range categories {
		stock := model.TicketStock{
			TicketCategoryID:  category.ID,
			TotalQuantity:     10000,
			AvailableQuantity: 10000,
			ReservedQuantity:  0,
			SoldQuantity:      0,
		}

		if err := tx.Where("ticket_category_id = ?", category.ID).Attrs(stock).FirstOrCreate(&stock).Error; err != nil {
			return err
		}
	}

	return nil
}
