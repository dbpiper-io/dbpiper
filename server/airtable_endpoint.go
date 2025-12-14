package server

import (
	"database/sql"
	"dbpiper/database/models"
	"dbpiper/internal/airtable"
	"dbpiper/types"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) addAirtableEndPoint(g *echo.Group) {
	airtables := g.Group("/airtable")

	air := airtables.Group("/:id")
	air.DELETE("", s.deleteAirtableConnectionHandler)
	air.GET("/tables", s.getAirtableTables)

	oauth := airtables.Group("/oauth")
	oauth.GET("/connect", s.connectHandler)
	oauth.GET("/callback", s.callbackHandler)

	apikey := airtables.Group("/apikey")
	apikey.POST("/connect", s.apiKeyConnecter)
}

func (s *Server) connectHandler(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not authenticated"})
	}
	air := airtable.New(nil, nil)
	authURL, err := air.OauthConnecter(userID)
	if err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{"error": "connected to airtable failed", "details": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"url": authURL,
	})
}

func (s *Server) callbackHandler(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing_code_or_state"})
	}

	air := airtable.New(&s.DB, nil)
	conn, err := air.OauthCallback(ctx, state, code)
	if err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{"error": "airtable callback failed", "details": err.Error()})
	}

	air.SetAirtableConnection(conn)
	bases, err := air.GetBases(ctx)
	if err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{"error": "airtable callback failed", "details": err.Error()})
	}

	if len(bases) == 0 || bases[0].ID == "" {
		return c.JSON(http.StatusBadGateway, echo.Map{"error": "airtable callback failed", "details": "no base allowed to access"})
	}
	conn.BaseID = bases[0].ID

	if err := s.DB.UpsertAirtableConnection(ctx, conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.Redirect(http.StatusFound, air.GetRedirectURL())
}

func (s *Server) apiKeyConnecter(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	var req types.APIKeyConnectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body",
      "details": err.Error(),
		})
	}

	if req.APIKey == "" || req.BaseID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "api_key and base_id are required",
		})
	}
	air := airtable.New(nil, nil)
	if err := air.CheckApiKey(ctx, req.BaseID, req.APIKey); err != nil {
		return c.JSON(http.StatusBadGateway, echo.Map{
			"error": "Invalid API key or base ID",
      "details": err.Error(),
		})
	}
	conn := models.AirtableConnection{
		CreatedAt:      time.Now(),
		UserID:         userID,
		ConnectionType: models.APIKey,
		APIKey:         sql.NullString{String: req.APIKey, Valid: req.APIKey != ""},
		BaseID:         req.BaseID,
	}

	if err := s.DB.UpsertAirtableConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "connected",
	})
}

func (s *Server) deleteAirtableConnectionHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}
	id := c.Param("id")
	if err := s.DB.DeleteAirtableConnection(ctx, userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Integration removed from our system. For complete removal, please also revoke access in your Airtable account",
	})
}

func (s *Server) getAirtableTables(c echo.Context) error {
	ctx := c.Request().Context()
	connID := c.Param("id")

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	air, err := s.DB.GetAirtableConnectionByID(ctx, userID, connID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "connection not found", "details": err.Error()})
	}
	client := airtable.New(&s.DB, air)
	tables, err := client.GetTables(ctx)
	if err != nil {
		return c.JSON(http.StatusMethodNotAllowed, echo.Map{"error": "airtable error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"id":     connID,
		"driver": "airtable",
		"tables": tables,
	})
}
