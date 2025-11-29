package models

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type AirtableConnection struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID            string `gorm:"uniqueIndex;not null"`
	ProviderAccountID sql.NullString
	AccessToken       string `gorm:"not null"`
	RefreshToken      sql.NullString
	Scope             sql.NullString
	TokenType         sql.NullString
	ExpiresAt         sql.NullTime
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
