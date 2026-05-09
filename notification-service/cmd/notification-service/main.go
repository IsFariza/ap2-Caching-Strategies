package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/jobqueue"
	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/subscriber"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	natsURL := os.Getenv("NATS_URL")

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	wp := jobqueue.NewWorkerPool(rdb, 100)
	wp.Start(5)

	subjects := []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated"}

	nc := subscriber.HandleEvents(natsURL, subjects, wp)
	defer nc.Drain()

	log.Println("Notification Service running ")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
