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
	UUID                string
	Login               string
	PWHash              string `json:"-"` // safety net
	Email               string
	IsEmailPublic       bool
	Title               sql.NullString
	FirstName           string
	MiddleName          sql.NullString
	LastName            string
	Institute           string
	Department          string
	City                string
	Country             string
	IsAffiliationPublic bool
	ActivationCode      sql.NullString
	ResetPWCode         sql.NullString
	IsDisabled          bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ListAccounts returns all accounts stored in the database
func ListAccounts() []Account {
	const q = `SELECT * FROM ActiveAccounts ORDER BY login`

	accounts := make([]Account, 0)
	err := database.Select(&accounts, q)
	if err != nil {
		panic(err)
	}

	return accounts
}

// SearchAccounts returns all accounts stored in the database where the account name (firstName, middleName, lastName
// or login) contains the search string.
func SearchAccounts(search string) []Account {
	const q = `SELECT * FROM ActiveAccounts a
	           WHERE a.firstName LIKE $1 OR a.middleName LIKE $1 OR a.lastName LIKE $1 OR a.login LIKE $1
	           ORDER BY login`

	accounts := make([]Account, 0)
	err := database.Select(&accounts, q, "%"+search+"%")
	if err != nil {
		panic(err)
	}

	return accounts
}

// GetAccount returns an account with matching UUID
// Returns false if no account with such UUID exists
func GetAccount(uuid string) (*Account, bool) {
	const q = `SELECT * FROM ActiveAccounts a WHERE a.uuid=$1`

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
	const q = `SELECT * FROM ActiveAccounts a WHERE a.login=$1`

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
	const q = `INSERT INTO Accounts (uuid, login, pwHash, email, isEmailPublic, title, firstName, middleName, lastName,
	                                 institute, department, city, country, isAffiliationPublic, activationCode,
	                                 createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, now(), now())
	           RETURNING *`

	if acc.UUID == "" {
		acc.UUID = uuid.NewRandom().String()
	}

	err := database.Get(acc, q, acc.UUID, acc.Login, acc.PWHash, acc.Email, acc.IsEmailPublic, acc.Title, acc.FirstName,
		acc.MiddleName, acc.LastName, acc.Institute, acc.Department, acc.City, acc.Country, acc.IsAffiliationPublic,
		acc.ActivationCode)

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
	           SET (pwHash, email, isemailpublic, title, firstName, middleName, lastName, institute, department, city,
	                country, isaffiliationpublic, activationCode, resetPWCode, isDisabled, updatedAt) =
	               ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, now())
	           WHERE uuid=$16
	           RETURNING *`

	err := database.Get(acc, q, acc.PWHash, acc.Email, acc.IsEmailPublic, acc.Title, acc.FirstName, acc.MiddleName,
		acc.LastName, acc.Institute, acc.Department, acc.City, acc.Country, acc.IsAffiliationPublic,
		acc.ActivationCode, acc.ResetPWCode, acc.IsDisabled, acc.UUID)

	// TODO There is a lot of room for improvement here concerning errors about constraints for certain fields
	return err
}

type jsonAccount struct {
	URL         string           `json:"url"`
	UUID        string           `json:"uuid"`
	Login       string           `json:"login"`
	Email       *jsonEmail       `json:"email"`
	Title       *string          `json:"title"`
	FirstName   string           `json:"first_name"`
	MiddleName  *string          `json:"middle_name"`
	LastName    string           `json:"last_name"`
	Affiliation *jsonAffiliation `json:"affiliation"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type jsonEmail struct {
	Email    string `json:"email"`
	IsPublic bool   `json:"is_public"`
}

type jsonAffiliation struct {
	Institute  string `json:"institute"`
	Department string `json:"department"`
	City       string `json:"city"`
	Country    string `json:"country"`
	IsPublic   bool   `json:"is_public"`
}

// AccountMarshaler handles JSON marshalling for Account
//
// Fields:
// - WithMail        If true, mail information will be serialized
// - WithAffiliation If true, affiliation will be serialized
type AccountMarshaler struct {
	WithMail        bool
	WithAffiliation bool
	Account         *Account
}

// MarshalJSON implements Marshaler for AccountMarshaler
func (am *AccountMarshaler) MarshalJSON() ([]byte, error) {
	jsonData := &jsonAccount{
		URL:       conf.MakeUrl("/api/accounts/%s", am.Account.Login),
		UUID:      am.Account.UUID,
		Login:     am.Account.Login,
		FirstName: am.Account.FirstName,
		LastName:  am.Account.LastName,
		CreatedAt: am.Account.CreatedAt,
		UpdatedAt: am.Account.UpdatedAt,
	}
	if am.Account.Title.Valid {
		jsonData.Title = &am.Account.Title.String
	}
	if am.Account.MiddleName.Valid {
		jsonData.MiddleName = &am.Account.MiddleName.String
	}
	if am.WithMail {
		jsonData.Email = &jsonEmail{
			Email:    am.Account.Email,
			IsPublic: am.Account.IsEmailPublic,
		}
	}
	if am.WithAffiliation {
		jsonData.Affiliation = &jsonAffiliation{
			Institute:  am.Account.Institute,
			Department: am.Account.Department,
			City:       am.Account.City,
			Country:    am.Account.Country,
			IsPublic:   am.Account.IsAffiliationPublic,
		}
	}
	return json.Marshal(jsonData)
}

// UnmarshalJSON implements Unmarshaler for AccountMarshaler.
// Only parses updatable fields: Title, FirstName, MiddleName and LastName
func (am *AccountMarshaler) UnmarshalJSON(bytes []byte) error {
	jsonData := &jsonAccount{}
	err := json.Unmarshal(bytes, jsonData)
	if err != nil {
		return err
	}

	if am.Account == nil {
		am.Account = &Account{}
	}

	am.Account.Login = jsonData.Login
	if jsonData.Title != nil {
		am.Account.Title = sql.NullString{String: *jsonData.Title, Valid: true}
	} else {
		am.Account.Title = sql.NullString{}
	}
	am.Account.FirstName = jsonData.FirstName
	if jsonData.MiddleName != nil {
		am.Account.MiddleName = sql.NullString{String: *jsonData.MiddleName, Valid: true}
	} else {
		am.Account.MiddleName = sql.NullString{}
	}
	am.Account.LastName = jsonData.LastName

	if jsonData.Email != nil {
		am.Account.Email = jsonData.Email.Email
		am.Account.IsEmailPublic = jsonData.Email.IsPublic
	}

	if jsonData.Affiliation != nil {
		am.Account.Institute = jsonData.Affiliation.Institute
		am.Account.Department = jsonData.Affiliation.Department
		am.Account.City = jsonData.Affiliation.City
		am.Account.Country = jsonData.Affiliation.Country
		am.Account.IsAffiliationPublic = jsonData.Affiliation.IsPublic
	}

	return nil
}
