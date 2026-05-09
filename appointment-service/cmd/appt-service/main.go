package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/appt_proto"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/client"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/middleware"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/repository"
	apptGRPC "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	natsURL := os.Getenv("NATS_URL")
	grpcPort := os.Getenv("PORT")
	doctorServiceAddr := os.Getenv("DOCTOR_ADDR")
	redisURL := os.Getenv("REDIS_URL")

	var db *sql.DB
	var err error
	log.Printf("Connecting to database at %s...", dbURL)

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("Successfully connected to the database!")
			break
		}

		log.Printf("Database not ready yet (attempt %d/10): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database after multiple attempts: %v", err)
	}
	defer db.Close()

	runMigrations(db)
	runMigrations(db)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Printf("NATS connection failed: %v", err)
	} else {
		defer nc.Close()
	}

	doctorConn, err := grpc.NewClient(doctorServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to Doctor Service: %v", err)
	}
	defer doctorConn.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL})

	doctorClient := client.NewDoctorClient(doctorConn)
	rawRepo := repository.NewAppointmentRepository(db)
	repo := repository.NewAppointmentCacheProxy(rawRepo, rdb, 60)
	pub := event.NewAppointmentPublisher(nc)
	uc := usecase.NewAppointmentUsecase(repo, doctorClient, pub)
	handler := apptGRPC.NewAppointmentHandler(uc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcPort, err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RateLimitInterceptor(rdb, 100)),
	)
	pb.RegisterAppointmentServiceServer(s, handler)

	log.Printf("Appointment Service starting on port %s", grpcPort)
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
	log.Println("Migrations applied successfully")
}
