package main

import (
	"log"

	"booking-service/internal/config"
	"booking-service/internal/database"
	"booking-service/internal/seeder"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := seeder.Run(db); err != nil {
		log.Fatalf("run seeder: %v", err)
	}

	log.Println("seed completed")
}
