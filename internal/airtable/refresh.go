package airtable

import (
	"context"
	"database/sql"
	"dbpiper/database/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type oauthRefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (a *Airtable) refreshToken(ctx context.Context) error {
	if !a.Conn.RefreshToken.Valid {
		return errors.New("missing refresh token")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.Conn.RefreshToken.String)
	// Do NOT include client_id when using Basic Auth with client_secret

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://airtable.com/oauth2/v1/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(a.clientId, a.clientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read body for debugging
	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("airtable refresh error (status %d): %s", resp.StatusCode, string(b))
	}

	var r oauthRefreshResp
	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("failed to parse response: %w, body: %s", err, string(b))
	}

	conn := models.AirtableConnection{
		CreatedAt:      a.Conn.CreatedAt,
		UserID:         a.Conn.UserID,
		BaseID:         a.Conn.BaseID,
		ConnectionType: models.OAuth,
		AccessToken: sql.NullString{
			String: r.AccessToken,
			Valid:  r.AccessToken != "",
		},
		RefreshToken: sql.NullString{
			String: r.RefreshToken,
			Valid:  r.RefreshToken != "",
		},
		ExpiresAt: sql.NullTime{
			Time:  time.Now().Add(time.Duration(r.ExpiresIn) * time.Second),
			Valid: r.ExpiresIn > 0,
		},
	}

	if err := a.DB.UpsertAirtableConnection(ctx, &conn); err != nil {
		return err
	}

	// Update in-memory connection
	a.Conn = &conn

	return nil
}

func (a *Airtable) tokenExpired() bool {
	if !a.Conn.ExpiresAt.Valid {
		return false
	}
	return time.Now().After(a.Conn.ExpiresAt.Time.Add(-20 * time.Second))
}
