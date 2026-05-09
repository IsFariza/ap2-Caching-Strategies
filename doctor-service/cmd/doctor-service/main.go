package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/middleware"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/repository"
	doctorGRPC "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/usecase"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	natsURL := os.Getenv("NATS_URL")
	port := os.Getenv("PORT")
	redisURL := os.Getenv("REDIS_URL")

	var db *sql.DB
	var err error
	log.Printf("Connecting to doctor-db at %s...", dbURL)

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("Successfully connected to doctor-db!")
			break
		}

		log.Printf("Doctor-db not ready (attempt %d/10): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}
	defer db.Close()
	runMigrations(db)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Printf("NATS unavailable: %v", err)
	} else {
		defer nc.Close()
	}

	ttlStr := os.Getenv("CACHE_TTL_SECONDS")
	ttl, _ := strconv.Atoi(ttlStr)
	if ttl == 0 {
		ttl = 60
	}

	opts, err := redis.ParseURL(redisURL)
	var rdb *redis.Client
	if err == nil {
		rdb = redis.NewClient(opts)
	} else {
		log.Printf("Redis unavailable: %v", err)
	}

	rawRepo := repository.NewDoctorRepository(db)
	repo := repository.NewDoctorCacheProxy(rawRepo, rdb, ttl)
	pub := event.NewDoctorPublisher(nc)
	uc := usecase.NewDoctorUseCase(repo, pub)
	handler := doctorGRPC.NewDoctorHandler(uc)

	limitRPM, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_RPM"))

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RateLimitInterceptor(rdb, limitRPM)),
	)
	doctorpb.RegisterDoctorServiceServer(s, handler)

	log.Printf("Doctor Service starting on port %s...", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully!")
}
