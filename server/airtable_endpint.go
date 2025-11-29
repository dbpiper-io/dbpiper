package server

import (
	"dbpiper/internal/airtable"

	"github.com/labstack/echo/v4"
)

func (s *Server) addAirtableEndPoint(g *echo.Group) {
	airtableHandle := airtable.New(s.db)
	air := g.Group("/airtable")
	oauth := air.Group("/oauth")
	oauth.GET("/connect", airtableHandle.ConnectHandler)
	oauth.GET("/callback", airtableHandle.CallbackHandler)
}
