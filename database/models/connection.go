package models

import (
	"database/sql"
	"time"
)

type ConnectionType string

const (
	APIKey ConnectionType = "api_key"
	OAuth  ConnectionType = "oauth"
)

type AirtableConnection struct {
	ID uint `gorm:"primaryKey"`
	// The user who owns this connection
	UserID string `gorm:"uniqueIndex;not null"`

	// "oauth" or "api_key"
	ConnectionType ConnectionType `gorm:"type:varchar(20);not null"`

	// --- OAuth fields ---
	ProviderAccountID sql.NullString `gorm:"default:null"`
	AccessToken       sql.NullString `gorm:"default:null"`
	RefreshToken      sql.NullString `gorm:"default:null"`
	Scope             sql.NullString `gorm:"default:null"`
	TokenType         sql.NullString `gorm:"default:null"`
	ExpiresAt         sql.NullTime   `gorm:"default:null"`

	// --- API key fields ---
	APIKey sql.NullString `gorm:"default:null"`
	BaseID string `gorm:"default:null"`

	CreatedAt time.Time
}

type DatabaseConnection struct {
	ID     int  `gorm:"primaryKey"`
	UserID int  `gorm:"index"`
	User   User `gorm:"constraint:OnDelete:CASCADE"`

	Type         string // postgres, mysql, mongodb
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string // encrypted
	SSLEnabled   bool
	SSLCert      string

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
