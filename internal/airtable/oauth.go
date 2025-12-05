package airtable

import (
	"database/sql"
	"dbpiper/database"
	"dbpiper/database/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type OAuthService struct {
	DB           database.DB
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	CallbackURI  string
	RedirectURI  string
}

func New(db database.DB) *OAuthService {
	base := os.Getenv("APP_BASE_URL")
	return &OAuthService{
		DB:           db,
		ClientID:     os.Getenv("AIRTABLE_CLIENT_ID"),
		ClientSecret: os.Getenv("AIRTABLE_CLIENT_SECRET"),
		AuthURL:      os.Getenv("AIRTABLE_AUTHORIZE_URL"),
		TokenURL:     os.Getenv("AIRTABLE_TOKEN_URL"),
		CallbackURI:  strings.TrimRight(base, "/") + "/api/v1/airtable/oauth/callback",
		RedirectURI:  strings.TrimRight(base, "/") + "/connections",
	}
}

// ConnectHandler: GET /v1/airtable/oauth/connect
func (o *OAuthService) ConnectHandler(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "not_authenticated"})
	}

	// PKCE
	verifier, err := generateCodeVerifier()
	if err != nil {
		return c.JSON(500, echo.Map{"error": "pkce_error", "details": err.Error()})
	}

	challenge := codeChallengeFromVerifier(verifier)
	fmt.Println(o.CallbackURI)

	// State (signed)
	state, err := signState(userID, verifier, 5*time.Minute)
	if err != nil {
		return c.JSON(500, echo.Map{"error": "state_error", "details": err.Error()})
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

	return c.JSON(200, echo.Map{
		"url": authURL,
	})
}

// CallbackHandler:
func (o *OAuthService) CallbackHandler(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return c.JSON(400, echo.Map{"error": "missing_code_or_state"})
	}

	// verify signed state
	payload, err := verifyState(state)
	if err != nil {
		return c.JSON(400, echo.Map{"error": "invalid_state", "details": err.Error()})
	}

	userID := payload.UserID
	codeVerifier := payload.CodeVerifier

	// Build token request
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", o.CallbackURI)
	form.Set("client_id", o.ClientID)
	form.Set("code_verifier", codeVerifier)

	req, _ := http.NewRequestWithContext(ctx, "POST", o.TokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.ClientID, o.ClientSecret)
	req.Header.Set("User-Agent", "Mozilla/5.0 dbpiper/1.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(502, echo.Map{"error": "token_request_failed", "details": err.Error()})
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
		return c.JSON(502, echo.Map{"error": "invalid_token_response", "details": err.Error()})
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
		UserID:         userID,
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

	client := NewClient(&conn, o.DB)
	bases, err := client.GetBases(ctx)
	if err != nil {
		return c.JSON(http.StatusMethodNotAllowed, echo.Map{"error": "airtable_error", "details": err.Error()})
	}
	if len(bases) == 0 || bases[0].ID == "" {
		return c.JSON(http.StatusMethodNotAllowed, echo.Map{"error": "airtable_error", "details": "not data allowed"})
	}
	conn.BaseID = bases[0].ID

	if err := o.DB.UpsertAirtableConnection(ctx, &conn); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db_error", "details": err.Error()})
	}

	return c.Redirect(http.StatusSeeOther, o.RedirectURI)
}
