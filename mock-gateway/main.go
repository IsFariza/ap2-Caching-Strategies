package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

type notifyRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

type gateway struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func main() {
	g := &gateway{seen: make(map[string]struct{})}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /notify", g.notify)

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Mock Notification Gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("gateway stopped: %v", err)
	}
}

func (g *gateway) notify(w http.ResponseWriter, r *http.Request) {
	var req notifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logJSON(map[string]any{
		"time":            time.Now().Format(time.RFC3339),
		"idempotency_key": req.IdempotencyKey,
		"channel":         req.Channel,
		"recipient":       req.Recipient,
		"message":         req.Message,
	})

	if rand.Intn(100) < 20 {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"temporary_failure"}`))
		return
	}

	g.mu.Lock()
	_, duplicate := g.seen[req.IdempotencyKey]
	if !duplicate {
		g.seen[req.IdempotencyKey] = struct{}{}
	}
	g.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if duplicate {
		_, _ = w.Write([]byte(`{"status":"duplicate"}`))
		return
	}
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

func logJSON(value any) {
	data, err := json.Marshal(value)
	if err != nil {
		fmt.Println(`{"error":"failed to marshal log"}`)
		return
	}
	fmt.Println(string(data))
}
