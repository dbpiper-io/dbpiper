package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserClaims struct {
	ID    string
	Email string
	Name  string
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	e.Use(JWTMiddleware)

	e.GET("/api/me", func(c echo.Context) error {
		user := c.Get("user").(*UserClaims)

		return c.JSON(200, echo.Map{
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
		})
	})

	return e
}
