package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/repository"
	doctorGRPC "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/usecase"
	doctorpb "github.com/IsFariza/ap2-Caching-Strategies/doctor-service/proto"
	"github.com/joho/godotenv"
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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	runMigrations(db)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Printf("NATS unavailable: %v", err)
	} else {
		defer nc.Close()
	}
	repo := repository.NewDoctorRepository(db)
	pub := event.NewDoctorPublisher(nc)
	uc := usecase.NewDoctorUseCase(repo, pub)
	handler := doctorGRPC.NewDoctorHandler(uc)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
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
