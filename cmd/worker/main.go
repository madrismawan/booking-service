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

	consumerMQ, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect rabbitmq consumer: %v", err)
	}
	defer consumerMQ.Close()

	publisherMQ, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect rabbitmq publisher: %v", err)
	}
	defer publisherMQ.Close()

	container := app.NewContainer(db, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitingRoomWorker := worker.NewWaitingRoomWorker(
		consumerMQ,
		container.WaitingRoomService,
		log.Default(),
	)
	outboxWorker := worker.NewOutboxWorker(
		container.OutboxEventRepo,
		publisherMQ,
		log.Default(),
	)

	errCh := make(chan error, 2)
	go func() {
		errCh <- waitingRoomWorker.Start(ctx)
	}()
	go func() {
		errCh <- outboxWorker.Start(ctx)
	}()

	var workerErr error
	for completed := 0; completed < 2; completed++ {
		err := <-errCh
		if err != nil && workerErr == nil {
			workerErr = err
			stop()
		}
	}

	if workerErr != nil {
		log.Fatalf("run worker: %v", workerErr)
	}
	log.Println("workers shutting down")
}
