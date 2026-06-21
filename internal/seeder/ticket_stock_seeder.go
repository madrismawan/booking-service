package seeder

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

func seedTicketStock(tx *gorm.DB, category model.TicketCategory) error {
	var existingStock model.TicketStock
	err := tx.Where("ticket_category_id = ?", category.ID).First(&existingStock).Error
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		stock := model.TicketStock{
			TicketCategoryID:  category.ID,
			TotalQuantity:     1000,
			AvailableQuantity: 1000,
			ReservedQuantity:  0,
			SoldQuantity:      0,
		}
		return tx.Create(&stock).Error
	default:
		return err
	}
}
