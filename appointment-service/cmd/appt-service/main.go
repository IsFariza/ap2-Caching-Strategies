package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/appt_proto"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/cache"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/client"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/middleware"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/repository"
	transport "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	doctorConn, err := grpc.NewClient(envOrDefault("DOCTOR_ADDR", "localhost:50051"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect doctor service: %v", err)
	}
	defer doctorConn.Close()

	nc, err := nats.Connect(envOrDefault("NATS_URL", nats.DefaultURL))
	if err != nil {
		log.Fatalf("connect nats: %v", err)
	}
	defer nc.Drain()

	baseRepo := repository.NewAppointmentRepository(db)
	cacheRepo := cache.NewRedisRepository(ctx)
	apptRepo := repository.NewCachedAppointmentRepository(baseRepo, cacheRepo)
	doctorClient := client.NewDoctorClient(doctorConn)
	publisher := event.NewAppointmentPublisher(nc)
	uc := usecase.NewAppointmentUsecase(apptRepo, doctorClient, publisher)
	handler := transport.NewAppointmentHandler(uc)
	limiter := middleware.NewRateLimiter(ctx)

	port := envOrDefault("GRPC_PORT", "50052")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(limiter.UnaryServerInterceptor()))
	pb.RegisterAppointmentServiceServer(server, handler)
	log.Printf("Appointment Service listening on :%s", port)
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
		"postgres", driver)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")
}
