package databases

import (
	"context"
	"database/sql"
	"dbpiper/types"
	"fmt"
	"net/url"
	"time"
)

func BuildPostgresDSN(req types.DBConnectRequest) string {
	sslMode := "disable"
	if req.UseSSL {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		req.Username,
		url.QueryEscape(req.Password),
		req.Host,
		req.Port,
		req.Database,
		sslMode,
	)
}

func TestConnection(ctx context.Context, driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
