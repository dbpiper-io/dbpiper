package models

import (
	"time"

	"gorm.io/datatypes"
)

type WebhookQueue struct {
    ID          int            `gorm:"primaryKey"`
    SyncID      string         `gorm:"type:uuid;index"`
    Sync        Sync           `gorm:"constraint:OnDelete:CASCADE"`

    Payload     datatypes.JSON // JSONB
    Source      string         // airtable or database
    Attempts    int            `gorm:"default:0"`
    MaxAttempts int            `gorm:"default:3"`

    NextRetryAt *time.Time
    CreatedAt   time.Time      `gorm:"autoCreateTime"`
}
