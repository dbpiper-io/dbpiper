package server

import (
	"context"
	"dbpiper/database/models"
	"dbpiper/internal/databases/pgx"
	"dbpiper/types"
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *Server) addSyncEndPoint(g *echo.Group) {
	sync := g.Group("/sync")
	sync.POST("", s.createSync)
}

func (s *Server) createSync(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not authenticated"})
	}
	var req types.CreateSyncRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid_payload", "details": err.Error()})
	}

	if req.Source.Type == req.Target.Type {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "source_and_target_cannot_be_same"})
	}

	if req.Direction != "one_way" && req.Direction != "two_way" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid_direction"})
	}

	if len(req.Tables) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "no_tables_specified"})
	}

	connID := req.Source.ConnectionID
	if req.Target.Type == models.Pgx {
		connID = req.Target.ConnectionID
	}

	if err := s.validatePgxTableMapping(c, ctx, userID, connID, req.Tables); err != nil {
		return err
	}

	tablesJSON, _ := json.Marshal(req.Tables)

	sync := models.Sync{
		ID:           uuid.New(),
		UserID:       userID,
		SourceType:   req.Source.Type,
		SourceConnID: req.Source.ConnectionID,
		TargetType:   req.Target.Type,
		TargetConnID: req.Target.ConnectionID,
		Direction:    req.Direction, //todo fix this 
		Tables:       tablesJSON,
		Status:       models.SyncSetup,
	}
	if err := s.DB.CreateSync(ctx, &sync); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed_to_create_sync", "details": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"id": sync.ID,
		"source": map[string]string{
			"connection_id": sync.SourceConnID,
			"type":          string(sync.SourceType),
		},
		"target": map[string]string{
			"connection_id": sync.TargetConnID,
			"type":          string(sync.TargetType),
		},
		"direction": sync.Direction,
		"fields":    req.Tables,
	})
}

func (s *Server) validatePgxTableMapping(c echo.Context, ctx context.Context, userID, connID string, tables []types.TableConfig) error {
	db, err := s.DB.GetDatabaseConnectionByID(ctx, userID, connID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid_source_connection", "details": err.Error()})
	}
	dsn := db.ConnectionURL.String
	if !db.ConnectionURL.Valid {
		dsn = pgx.BuildPostgresDSN(db.Username, db.Password, db.Host, strconv.Itoa(db.Port), db.DatabaseName, db.SSLEnabled)
	}
	pgxPool, err := s.PgxPool.GetPool(ctx, connID, dsn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "pool error", "details": err.Error()})
	}

	for _, t := range tables {
		keys := slices.Collect(maps.Keys(t.Fields))
		rows, err := pgxPool.Query(ctx, pgx.SelectQuery(t.SourceTable, keys))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error":   "invalid table",
				"details": err.Error(),
			})
		}
		defer rows.Close()
	}
	return nil
}
