package seeder

import (
	"booking-service/internal/model"

	"gorm.io/gorm"
)

func seedTicketCategories(tx *gorm.DB, event model.Event) ([]model.TicketCategory, error) {
	seedCategories := []model.TicketCategory{
		{
			EventID:       event.ID,
			Name:          "Festival",
			Description:   "Tiket festival standing area.",
			Price:         750000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "VIP",
			Description:   "Tiket VIP dengan area terbaik.",
			Price:         1500000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "VVIP",
			Description:   "Tiket VVIP dengan akses eksklusif.",
			Price:         2500000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Early Bird Festival",
			Description:   "Tiket festival harga early bird.",
			Price:         500000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Early Bird VIP",
			Description:   "Tiket VIP harga early bird.",
			Price:         1200000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Presale Festival",
			Description:   "Tiket festival harga presale.",
			Price:         650000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Presale VIP",
			Description:   "Tiket VIP harga presale.",
			Price:         1350000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Tribune A",
			Description:   "Tiket duduk area Tribune A.",
			Price:         900000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Tribune B",
			Description:   "Tiket duduk area Tribune B.",
			Price:         800000,
			MaxPerBooking: 5,
		},
		{
			EventID:       event.ID,
			Name:          "Backstage Pass",
			Description:   "Tiket dengan akses backstage.",
			Price:         3500000,
			MaxPerBooking: 5,
		},
	}

	categories := make([]model.TicketCategory, 0, len(seedCategories))
	for _, category := range seedCategories {
		if err := tx.Where("event_id = ? AND name = ?", event.ID, category.Name).Attrs(category).FirstOrCreate(&category).Error; err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
