package server

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"dbpiper/database"
	"dbpiper/internal/databases/pgx"
)

type Server struct {
	Port    int
	PgxPool *pgx.PoolManager
	DB      database.DB
}

func NewServer(serv *Server) *http.Server {

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serv.Port),
		Handler:      serv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
