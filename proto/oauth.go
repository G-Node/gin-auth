package proto

import "time"

// TokenResponse Token request response json struct
type TokenResponse struct {
	TokenType    string  `json:"token_type"`
	Scope        string  `json:"scope"`
	AccessToken  string  `json:"access_token"`
	RefreshToken *string `json:"refresh_token"`
}

// TokenInfo Token information json struct
type TokenInfo struct {
	URL        string    `json:"url"`
	JTI        string    `json:"jti"`
	EXP        time.Time `json:"exp"`
	ISS        string    `json:"iss"`
	Login      string    `json:"login"`
	AccountURL string    `json:"account_url"`
	Scope      string    `json:"scope"`
}
