package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	port := os.Getenv("GATEWAY_PORT")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var payload interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		log.Printf("[%s] recieved external notification: %v", time.Now().Format(time.Kitchen), payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	fmt.Printf("Mock gateway listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
