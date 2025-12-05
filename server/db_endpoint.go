package server

import (
	"dbpiper/database/models"
	"dbpiper/internal/databases"
	"dbpiper/types"
	"net/http"
	"strconv"

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
	dsn := databases.BuildPostgresDSN(req)
	if err := databases.TestConnection(ctx, req.Engine, dsn); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "failed to connect to database", "details": err.Error()})
	}
	port, err := strconv.Atoi(req.Port)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid database port", "details": err.Error()})
	}
	conn := models.DatabaseConnection{
		UserID:       userID,
		Host:         req.Host,
		Engine:       models.Engine(req.Engine),
		Port:         port,
		DatabaseName: req.Database,
		Username:     req.Username,
		Password:     req.Password,
		SSLEnabled:   req.UseSSL,
	}
	if err := s.db.CreateDatabaseConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save connection", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "New Database connected"})
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
