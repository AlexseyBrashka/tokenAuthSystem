package suite

import (
	"AuthService/internal/service"
	"AuthService/internal/storage/postgre"
	"fmt"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

var (
	JWT_SECRET        = "Pushkin"
	DB_HOST           = "127.0.0.1"
	DB_PORT           = "5432"
	DB_USER           = "postgres"
	DB_PASSWORD       = "553782"
	DB_NAME           = "AuthDB"
	DB_SSLMODE        = "disable"
	SERVER_PORT       = "8080"
	ACCESS_TOKEN_TTL  = "15m"
	REFRESH_TOKEN_TTL = "24h"
)

type mockEmailService struct{}

func (m *mockEmailService) SendIPChangeWarning(userID uuid.UUID, oldIP, newIP string) error {
	log.Printf("Mock email sent to user %s: IP changed from %s to %s", userID, oldIP, newIP)
	return nil
}

func NewAuthService(t *testing.T) *service.Service {
	dbConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		DB_HOST,
		DB_PORT,
		DB_USER,
		DB_PASSWORD,
		DB_NAME,
		DB_SSLMODE,
	)

	storage, err := postgre.New(dbConn)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	accessTTL, err := time.ParseDuration(ACCESS_TOKEN_TTL)
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_TTL: %v", err)
	}

	refreshTTL, err := time.ParseDuration(REFRESH_TOKEN_TTL)
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_TTL: %v", err)
	}

	if JWT_SECRET == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	authService := service.New(
		storage,
		[]byte(JWT_SECRET),
		accessTTL,
		refreshTTL,
		&mockEmailService{},
	)
	return authService
}
