package data

import (
	"database/sql"
	"time"
)

// Account data as stored in the database
type Account struct {
	UUID           string
	Login          string
	Email          string
	Title          sql.NullString
	FirstName      string
	MiddleName     sql.NullString
	LastName       string
	Password       string
	ActivationCode sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ListAccounts returns all accounts stored in the database
func ListAccounts() []Account {
	const q = `SELECT * FROM Accounts ORDER BY login`

	accounts := make([]Account, 0)
	err := database.Select(&accounts, q)
	if err != nil {
		panic(err)
	}

	return accounts
}

// GetAccount returns an account with matching UUID
// Returns an error if no account with such UUID exists
func GetAccount(uuid string) (*Account, error) {
	const q = `SELECT * FROM Accounts a WHERE a.uuid=$1`

	account := &Account{}
	err := database.Get(account, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err
}
