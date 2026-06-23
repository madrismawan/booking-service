package main

import (
	"log"

	"booking-service/internal/app"
	"booking-service/internal/config"
	"booking-service/internal/database"
	"booking-service/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	container := app.NewContainer(db, cfg.Payment.WebhookSecret)
	engine := router.New(container, cfg.CORSAllowedOrigins)

	if err := engine.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
