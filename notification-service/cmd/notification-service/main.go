package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/subscriber"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	subjects := []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated"}

	nc := subscriber.HandleEvents(os.Getenv("NATS_URL"), subjects)
	defer nc.Drain()

	log.Println("Notification Service running ")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
