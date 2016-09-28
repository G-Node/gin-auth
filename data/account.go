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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/util"
	"github.com/G-Node/gin-core/gin"
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
	const q = `SELECT * FROM ActiveAccounts
	           WHERE lower(firstName) LIKE $1 OR lower(middleName) LIKE $1 OR lower(lastName) LIKE $1 OR lower(login) LIKE $1
	           ORDER BY login`

	accounts := make([]Account, 0)
	err := database.Select(&accounts, q, "%"+strings.ToLower(search)+"%")
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

// GetAccountByLogin returns an active account (non disabled, no activation code, no reset password code)
// with matching login.
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

// GetAccountByActivationCode returns an account with matching activation code.
// Returns false if no account with the activation code can be found.
func GetAccountByActivationCode(code string) (*Account, bool) {
	const q = `SELECT * FROM Accounts WHERE activationCode=$1 AND NOT isDisabled`

	account := &Account{}
	err := database.Get(account, q, code)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err == nil
}

// GetAccountByResetPWCode returns an account with matching reset password code.
// Returns false if no account with the reset password code can be found.
func GetAccountByResetPWCode(code string) (*Account, bool) {
	const q = `SELECT * FROM Accounts WHERE resetPWCode=$1 AND NOT isDisabled`

	account := &Account{}
	err := database.Get(account, q, code)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err == nil
}

// GetAccountDisabled returns a disabled account with a matching uuid.
// Returns false if no account with the uuid can be found or if it is not disabled.
func GetAccountDisabled(uuid string) (*Account, bool) {
	const q = `SELECT * FROM Accounts WHERE uuid=$1 AND isDisabled`

	account := &Account{}
	err := database.Get(account, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return account, err == nil
}

// SetPasswordReset updates the password reset code with a new token, if an
// account can be found, that is non disabled and has either email or login of a provided credential.
// Returns false, if no non-disabled account with the credential as email or login can be found.
func SetPasswordReset(credential string) (*Account, bool) {
	const q = `UPDATE Accounts SET resetpwcode=$2
		   WHERE NOT isdisabled AND (login=$1 OR email=$1) RETURNING *`

	code := util.RandomToken()
	account := &Account{}
	err := database.Get(account, q, credential, code)
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

// UpdatePassword hashes a plain text password
// and updates the database entry of the corresponding account.
func (acc *Account) UpdatePassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	const q = `UPDATE Accounts SET pwhash=$1 WHERE uuid=$2 RETURNING *`
	err = database.Get(acc, q, string(hash), acc.UUID)
	if err == nil {
		acc.PWHash = string(hash)
	}
	return err
}

// UpdateEmail checks validity of a new e-mail address and updates the current account
// with a valid new e-mail address.
// The normal account update does not include the e-mail address for safety reasons.
func (acc *Account) UpdateEmail(email string) error {
	if !(len(email) > 2) || !strings.Contains(email, "@") {
		return &util.ValidationError{
			Message:     "Invalid e-mail address",
			FieldErrors: map[string]string{"email": "Please use a valid e-mail address"}}
	}
	if len(email) > 512 {
		return &util.ValidationError{
			Message:     "Invalid e-mail address",
			FieldErrors: map[string]string{"email": "Address too long, please shorten to 512 characters"}}
	}
	exists := &struct {
		Email bool
	}{}

	const check = `SELECT (SELECT COUNT(*) FROM accounts WHERE email = $1) <> 0 AS email`
	err := database.Get(exists, check, email)
	if err != nil {
		panic(err)
	}
	if exists.Email {
		return &util.ValidationError{
			Message:     "E-Mail address already exists",
			FieldErrors: map[string]string{"email": "Please choose a different e-mail address"}}
	}

	const q = `UPDATE Accounts SET email=$1 WHERE uuid=$2 RETURNING *`
	err = database.Get(acc, q, email, acc.UUID)
	if err != nil {
		panic(err)
	}

	acc.Email = email
	return nil
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
// Field ActivationCode is not set via this update function, since this field fulfills a special role.
// It can only be set to a value once by account create and can only be set to null via its own function.
// Fields password and email are not set via this update function, since they require sufficient scope to change.
func (acc *Account) Update() error {
	const q = `UPDATE Accounts
	           SET (isemailpublic, title, firstName, middleName, lastName, institute,
	                department, city, country, isaffiliationpublic, resetPWCode, isDisabled, updatedAt) =
	               ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now())
	           WHERE uuid=$13
	           RETURNING *`

	err := database.Get(acc, q, acc.IsEmailPublic, acc.Title, acc.FirstName, acc.MiddleName,
		acc.LastName, acc.Institute, acc.Department, acc.City, acc.Country, acc.IsAffiliationPublic,
		acc.ResetPWCode, acc.IsDisabled, acc.UUID)

	// TODO There is a lot of room for improvement here concerning errors about constraints for certain fields
	return err
}

// RemoveActivationCode is the only way to remove an ActivationCode from an Account,
// since this field should never be set via the Update function by accident.
func (acc *Account) RemoveActivationCode() error {
	const q = `UPDATE Accounts
	           SET activationcode = NULL
	           WHERE uuid=$1
	           RETURNING *`

	err := database.Get(acc, q, acc.UUID)

	return err
}

// Validate the content of an Account.
// First name, last name, login, email, institute, department, city and country must not be empty;
// Title, first name, middle name last name, login, email, institute, department, city
// and country must not be longer than 521 characters;
// A given login and e-mail address must not exist in the database; An e-mail address must contain an "@".
func (acc *Account) Validate() *util.ValidationError {
	valErr := &util.ValidationError{FieldErrors: make(map[string]string)}

	if acc.Login == "" {
		valErr.FieldErrors["login"] = "Please add login"
	}
	re := regexp.MustCompile("^[a-zA-Z0-9-_]*$")
	if !re.MatchString(acc.Login) {
		valErr.FieldErrors["login"] = "Please use only the following characters: 'a-zA-Z0-9-_'"
	}

	if !(len(acc.Email) > 2) || !strings.Contains(acc.Email, "@") {
		valErr.FieldErrors["email"] = "Please add a valid e-mail address"
	}
	if acc.FirstName == "" {
		valErr.FieldErrors["first_name"] = "Please add first name"
	}
	if acc.LastName == "" {
		valErr.FieldErrors["last_name"] = "Please add last name"
	}
	if acc.Institute == "" {
		valErr.FieldErrors["institute"] = "Please add institute"
	}
	if acc.Department == "" {
		valErr.FieldErrors["department"] = "Please add department"
	}
	if acc.City == "" {
		valErr.FieldErrors["city"] = "Please add city"
	}
	if acc.Country == "" {
		valErr.FieldErrors["country"] = "Please add country"
	}

	const fieldLength = 512
	var lenMessage = fmt.Sprintf("Entry too long, please shorten to %d characters", fieldLength)

	if len(acc.Login) > fieldLength {
		valErr.FieldErrors["login"] = lenMessage
	}
	if len(acc.Email) > fieldLength {
		valErr.FieldErrors["email"] = lenMessage
	}
	if len(acc.Title.String) > fieldLength {
		valErr.FieldErrors["title"] = lenMessage
	}
	if len(acc.FirstName) > fieldLength {
		valErr.FieldErrors["first_name"] = lenMessage
	}
	if len(acc.MiddleName.String) > fieldLength {
		valErr.FieldErrors["middle_name"] = lenMessage
	}
	if len(acc.LastName) > fieldLength {
		valErr.FieldErrors["last_name"] = lenMessage
	}
	if len(acc.Institute) > fieldLength {
		valErr.FieldErrors["institute"] = lenMessage
	}
	if len(acc.Department) > fieldLength {
		valErr.FieldErrors["department"] = lenMessage
	}
	if len(acc.City) > fieldLength {
		valErr.FieldErrors["city"] = lenMessage
	}
	if len(acc.Country) > fieldLength {
		valErr.FieldErrors["country"] = lenMessage
	}

	exists := &struct {
		Login bool
		Email bool
	}{}

	const q = `SELECT
	             (SELECT COUNT(*) FROM accounts WHERE login = $1) <> 0 AS login,
	             (SELECT COUNT(*) FROM accounts WHERE email = $2) <> 0 AS email`

	err := database.Get(exists, q, acc.Login, acc.Email)
	if err != nil {
		panic(err)
	}
	if exists.Login {
		valErr.FieldErrors["login"] = "Please choose a different login"
	}
	if exists.Email {
		valErr.FieldErrors["email"] = "Please choose a different email address"
	}

	if len(valErr.FieldErrors) > 0 {
		valErr.Message = "Registration requirements are not met"
	}

	return valErr
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
	jsonData := &gin.Account{
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
		jsonData.Email = &gin.Email{
			Email:    am.Account.Email,
			IsPublic: am.Account.IsEmailPublic,
		}
	}
	if am.WithAffiliation {
		jsonData.Affiliation = &gin.Affiliation{
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
	jsonData := &gin.Account{}
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
