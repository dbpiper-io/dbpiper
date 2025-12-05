package airtable

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	AccountID    string `json:"account_id"`
}

type Base struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Table struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
