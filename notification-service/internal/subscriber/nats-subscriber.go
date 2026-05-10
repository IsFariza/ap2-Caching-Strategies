package subscriber

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/logger"
	"github.com/nats-io/nats.go"
)

func HandleEvents(url string, subjects []string, onEvent func(subject string, data []byte)) *nats.Conn {
	nc := connect(url)
	handler := func(m *nats.Msg) {
		var payload interface{}
		if err := json.Unmarshal(m.Data, &payload); err != nil {
			log.Printf("failed to parse JSON: %v", err)
			return
		}
		logger.LogEvent(m.Subject, payload)
		if onEvent != nil {
			onEvent(m.Subject, m.Data)
		}
	}
	for _, s := range subjects {
		if _, err := nc.Subscribe(s, handler); err != nil {
			log.Printf("subscribe subject=%s error=%v", s, err)
		}
	}
	return nc
}

func connect(url string) *nats.Conn {
	if url == "" {
		url = nats.DefaultURL
	}
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
