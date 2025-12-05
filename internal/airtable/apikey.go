package airtable

import (
	"database/sql"
	"dbpiper/database/models"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type APIKeyConnectRequest struct {
	APIKey string `json:"api_key"`
	BaseID string `json:"base_id"`
}

type AirtableTableResponse struct {
	Tables []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tables"`
}

func (o *OAuthService) APIKeyConnect(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	var req APIKeyConnectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body",
		})
	}

	if req.APIKey == "" || req.BaseID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "api_key and base_id are required",
		})
	}
	url := fmt.Sprintf("https://api.airtable.com/v0/meta/bases/%s/tables", req.BaseID)
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Failed to connect to Airtable",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "Invalid API key or base ID",
		})
	}
	conn := models.AirtableConnection{
		CreatedAt:      time.Now(),
		UserID:         userID,
		ConnectionType: models.APIKey,
		APIKey:         sql.NullString{String: req.APIKey, Valid: req.APIKey != ""},
		BaseID:         req.BaseID,
	}

	if err := o.DB.UpsertAirtableConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "connected",
	})
}
