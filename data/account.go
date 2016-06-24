// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/bcrypt"
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
	PWHash         string `json:"-"` // safety net
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
// Returns false if no account with such UUID exists
func GetAccount(uuid string) (*Account, bool) {
	const q = `SELECT * FROM Accounts a WHERE a.uuid=$1`

	account := &Account{}
	err := database.Get(account, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err == nil
}

// GetAccountByLogin returns an account with matching login.
// Returns false if no account with such login exists.
func GetAccountByLogin(login string) (*Account, bool) {
	const q = `SELECT * FROM Accounts a WHERE a.login=$1`

	account := &Account{}
	err := database.Get(account, q, login)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err == nil
}

// SetPassword hashes the plain text password and
// sets PWHash to the new value.
func (acc *Account) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err == nil {
		acc.PWHash = string(hash)
	}
	return err
}

// VerifyPassword checks whether the stored hash matches the plain text password
func (acc *Account) VerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(acc.PWHash), []byte(plain))
	return err == nil
}

// Create stores the account as new Account in the database.
// If the UUID string is empty a new UUID will be generated.
func (acc *Account) Create() error {
	const q = `INSERT INTO Accounts (uuid, login, email, title, firstName, middleName, lastName, pwHash,
	                                 activationCode, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
	           RETURNING *`

	if acc.UUID == "" {
		acc.UUID = uuid.NewRandom().String()
	}

	err := database.Get(acc, q, acc.UUID, acc.Login, acc.Email, acc.Title, acc.FirstName, acc.MiddleName, acc.LastName,
		acc.PWHash, acc.ActivationCode)

	// TODO There is a lot of room for improvement here concerning errors about constraints for certain fields
	return err
}

// SSHKeys returns a slice with all ssh key belonging to this account.
func (acc *Account) SSHKeys() []SSHKey {
	const q = `SELECT * FROM SSHKeys WHERE accountUUID = $1 ORDER BY fingerprint`

	keys := make([]SSHKey, 0)
	err := database.Select(&keys, q, acc.UUID)
	if err != nil {
		panic(err)
	}

	return keys
}

// Update stores the new values of an Account in the database.
// New values for Login and CreatedAt are ignored. UpdatedAt will be set
// automatically to the current date and time.
func (acc *Account) Update() error {
	const q = `UPDATE Accounts
	           SET (email, title, firstName, middleName, lastName, pwHash, activationCode, updatedAt) =
	               ($1, $2, $3, $4, $5, $6, $7, now())
	           WHERE uuid=$8
	           RETURNING *`

	err := database.Get(acc, q, acc.Email, acc.Title, acc.FirstName, acc.MiddleName, acc.LastName, acc.PWHash,
		acc.ActivationCode, acc.UUID)

	// TODO There is a lot of room for improvement here concerning errors about constraints for certain fields
	return err
}

type jsonAccount struct {
	URL        string    `json:"url"`
	UUID       string    `json:"uuid"`
	Login      string    `json:"login"`
	Email      string    `json:"email,omitempty"`
	Title      *string   `json:"title"`
	FirstName  string    `json:"first_name"`
	MiddleName *string   `json:"middle_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MarshalJSON implements Marshaler for Account
func (acc *Account) MarshalJSON() ([]byte, error) {
	jsonData := &jsonAccount{
		URL:       conf.MakeUrl("/api/accounts/%s", acc.Login),
		UUID:      acc.UUID,
		Login:     acc.Login,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
	}
	if acc.Title.Valid {
		jsonData.Title = &acc.Title.String
	}
	if acc.MiddleName.Valid {
		jsonData.MiddleName = &acc.MiddleName.String
	}
	return json.Marshal(jsonData)
}

// UnmarshalJSON implements Unmarshaler for Account.
// Only parses updatable fields: Title, FirstName, MiddleName and LastName
func (acc *Account) UnmarshalJSON(bytes []byte) error {
	jsonData := &jsonAccount{}
	err := json.Unmarshal(bytes, jsonData)
	if err != nil {
		return err
	}

	if jsonData.Title != nil {
		acc.Title = sql.NullString{String: *jsonData.Title, Valid: true}
	} else {
		acc.Title = sql.NullString{}
	}
	acc.FirstName = jsonData.FirstName
	if jsonData.MiddleName != nil {
		acc.MiddleName = sql.NullString{String: *jsonData.MiddleName, Valid: true}
	} else {
		acc.MiddleName = sql.NullString{}
	}
	acc.LastName = jsonData.LastName

	return nil
}
