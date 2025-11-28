package database

import (
	"dbpiper/internal/database/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type DB interface {
	WithTx(fn func(tx DB) error) error
}

type service struct {
	db *gorm.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() DB {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)

	db, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(
    &models.AirtableConnection{},
		&models.DatabaseConnection{},
		&models.Sync{},
		&models.SyncLog{},
		&models.WebhookQueue{})

	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) WithTx(fn func(tx DB) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(&service{db: tx})
	})
}
