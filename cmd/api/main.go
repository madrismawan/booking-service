package main

import (
	"log"

	"booking-service/internal/app"
	"booking-service/internal/config"
	"booking-service/internal/database"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	mq, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Printf("connect rabbitmq: %v", err)
	} else {
		defer mq.Close()
	}

	container := app.NewContainer(db, mq)
	engine := router.New(container, cfg.CORSAllowedOrigins)

	if err := engine.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
