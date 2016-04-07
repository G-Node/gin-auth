package data

import (
	"database/sql"
	"github.com/G-Node/gin-auth/util"
	"time"
)

const (
	defaultTokenLifeTime = time.Hour * 24
)

type AccessToken struct {
	Token           string
	Scope           SqlStringSlice
	Expires         time.Time
	OAuthClientUUID string
	AccountUUID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ListAccessTokens returns all access tokens sorted by creation time.
func ListAccessTokens() []AccessToken {
	const q = `SELECT * FROM AccessTokens ORDER BY createdAt`

	accessTokens := make([]AccessToken, 0)
	err := database.Select(&accessTokens, q)
	if err != nil {
		panic(err)
	}

	return accessTokens
}

// GetAccessToken returns a access token with a given token.
// Returns false if no such access token exists.
func GetAccessToken(token string) (*AccessToken, bool) {
	const q = `SELECT * FROM AccessTokens WHERE token=$1`

	accessToken := &AccessToken{}
	err := database.Get(accessToken, q, token)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return accessToken, err == nil
}

// ClearOldAccessTokens removes all expired access tokens from the database
// and returns the number of removed access tokens.
func ClearOldAccessTokens() int64 {
	const q = `DELETE FROM AccessTokens WHERE expires < now()`

	res, err := database.Exec(q)
	if err != nil {
		panic(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	return rows
}

// Create stores a new access token in the database.
// If the token is empty a random token will be generated.
func (tok *AccessToken) Create() error {
	const q = `INSERT INTO AccessTokens (token, scope, expires, oAuthClientUUID, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, now(), now())
	           RETURNING *`

	if tok.Token == "" {
		tok.Token = util.RandomToken()
	}

	return database.Get(tok, q, tok.Token, tok.Scope, tok.Expires, tok.OAuthClientUUID, tok.AccountUUID)
}

// UpdateExpirationTime updates the expiration time and stores
// the new time in the database.
func (tok *AccessToken) UpdateExpirationTime() error {
	const q = `UPDATE AccessTokens SET (expires, updatedAt) = ($1, now())
	           WHERE token=$2
	           RETURNING *`

	return database.Get(tok, q, time.Now().Add(defaultTokenLifeTime), tok.Token)
}

// Delete removes an access token from the database.
func (tok *AccessToken) Delete() error {
	const q = `DELETE FROM AccessTokens WHERE token=$1`

	_, err := database.Exec(q, tok.Token)
	return err
}
