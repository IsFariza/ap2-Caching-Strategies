package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/cache"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/middleware"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/repository"
	transport "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/usecase"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	db, err := sql.Open("postgres", requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := pingDB(ctx, db); err != nil {
		log.Fatalf("connect database: %v", err)
	}
	runMigrations(db)

	nc, err := nats.Connect(envOrDefault("NATS_URL", nats.DefaultURL))
	if err != nil {
		log.Fatalf("connect nats: %v", err)
	}
	defer nc.Drain()

	baseRepo := repository.NewDoctorRepository(db)
	cacheRepo := cache.NewRedisRepository(ctx)
	doctorRepo := repository.NewCachedDoctorRepository(baseRepo, cacheRepo)
	publisher := event.NewDoctorPublisher(nc)
	uc := usecase.NewDoctorUseCase(doctorRepo, publisher)
	handler := transport.NewDoctorHandler(uc)
	limiter := middleware.NewRateLimiter(ctx)

	port := envOrDefault("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(limiter.UnaryServerInterceptor()))
	doctorpb.RegisterDoctorServiceServer(server, handler)
	log.Printf("Doctor Service listening on :%s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func pingDB(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully")
}
