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
	"github.com/G-Node/gin-auth/util"
	"testing"
)

const (
	uuidAlice = "bf431618-f696-4dca-a95d-882618ce4ef9"
	uuidBob   = "51f5ac36-d332-4889-8023-6e033fcd8e17"
)

func TestListAccounts(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	accounts := ListAccounts()
	if len(accounts) != 2 {
		t.Error("Two accounts expected in list")
	}
}

func TestGetAccount(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.Login != "alice" {
		t.Error("Login was expected to be 'alice'")
	}

	_, ok = GetAccount("doesNotExist")
	if ok {
		t.Error("Account should not exist")
	}
}

func TestGetAccountByLogin(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	acc, ok := GetAccountByLogin("bob")
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.UUID != uuidBob {
		t.Errorf("UUID was expected to be '%s'", uuidBob)
	}

	_, ok = GetAccount("doesNotExist")
	if ok {
		t.Error("Account should not exist")
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
	InitTestDb(t)

	fresh := &Account{Login: "theo", Email: "theo@foo.com", FirstName: "Theo", LastName: "Test"}
	fresh.SetPassword("testtest")
	err := fresh.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetAccount(fresh.UUID)
	if !ok {
		t.Error("Account does not exist")
	}
	if check.Login != "theo" {
		t.Error("Login was expected to be 'theo'")
	}
}

func TestUpdateAccount(t *testing.T) {
	InitTestDb(t)

	newLogin := "alice_in_wonderland"
	newEmail := "alice_in_wonderland@example.com"
	newPw := "secret"
	newTitle := "Dr."
	newFirstName := "I am actually not Alice"
	newMiddleName := "and my last name is"
	newLastName := "Badchild"
	newActivationCode := "1234567890"

	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}

	acc.SetPassword(newPw)
	acc.Login = newLogin
	acc.Email = newEmail
	acc.Title = sql.NullString{String: newTitle, Valid: true}
	acc.FirstName = newFirstName
	acc.MiddleName = sql.NullString{String: newMiddleName, Valid: true}
	acc.LastName = newLastName
	acc.ActivationCode = sql.NullString{String: newActivationCode, Valid: true}

	err := acc.Update()
	if err != nil {
		t.Error(err)
	}

	acc, ok = GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
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
