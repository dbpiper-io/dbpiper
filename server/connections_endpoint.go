package server

import (
	"dbpiper/internal/airtable"
	"dbpiper/types"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (s *Server) addConnectionEndPoint(g *echo.Group) {
	conn := g.Group("/connections")
	conn.GET("", s.listConnectionsHandler)
}

func (s *Server) listConnectionsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	air, err := s.db.GetAirtableConnection(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusOK, echo.Map{})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	client := airtable.New(&s.db, air)
	bases, err := client.GetBases(ctx)
	if err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{"error": "Failed to get base data from airtable", "details": err.Error()})
	}
	var base types.Base
	for _, b := range bases {
		if b.ID == air.BaseID {
			base = b
		}
	}
	return c.JSON(http.StatusOK, echo.Map{
		"airtable": map[string]any{
			"id":              air.ID,
			"connection_type": air.ConnectionType,
			"created_at":      air.CreatedAt,
			"base": map[string]any{
				"id":   base.ID,
				"name": base.Name,
			},
		},
	})
}
