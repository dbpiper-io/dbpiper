package airtable

import (
	"context"
	"database/sql"
	"dbpiper/database"
	"dbpiper/database/models"
	"dbpiper/types"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	metaBasesURL = "https://api.airtable.com/v0/meta/bases"
	authorizeURL = "https://airtable.com/oauth2/v1/authorize"
	tokenURL     = "https://airtable.com/oauth2/v1/token"
	tableBase    = "https://api.airtable.com/v0/meta/bases/%s/tables"
)

type Client interface {
	GetBases(ctx context.Context) ([]types.Base, error)
	OauthConnecter(userID string) (string, error)
	OauthCallback(ctx context.Context, state, code string) (*models.AirtableConnection, error)
	CheckApiKey(ctx context.Context, baseID, apiKey string) error
	GetRedirectURL() string
  SetAirtableConnection(conn *models.AirtableConnection)
}

type Airtable struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	CallbackURI  string
	RedirectURI  string
	DB           *database.DB
	Conn         *models.AirtableConnection
}

func New(db *database.DB, conn *models.AirtableConnection) Client {
	base := os.Getenv("APP_BASE_URL")
	return &Airtable{
		ClientID:     os.Getenv("AIRTABLE_CLIENT_ID"),
		ClientSecret: os.Getenv("AIRTABLE_CLIENT_SECRET"),
		AuthURL:      authorizeURL,
		TokenURL:     tokenURL,
		CallbackURI:  strings.TrimRight(base, "/") + "/api/v1/airtable/oauth/callback",
		RedirectURI:  strings.TrimRight(base, "/") + "/connections",
		DB:           db,
		Conn:         conn,
	}
}

func (a *Airtable) CheckApiKey(ctx context.Context, baseID, apiKey string) error {
	url := fmt.Sprintf(tableBase, baseID)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid API key or base ID")

	}
	return nil
}

func (a *Airtable) GetRedirectURL() string {
	return a.RedirectURI
}

func (a *Airtable) SetAirtableConnection(conn *models.AirtableConnection) {
  a.Conn = conn
}

func (a *Airtable) GetBases(ctx context.Context) ([]types.Base, error) {
	var data struct {
		Bases []types.Base `json:"bases"`
	}
	if err := a.doRequest(ctx, "GET", metaBasesURL, nil, &data); err != nil {
		return nil, err
	}
	return data.Bases, nil
}

func (o *Airtable) OauthConnecter(userID string) (string, error) {
	// PKCE
	verifier, err := generateCodeVerifier()
	if err != nil {
		return "", err
	}

	challenge := codeChallengeFromVerifier(verifier)

	// State (signed)
	state, err := signState(userID, verifier, 5*time.Minute)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", o.ClientID)
	q.Set("redirect_uri", o.CallbackURI)
	q.Set("scope",
		strings.Join([]string{
			"data.records:read",
			"data.records:write",
			"schema.bases:read",
			"schema.bases:write",
			"webhook:manage",
		}, " "),
	)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")

	authURL := o.AuthURL + "?" + q.Encode()

	return authURL, nil
}

func (c *Airtable) OauthCallback(ctx context.Context, state, code string) (*models.AirtableConnection, error) {
	// verify signed state
	payload, err := verifyState(state)
	if err != nil {
		return nil, err
	}

	// Build token request
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.CallbackURI)
	form.Set("client_id", c.ClientID)
	form.Set("code_verifier", payload.CodeVerifier)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.TokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)
	req.Header.Set("User-Agent", "Mozilla/5.0 dbpiper/1.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		AccountID    string `json:"account_id"`
	}

	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	// expiry
	expires := sql.NullTime{}
	if tokenResp.ExpiresIn > 0 {
		expires = sql.NullTime{
			Time:  time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			Valid: true,
		}
	}

	conn := models.AirtableConnection{
		CreatedAt:      time.Now(),
		ConnectionType: models.OAuth,
		UserID:         payload.UserID,
		ProviderAccountID: sql.NullString{
			String: tokenResp.AccountID,
			Valid:  tokenResp.AccountID != "",
		},
		AccessToken: sql.NullString{
			String: tokenResp.AccessToken,
			Valid:  tokenResp.AccessToken != "",
		},
		RefreshToken: sql.NullString{
			String: tokenResp.RefreshToken,
			Valid:  tokenResp.RefreshToken != "",
		},
		Scope: sql.NullString{
			String: tokenResp.Scope,
			Valid:  tokenResp.Scope != "",
		},
		TokenType: sql.NullString{
			String: tokenResp.TokenType,
			Valid:  tokenResp.TokenType != "",
		},
		ExpiresAt: expires,
	}

	return &conn, nil
}
