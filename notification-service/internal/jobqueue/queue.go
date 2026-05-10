package jobqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
	"github.com/redis/go-redis/v9"
)

const idempotencyTTL = 24 * time.Hour

type Queue struct {
	jobs       chan models.Job
	client     *redis.Client
	redisOK    bool
	gatewayURL string
	httpClient *http.Client
	mu         sync.Mutex
	localKeys  map[string]string
}

type statusUpdatedEvent struct {
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	ID         string `json:"id"`
	DoctorID   string `json:"doctor_id"`
	NewStatus  string `json:"new_status"`
}

func New(ctx context.Context) *Queue {
	wp_size, _ := strconv.Atoi(os.Getenv("WORKER_POOL_SIZE"))

	queue := &Queue{
		jobs:       make(chan models.Job, wp_size*10),
		gatewayURL: strings.TrimRight(os.Getenv("GATEWAY_URL"), "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
		localKeys:  make(map[string]string),
	}

	if url := os.Getenv("REDIS_URL"); url != "" {
		opts, err := redis.ParseURL(url)
		if err != nil {
			opts = &redis.Options{Addr: url}
		}
		queue.client = redis.NewClient(opts)
		if err := queue.client.Ping(ctx).Err(); err != nil {
			log.Printf("warning: redis unavailable: %v", err)
		} else {
			queue.redisOK = true
		}
	} else {
		log.Printf("warning: REDIS_URL is empty; notification idempotency will be in-memory only")
	}

	for i := 0; i < wp_size; i++ {
		go queue.worker(ctx)
	}
	return queue
}

func (q *Queue) EnqueueFromEvent(ctx context.Context, subject string, data []byte) {
	if subject != "appointments.status_updated" {
		return
	}

	var event statusUpdatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		q.log(os.Stderr, "error", "", 1, "dead_letter", err)
		return
	}
	if event.NewStatus != "done" {
		return
	}
	if event.EventType == "" {
		event.EventType = subject
	}
	if event.DoctorID == "" {
		event.DoctorID = "unknown"
	}

	key := idempotencyKey(event.EventType, event.ID, event.OccurredAt)
	status, err := q.getKey(ctx, key)
	if err != nil {
		q.log(os.Stdout, "warn", key, 1, "retry", err)
	}
	if status == "done" || status == "queued" {
		q.log(os.Stdout, "info", key, 1, "duplicate", nil)
		return
	}
	if ok, err := q.reserveKey(ctx, key); err != nil {
		q.log(os.Stdout, "warn", key, 1, "retry", err)
	} else if !ok {
		q.log(os.Stdout, "info", key, 1, "duplicate", nil)
		return
	}

	job := models.Job{
		IdempotencyKey: key,
		AppointmentID:  event.ID,
		DoctorID:       event.DoctorID,
		OccurredAt:     event.OccurredAt,
		Channel:        "email",
		Recipient:      "patient@clinic.kz",
		Message:        fmt.Sprintf("Your appointment %s with doctor %s is complete.", event.ID, event.DoctorID),
	}

	select {
	case q.jobs <- job:
		q.log(os.Stdout, "info", key, 1, "enqueued", nil)
	default:
		q.log(os.Stderr, "error", key, 1, "dead_letter", errors.New("job queue is full"))
	}
}

func (q *Queue) callGateway(ctx context.Context, job models.Job) error {
	body, _ := json.Marshal(job)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.gatewayURL+"/notify", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("gateway returned 503")
	}
	return fmt.Errorf("gateway returned status %d", resp.StatusCode)
}

func (q *Queue) log(writer *os.File, level, jobID string, attempt int, status string, err error) {
	entry := models.JobLog{
		Time:    time.Now().Format(time.RFC3339),
		Level:   level,
		JobID:   jobID,
		Attempt: attempt,
		Status:  status,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(writer, string(data))
}
