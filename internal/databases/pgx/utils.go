package pgx

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
)

func BuildPostgresDSN(username, password, host, port, database string, useSSL bool) string {
	sslMode := "disable"
	if useSSL {
		sslMode = "require"
	}

	if port != "" {
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			username,
			url.QueryEscape(password),
			host,
			port,
			database,
			sslMode,
		)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		username,
		url.QueryEscape(password),
		host,
		database,
		sslMode,
	)
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
