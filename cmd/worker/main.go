package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"booking-service/internal/app"
	"booking-service/internal/config"
	"booking-service/internal/database"
	"booking-service/internal/rabbitmq"
	"booking-service/internal/worker"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	mq, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer mq.Close()

	container := app.NewContainer(db, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitingRoomWorker := worker.NewWaitingRoomWorker(mq, container.WaitingRoomService, log.Default())
	if err := waitingRoomWorker.Start(ctx); err != nil {
		log.Fatalf("run waiting room worker: %v", err)
	}

	log.Println("waiting room worker shutting down")
}
