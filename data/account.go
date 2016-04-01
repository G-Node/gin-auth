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
	"time"

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
	PWHash         string
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

// SetPassword hashes the plain text password in
func (acc *Account) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err == nil {
		acc.PWHash = string(hash)
	}
	return err
}

func (acc *Account) VerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(acc.PWHash), []byte(plain))
	return err == nil
}

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
