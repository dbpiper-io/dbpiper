package database

import (
	"context"
	"database/sql"
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

const (
	idAndUserId = "id = ? AND user_id = ?"
)

// Service represents a service that interacts with a database.
type DB interface {
	WithTx(fn func(tx DB) error) error
	UpsertAirtableConnection(ctx context.Context, conn *models.AirtableConnection) error
	GetAirtableConnections(ctx context.Context, userID string) ([]models.AirtableConnection, error)
	DeleteAirtableConnection(ctx context.Context, userID, id string) error
	CreateDatabaseConnection(ctx context.Context, db *models.DatabaseConnection) error
	DeleteDatabaseConnection(ctx context.Context, userID, id string) error
	GetDatabaseConnections(ctx context.Context, userID string) ([]models.DatabaseConnection, error)
	GetDatabaseConnectionByID(ctx context.Context, userID, id string) (*models.DatabaseConnection, error)
	GetAirtableConnectionByID(ctx context.Context, userID, id string) (*models.AirtableConnection, error)
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
	// Build field updates depending on connection type
	updates := map[string]any{
		"user_id":         conn.UserID,
		"connection_type": conn.ConnectionType,
		"created_at":      conn.CreatedAt,
		"base_id":         conn.BaseID,
	}

	if conn.ConnectionType == models.OAuth {
		updates["provider_account_id"] = conn.ProviderAccountID
		updates["access_token"] = conn.AccessToken
		updates["refresh_token"] = conn.RefreshToken
		updates["scope"] = conn.Scope
		updates["token_type"] = conn.TokenType
		updates["expires_at"] = conn.ExpiresAt

		// CLEAR API key fields
		updates["api_key"] = sql.NullString{}
	}

	if conn.ConnectionType == models.APIKey {
		updates["api_key"] = conn.APIKey

		// CLEAR OAuth fields
		updates["provider_account_id"] = sql.NullString{}
		updates["access_token"] = sql.NullString{}
		updates["refresh_token"] = sql.NullString{}
		updates["scope"] = sql.NullString{}
		updates["token_type"] = sql.NullString{}
		updates["expires_at"] = sql.NullTime{}
	}
	return s.db.WithContext(ctx).
		Model(&models.AirtableConnection{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(updates).
		Error
}

func (s *service) GetAirtableConnections(ctx context.Context, userID string) ([]models.AirtableConnection, error) {
	var airtable []models.AirtableConnection
	if err := s.db.WithContext(ctx).
		Model(&models.AirtableConnection{}).
		Where("user_id = ?", userID).
		Find(&airtable).Error; err != nil {
		return nil, err
	}
	return airtable, nil
}

func (s *service) GetDatabaseConnections(ctx context.Context, userID string) ([]models.DatabaseConnection, error) {
	var db []models.DatabaseConnection
	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseConnection{}).
		Where("user_id = ?", userID).
		Find(&db).Error; err != nil {
		return nil, err
	}
	return db, nil
}

func (s *service) DeleteAirtableConnection(ctx context.Context, userID, id string) error {
	return s.db.WithContext(ctx).
		Delete(&models.AirtableConnection{}, idAndUserId, id, userID).Error
}

func (s *service) CreateDatabaseConnection(ctx context.Context, db *models.DatabaseConnection) error {
	return s.db.WithContext(ctx).
		Model(&models.DatabaseConnection{}).
		Create(db).Error
}

func (s *service) DeleteDatabaseConnection(ctx context.Context, userID, id string) error {
	return s.db.WithContext(ctx).
		Delete(&models.DatabaseConnection{}, idAndUserId, id, userID).Error
}

func (s *service) GetDatabaseConnectionByID(ctx context.Context, userID, id string) (*models.DatabaseConnection, error) {
	var db *models.DatabaseConnection
	if err := s.db.WithContext(ctx).
		Model(&models.DatabaseConnection{}).
		Where(idAndUserId, id, userID).
		First(&db).
		Error; err != nil {
		return nil, err
	}
	return db, nil
}

func (s *service) GetAirtableConnectionByID(ctx context.Context, userID, id string) (*models.AirtableConnection, error) {
	var airtable models.AirtableConnection
	if err := s.db.WithContext(ctx).
		Model(&models.AirtableConnection{}).
		Where(idAndUserId, id, userID).
		First(&airtable).Error; err != nil {
		return nil, err
	}
	return &airtable, nil
}
