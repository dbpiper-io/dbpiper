package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SyncDirection string

const (
	AirtableToPg  SyncDirection = "airtable_to_pgx"
	PgToAirtable  SyncDirection = "pgx_to_airtable"
	Bidirectional SyncDirection = "two_way"
)

type SyncStatus string
type RepoType string

const (
	SyncSetup      SyncStatus = "setup"
	SyncInstalling SyncStatus = "installing"
	SyncActive     SyncStatus = "active"
	SyncPaused     SyncStatus = "paused"
	SyncError      SyncStatus = "error"
)

const (
	Pgx      RepoType = "pgx"
	Airtable RepoType = "airtable"
)

type Sync struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID string    `gorm:"index;not null"`

	// Source
	SourceType   RepoType // "airtable" | "postgres"
	SourceConnID string

	// Target
	TargetType   RepoType // "airtable" | "postgres"
	TargetConnID string

	// Direction
	Direction SyncDirection // "one_way" | "two_way"

	Tables datatypes.JSON
	/*
	  [
	    {
	      "source_table": "users",
	      "target_table": "tblAbc123",
	      "fields": {
	        "id": "fldID",
	        "email": "fldEmail"
	      }
	    }
	  ]
	*/

	// State
	Status SyncStatus // setup | active | paused | error

	LastError sql.NullString

	CreatedAt time.Time
	UpdatedAt time.Time
}
