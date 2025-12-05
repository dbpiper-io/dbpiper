package server

import (
	"dbpiper/internal/airtable"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) addAirtableEndPoint(g *echo.Group) {
	airtableHandle := airtable.New(s.db)
	air := g.Group("/airtable")

	oauth := air.Group("/oauth")
	oauth.GET("/connect", airtableHandle.ConnectHandler)
	oauth.GET("/callback", airtableHandle.CallbackHandler)

	apikey := air.Group("/apikey")
	apikey.POST("/connect", airtableHandle.APIKeyConnect)

	air.DELETE("/:id", s.deleteConnectionsHandler)
}

func (s *Server) deleteConnectionsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}
	id := c.Param("id")
	if err := s.db.DeleteAirtableConnection(ctx, userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success":             true,
		"message":             "Integration removed from our system. For complete removal, please also revoke access in your Airtable account: Avatar → Integrations → Third-party integrations → [Your App Name] → Revoke access",
		"airtable_revoke_url": "https://airtable.com/account/integrations",
	})
}
