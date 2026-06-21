package seeder

import (
	"errors"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

func seedTicketCategory(tx *gorm.DB, event model.Event) (model.TicketCategory, error) {
	category := model.TicketCategory{
		EventID:       event.ID,
		Name:          "Festival",
		Description:   "Tiket festival standing area.",
		Price:         750000,
		MaxPerBooking: 4,
	}

	var existingCategory model.TicketCategory
	err := tx.Where("event_id = ? AND name = ?", event.ID, category.Name).First(&existingCategory).Error
	switch {
	case err == nil:
		return existingCategory, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := tx.Create(&category).Error; err != nil {
			return model.TicketCategory{}, err
		}
		return category, nil
	default:
		return model.TicketCategory{}, err
	}
}
