package models

import (
	"time"
)

type AirtableConnection struct {
	ID            int    `gorm:"primaryKey"`
	UserID        int    `gorm:"index"`
	User          User   `gorm:"constraint:OnDelete:CASCADE"`
	AccessToken   string // encrypted
	RefreshToken  string // encrypted
	WebhookSecret string
	CreatedAt     time.Time `gorm:"autoCreateTime"`
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
