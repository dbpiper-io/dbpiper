package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var (
	jwksCache jwk.Set
	jwksOnce  sync.Once
	jwksErr   error
)

// load JWKS once (Better Auth endpoint)
func loadJWKS() (jwk.Set, error) {
	jwksOnce.Do(func() {
		jwksCache, jwksErr = jwk.Fetch(
			context.Background(), 
			"http://localhost:3000/api/auth/jwks",
		)
	})
	return jwksCache, jwksErr
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		keyset, err := loadJWKS()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error":   "failed to load jwks",
			})
		}

		// Parse JWT automatically from Authorization: Bearer <token>
		token, err := jwt.ParseRequest(
			c.Request(),
			jwt.WithKeySet(keyset),
		)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error":   "Invalid token",
			})
		}

		// Extract claims
		sub, ok := token.Subject()
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "missing subject claim",
			})
		}

		var email, name string
		_ = token.Get("email", &email)
		_ = token.Get("name", &name)

		// Save in Echo context
    c.Set("user_id", sub)

		return next(c)
	}
}
