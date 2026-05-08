package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	pb "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/appt_proto"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/client"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/event"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/repository"
	apptGRPC "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
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
		log.Printf("NATS connection failed: %v", err)
	} else {
		defer nc.Close()
	}

	doctorConn, err := grpc.NewClient(doctorServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to Doctor Service: %v", err)
	}
	defer doctorConn.Close()

	doctorClient := client.NewDoctorClient(doctorConn)
	repo := repository.NewAppointmentRepository(db)
	pub := event.NewAppointmentPublisher(nc)
	uc := usecase.NewAppointmentUsecase(repo, doctorClient, pub)
	handler := apptGRPC.NewAppointmentHandler(uc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcPort, err)
	}

	s := grpc.NewServer()
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
