package databases

import (
	"context"
	"database/sql"
	"dbpiper/types"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/url"
	"strings"
)

func BuildPostgresDSN(req types.DBConnectRequest) (string, error) {
	if req.ConnectionURL != "" {
		return req.ConnectionURL, nil
	}

	sslMode := "disable"
	if req.UseSSL {
		sslMode = "require"
	}

	if req.Port != "" {
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			req.Username,
			url.QueryEscape(req.Password),
			req.Host,
			req.Port,
			req.Database,
			sslMode,
		), nil
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		req.Username,
		url.QueryEscape(req.Password),
		req.Host,
		req.Database,
		sslMode,
	), nil
}

func TestConnection(ctx context.Context, driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("Ping error: %w", err)
	}

	return nil
}

func DetectSSL(req types.DBConnectRequest) bool {
	if req.ConnectionURL != "" {
		return strings.Contains(req.ConnectionURL, "sslmode=require")
	}
	return req.UseSSL
}
