package airtable

import (
	"bytes"
	"context"
	"dbpiper/database/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (a *Airtable) doRequest(ctx context.Context, method, url string, body []byte, response any) error {
	client := &http.Client{}
	if a.DB == nil {
		return fmt.Errorf("database required for this call")
	}
	if a.Conn == nil {
		return fmt.Errorf("airtable required for this call")
	}

	if a.Conn.ConnectionType == models.OAuth && a.tokenExpired() {
		if err := a.refreshToken(ctx); err != nil {
			return err
		}
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}

	var access string
	switch a.Conn.ConnectionType {
	case models.OAuth:
		access = a.Conn.AccessToken.String
	case models.APIKey:
		access = a.Conn.APIKey.String
	}
	req.Header.Add("Authorization", "Bearer "+access)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return json.NewDecoder(res.Body).Decode(&response)
}
