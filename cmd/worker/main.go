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

	waitingRoomMQ, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect waiting room rabbitmq consumer: %v", err)
	}
	defer waitingRoomMQ.Close()

	accountingMQ, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect accounting rabbitmq consumer: %v", err)
	}
	defer accountingMQ.Close()

	publisherMQ, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("connect rabbitmq publisher: %v", err)
	}
	defer publisherMQ.Close()

	container := app.NewContainer(db, cfg.Payment.WebhookSecret)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitingRoomWorker := worker.NewWaitingRoomWorker(
		waitingRoomMQ,
		container.WaitingRoomService,
		log.Default(),
	)
	outboxWorker := worker.NewOutboxWorker(
		container.OutboxEventRepo,
		publisherMQ,
		log.Default(),
	)
	accountingWorker := worker.NewAccountingWorker(
		accountingMQ,
		publisherMQ,
		cfg.Accounting,
		log.Default(),
	)

	errCh := make(chan error, 3)
	go func() {
		errCh <- waitingRoomWorker.Start(ctx)
	}()
	go func() {
		errCh <- outboxWorker.Start(ctx)
	}()
	go func() {
		errCh <- accountingWorker.Start(ctx)
	}()

	var workerErr error
	for completed := 0; completed < 3; completed++ {
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
