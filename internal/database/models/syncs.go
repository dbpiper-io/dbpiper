package models

import (
	"gorm.io/datatypes"
	"time"
)

type Sync struct {
	ID     string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID int    `gorm:"index"`
	User   User   `gorm:"constraint:OnDelete:CASCADE"`

	AirtableConnectionID int
	AirtableConnection   AirtableConnection `gorm:"constraint:OnDelete:SET NULL"`

	DatabaseConnectionID int
	DatabaseConnection   DatabaseConnection `gorm:"constraint:OnDelete:SET NULL"`

	AirtableBaseID    string
	AirtableTableID   string
	AirtableTableName string
	AirtableWebhookID string

	DatabaseTableName        string
	DatabaseTriggerInstalled bool

	FieldMappings datatypes.JSON // JSONB {"pg_col": "airtable_field"}

	SyncDirection      string // db_to_airtable, airtable_to_db, bidirectional
	ConflictResolution string `gorm:"default:last_write_wins"`

	IsActive     bool       `gorm:"default:true"`
	LastSyncedAt *time.Time // nullable
	Status       string     `gorm:"default:active"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type SyncLog struct {
	ID     int    `gorm:"primaryKey"`
	SyncID string `gorm:"type:uuid;index"`
	Sync   Sync   `gorm:"constraint:OnDelete:CASCADE"`

	Direction      string // airtable_to_db, db_to_airtable
	Status         string // success, error, conflict
	RecordsAdded   int
	RecordsUpdated int
	RecordsDeleted int

	ErrorMessage    string
	ExecutionTimeMs int
	SyncedAt        time.Time `gorm:"autoCreateTime"`
}
