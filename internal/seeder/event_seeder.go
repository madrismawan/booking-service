package seeder

import (
	"time"

	"booking-service/internal/model"

	"gorm.io/gorm"
)

func seedEvent(tx *gorm.DB) (model.Event, error) {
	event := model.Event{
		Slug:         "tripla-live-concert",
		Name:         "Tripla Live Concert",
		Description:  "Konser demo untuk flow booking tiket.",
		VenueName:    "Jakarta Convention Center",
		VenueAddress: "Jl. Gatot Subroto, Jakarta",
		StartsAt:     time.Date(2026, 8, 1, 19, 0, 0, 0, time.FixedZone("WIB", 7*60*60)),
		Status:       "published",
	}

	if err := tx.Where("slug = ?", event.Slug).Attrs(event).FirstOrCreate(&event).Error; err != nil {
		return model.Event{}, err
	}

	return event, nil
}
