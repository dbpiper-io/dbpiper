package types

import (
	"dbpiper/database/models"
)

type APIKeyConnectRequest struct {
	APIKey string `json:"api_key"`
	BaseID string `json:"base_id"`
}

type AirtableTableResponse struct {
	Tables []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tables"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	AccountID    string `json:"account_id"`
}

type Base struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Table struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type DBConnectRequest struct {
	Engine        string `json:"engine"`
	ConnectionURL string `json:"connection_url"` // optional

	Host     string `json:"host"`
	Port     string `json:"port"` // keep string
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`

	UseSSL bool `json:"use_ssl"`
}

type CreateSyncRequest struct {
	Source SyncEndpoint `json:"source"`
	Target SyncEndpoint `json:"target"`

	Direction string `json:"direction"` // one_way | two_way
	Tables    []TableConfig        `json:"tables"`
}

type SyncEndpoint struct {
	Type         models.RepoType `json:"type"` // pgx | airtable
	ConnectionID string          `json:"connection_id"`
}

type TableConfig struct {
	SourceTable string            `json:"source_table"`
	TargetTable string            `json:"target_table"`
	Fields      map[string]string `json:"fields"`
	// source_column -> target_field_id
}
