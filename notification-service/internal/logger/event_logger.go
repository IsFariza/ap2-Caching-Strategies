package logger

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
)

func LogEvent(subject string, event any) {
	output := models.NotificationLog{
		Time:    time.Now().Format(time.RFC3339),
		Subject: subject,
		Event:   event,
	}

	data, _ := json.Marshal(output)
	fmt.Println(string(data))
}
