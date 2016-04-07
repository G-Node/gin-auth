package data

import (
	"database/sql"
	"github.com/G-Node/gin-auth/util"
	"time"
)

// RefreshToken represents an OAuth refresh token issued
// in a `code` grant request.
type RefreshToken struct {
	Token           string
	Scope           SqlStringSlice
	OAuthClientUUID string
	AccountUUID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ListRefreshTokens returns all refresh tokens sorted by creation time.
func ListRefreshTokens() []RefreshToken {
	const q = `SELECT * FROM RefreshTokens ORDER BY createdAt`

	refreshTokens := make([]RefreshToken, 0)
	err := database.Select(&refreshTokens, q)
	if err != nil {
		panic(err)
	}

	return refreshTokens
}

// GetRefreshToken returns a refresh token with a given token value.
// Returns false if no such refresh token exists.
func GetRefreshToken(token string) (*RefreshToken, bool) {
	const q = `SELECT * FROM RefreshTokens WHERE token=$1`

	refreshToken := &RefreshToken{}
	err := database.Get(refreshToken, q, token)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return refreshToken, err == nil
}

// Create stores a new refresh token in the database.
// If the token is empty a random token will be generated.
func (tok *RefreshToken) Create() error {
	const q = `INSERT INTO RefreshTokens (token, scope, oAuthClientUUID, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, now(), now())
	           RETURNING *`

	if tok.Token == "" {
		tok.Token = util.RandomToken()
	}

	return database.Get(tok, q, tok.Token, tok.Scope, tok.OAuthClientUUID, tok.AccountUUID)
}

// Delete removes an refresh token from the database.
func (tok *RefreshToken) Delete() error {
	const q = `DELETE FROM RefreshTokens WHERE token=$1`

	_, err := database.Exec(q, tok.Token)
	return err
}
