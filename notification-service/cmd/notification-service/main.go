package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/jobqueue"
	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/subscriber"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	queue := jobqueue.New(ctx)
	subjects := []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated"}

	nc := subscriber.HandleEvents(os.Getenv("NATS_URL"), subjects, func(subject string, data []byte) {
		queue.EnqueueFromEvent(ctx, subject, data)
	})
	defer nc.Drain()

	log.Println("Notification Service running ")
	<-ctx.Done()
}
