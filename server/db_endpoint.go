package server

import (
	"database/sql"
	"dbpiper/database/models"
	"dbpiper/internal/databases/pgx"
	"dbpiper/types"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) addDBConnectionEndPoint(g *echo.Group) {
	conns := g.Group("/databases")
	conns.POST("/connect", s.connectDatabase)
	conns.DELETE("/:id", s.deleteDatabaseConnection)
	conn := conns.Group("/:id")
	tables := conn.Group("/tables")
	tables.GET("", s.getTables)
	table := tables.Group("/:table")
	table.GET("/columns", s.GetTableColumns)
}

func (s *Server) connectDatabase(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not authenticated"})
	}

	var req types.DBConnectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request", "details": err.Error()})
	}

	if req.Engine != string(models.Postgres) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "unsupported database"})
	}
	dsn := req.ConnectionURL
	if dsn != "" {
		dsn = pgx.BuildPostgresDSN(req.Username, req.Password, req.Host, req.Port, req.Database, req.UseSSL)
	}

	if err := pgx.TestConnection(ctx, req.Engine, dsn); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "failed to connect to database", "details": err.Error()})
	}

	var host, db, user, pass, port string
	var sslEnabled bool
	if req.ConnectionURL != "" {
		sslEnabled = strings.Contains(req.ConnectionURL, "sslmode=require")
		u, err := url.Parse(req.ConnectionURL)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid url", "details": err.Error()})
		}
		user = u.User.Username()
		pass, _ = u.User.Password()
		host = u.Hostname()
		port = u.Port()
		db = strings.TrimPrefix(u.Path, "/")
	} else {
		user = req.Username
		pass = req.Password
		host = req.Host
		port = req.Port
		db = req.Database
		sslEnabled = req.UseSSL
	}

	var err error
	var portInt int

	if port != "" {
		portInt, err = strconv.Atoi(port)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid database port", "details": err.Error()})
		}
	}

	// Save to DB
	conn := models.DatabaseConnection{
		UserID:        userID,
		Engine:        models.Engine(req.Engine),
		Host:          host,
		Port:          portInt,
		DatabaseName:  db,
		Username:      user,
		Password:      pass,
		ConnectionURL: sql.NullString{String: req.ConnectionURL, Valid: req.ConnectionURL != ""},
		SSLEnabled:    sslEnabled,
	}

	if err := s.DB.CreateDatabaseConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "database connected"})
}

func (s *Server) deleteDatabaseConnection(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not authenticated"})
	}

	if err := s.DB.DeleteDatabaseConnection(ctx, userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Database removed"})
}

func (s *Server) getTables(c echo.Context) error {
	ctx := c.Request().Context()
	connID := c.Param("id")

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	db, err := s.DB.GetDatabaseConnectionByID(ctx, userID, connID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "connection not found", "details": err.Error()})
	}

	dsn := db.ConnectionURL.String
	if !db.ConnectionURL.Valid {
		dsn = pgx.BuildPostgresDSN(db.Username, db.Password, db.Host, strconv.Itoa(db.Port), db.DatabaseName, db.SSLEnabled)
	}
	pool, err := s.PgxPool.GetPool(ctx, connID, dsn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "pool error", "details": err.Error()})
	}
	rows, err := pool.Query(ctx, pgx.AllTables)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "query error", "details": err.Error()})
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"id":     connID,
		"driver": db.Engine,
		"tables": tables,
	})
}

func (s *Server) GetTableColumns(c echo.Context) error {
	ctx := c.Request().Context()

	connID := c.Param("id")
	table := c.Param("table")

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	db, err := s.DB.GetDatabaseConnectionByID(ctx, userID, connID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "connection not found", "details": err.Error()})
	}

	dsn := db.ConnectionURL.String
	if !db.ConnectionURL.Valid {
		dsn = pgx.BuildPostgresDSN(db.Username, db.Password, db.Host, strconv.Itoa(db.Port), db.DatabaseName, db.SSLEnabled)
	}
	pool, err := s.PgxPool.GetPool(ctx, connID, dsn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "pool error", "details": err.Error()})
	}
	rows, err := pool.Query(ctx, pgx.ColumnType, table)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "query error", "details": err.Error()})
	}
	defer rows.Close()
	type ColumnInfo struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	var result []ColumnInfo

	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "scan failed", "details": err.Error()})
		}
		result = append(result, col)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"table":   table,
		"columns": result,
	})
}
