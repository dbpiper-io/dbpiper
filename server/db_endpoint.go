package server

import (
	"database/sql"
	"dbpiper/database/models"
	"dbpiper/internal/databases"
	"dbpiper/types"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) addDBConnectionEndPoint(g *echo.Group) {
	conn := g.Group("/databases")
	conn.POST("/connect", s.connectDatabase)
	conn.DELETE("/:id", s.deleteDatabaseConnection)
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

	dsn, err := databases.BuildPostgresDSN(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid configuration", "details": err.Error()})
	}

	if err := databases.TestConnection(ctx, req.Engine, dsn); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "failed to connect to database", "details": err.Error()})
	}

	var host, db, user, pass, port string
	if req.ConnectionURL != "" {
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
	}

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
		SSLEnabled:    databases.DetectSSL(req),
	}

	if err := s.db.CreateDatabaseConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "database connected"})
}

func (s *Server) deleteDatabaseConnection(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not authenticated"})
	}
	id := c.Param("id")
	if err := s.db.DeleteDatabaseConnection(ctx, userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Database removed"})
}
