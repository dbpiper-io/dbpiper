package airtable

import (
	"context"
	"dbpiper/database"
	"dbpiper/database/models"
	"os"
)

const (
	metaBasesURL = "https://api.airtable.com/v0/meta/bases"

//	revokeOauthURL = "https://www.airtable.com/oauth2/v1/revoke"

/* 	metaTablesURL = "https://api.airtable.com/v0/meta/bases/%s/tables" */
)

type Client interface {
	GetBases(ctx context.Context) ([]Base, error)
	GetTables(ctx context.Context, baseID string) ([]Table, error)
}

type Airtable struct {
	Conn         *models.AirtableConnection
	DB           database.DB
	clientSecret string
	clientId     string
}

func NewClient(conn *models.AirtableConnection, db database.DB) Client {
	return &Airtable{
		Conn:         conn,
		DB:           db,
		clientSecret: os.Getenv("AIRTABLE_CLIENT_SECRET"),
		clientId:     os.Getenv("AIRTABLE_CLIENT_ID"),
	}
}

func (a *Airtable) GetBases(ctx context.Context) ([]Base, error) {
	var data struct {
		Bases []Base `json:"bases"`
	}

	if err := a.doRequest(ctx, "GET", metaBasesURL, nil, &data); err != nil {
		return nil, err
	}

	return data.Bases, nil
}

func (a *Airtable) GetTables(ctx context.Context, baseID string) ([]Table, error) {
	//todo
	return nil, nil
}
