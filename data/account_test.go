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
	"testing"
)

const (
	uuidAlice = "bf431618-f696-4dca-a95d-882618ce4ef9"
)

func TestListAccounts(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	accounts := ListAccounts()
	if len(accounts) != 2 {
		t.Error("Two accounts expected in list")
	}
}

func TestGetAccount(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	acc, err := GetAccount(uuidAlice)
	if err != nil {
		t.Error(err)
	}
	if acc.Login != "alice" {
		t.Error("Login was expected to be 'alice'")
	}

	_, err = GetAccount("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

func TestAccountPassword(t *testing.T) {
	acc := &Account{}
	acc.SetPassword("foobar")
	if acc.PWHash == "foobar" {
		t.Error("PWHash equals plain text password")
	}
	if len(acc.PWHash) < 60 {
		t.Error("PWHash is too short")
	}
	if !acc.VerifyPassword("foobar") {
		t.Error("Unable to verify password")
	}
}

func TestCreateAccount(t *testing.T) {
	initTestDb(t)

	new := &Account{Login: "theo", Email: "theo@foo.com", FirstName: "Theo", LastName: "Test"}
	new.SetPassword("testtest")
	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, err := GetAccount(new.UUID)
	if err != nil {
		t.Error(err)
	}
	if check.Login != "theo" {
		t.Error("Login was expected to be 'theo'")
	}
}

func TestUpdateAccount(t *testing.T) {
	initTestDb(t)

	newLogin := "alice_in_wonderland"
	newEmail := "alice_in_wonderland@example.com"
	newPw := "secret"
	newTitle := "Dr."
	newFirstName := "I am actually not Alice"
	newMiddleName := "and my last name is"
	newLastName := "Badchild"
	newActivationCode := "1234567890"

	acc, err := GetAccount(uuidAlice)
	if err != nil {
		t.Error(err)
	}

	acc.SetPassword(newPw)
	acc.Login = newLogin
	acc.Email = newEmail
	acc.Title = sql.NullString{String: newTitle, Valid: true}
	acc.FirstName = newFirstName
	acc.MiddleName = sql.NullString{String: newMiddleName, Valid: true}
	acc.LastName = newLastName
	acc.ActivationCode = sql.NullString{String: newActivationCode, Valid: true}

	err = acc.Update()
	if err != nil {
		t.Error(err)
	}

	acc, err = GetAccount(uuidAlice)
	if err != nil {
		t.Error(err)
	}

	if !acc.VerifyPassword(newPw) {
		t.Error("PWHash was not updated")
	}
	if acc.Login == newLogin {
		t.Error("Login was updated although this should never happen")
	}
	if acc.Email != newEmail {
		t.Error("Email was not updated")
	}
	if acc.Title.String != newTitle {
		t.Error("Title was not updated")
	}
	if acc.FirstName != newFirstName {
		t.Error("FirstName was not updated")
	}
	if acc.MiddleName.String != newMiddleName {
		t.Error("MiddleName was not updated")
	}
	if acc.LastName != newLastName {
		t.Error("LastName was not updated")
	}
	if acc.ActivationCode.String != newActivationCode {
		t.Error("ActivationCode was not updated")
	}
}
