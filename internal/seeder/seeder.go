package seeder

import (
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		event, err := seedEvent(tx)
		if err != nil {
			return err
		}

		categories, err := seedTicketCategories(tx, event)
		if err != nil {
			return err
		}

		return seedTicketStocks(tx, categories)
	})
}
