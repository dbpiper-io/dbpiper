package database

import (
	"context"
	"dbpiper/database/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type DB interface {
	WithTx(fn func(tx DB) error) error
	UpsertAirtableConnection(ctx context.Context, conn *models.AirtableConnection) error
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
	)

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

func (s *service) UpsertAirtableConnection(ctx context.Context, conn *models.AirtableConnection) error {
	return s.db.WithContext(ctx).Model(&models.AirtableConnection{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"provider_account_id": conn.ProviderAccountID,
			"access_token":        conn.AccessToken,
			"refresh_token":       conn.RefreshToken,
			"scope":               conn.Scope,
			"token_type":          conn.TokenType,
			"expires_at":          conn.ExpiresAt,
			"updated_at":          conn.UpdatedAt,
		})}).Create(conn).Error
}
