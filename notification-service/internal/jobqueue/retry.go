package jobqueue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
)

func (wp *WorkerPool) executewithRetry(job models.Job) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		lastErr = wp.callExternalAPI(job)
		if lastErr == nil {
			return nil
		}
		time.Sleep(time.Duration(1<<uint(i)) * time.Second)
	}
	return lastErr
}

func (wp *WorkerPool) callExternalAPI(job models.Job) error {
	url := "https://httpmock/post"
	body, _ := json.Marshal(job)
	res, err := wp.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("external API status: %d", res.StatusCode)
	}
	return nil
}
