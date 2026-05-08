package subscriber

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
	"github.com/nats-io/nats.go"
)

func HandleEvents(url string, subjects []string) *nats.Conn {
	nc := connect(url)
	handler := func(m *nats.Msg) {
		var payload interface{}
		if err := json.Unmarshal(m.Data, &payload); err != nil {
			log.Printf("failed to parse JSON: %v", err)
			return
		}
		output := models.NotificationLog{
			Time:    time.Now().Format(time.RFC3339),
			Subject: m.Subject,
			Event:   payload,
		}

		jsonLog, _ := json.Marshal(output)
		fmt.Println((string(jsonLog)))
	}
	for _, s := range subjects {
		nc.Subscribe(s, handler)
	}
	return nc
}

func connect(url string) *nats.Conn {
	delay := 1 * time.Second
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(url)
		if err == nil {
			return nc
		}
		log.Printf("Retry %d: %v", i+1, err)
		time.Sleep(delay)
		delay *= 2
	}
	os.Exit(1)
	return nil
}
