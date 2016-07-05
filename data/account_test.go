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
	if len(accounts) != 3 {
		t.Error("Three accounts expected in list")
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

	// Test whole barrage of inactive accounts
	inactiveUUID := []string{"test0001", "test0002", "test0003", "test0004", "test0005", "test0006"}
	suffix := "-1234-6789-1234-678901234567"
	for _, v := range inactiveUUID {
		currUUID := v + suffix
		_, ok = GetAccount(currUUID)
		if ok {
			t.Errorf("Account with login '%s' should not exist", currUUID)
		}
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

	_, ok = GetAccountByLogin("doesNotExist")
	if ok {
		t.Error("Account should not exist")
	}

	// Test whole barrage of inactive accounts
	inactiveLogin := []string{"inact_log1", "inact_log2", "inact_log3", "inact_log4", "inact_log5", "inact_log6"}
	for _, v := range inactiveLogin {
		_, ok = GetAccountByLogin(v)
		if ok {
			t.Errorf("Account with login '%s' should not exist", v)
		}
	}
}

func TestGetAccountByActivationCode(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const enabledUUID = "test0001-1234-6789-1234-678901234567"
	const enabledCode = "ac_a"
	const disabledCode = "ac_b"

	acc, ok := GetAccountByActivationCode(enabledCode)
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.UUID != enabledUUID {
		t.Errorf("UUID was expected to be '%s'", enabledUUID)
	}

	_, ok = GetAccountByActivationCode(disabledCode)
	if ok {
		t.Error("Account should not exist")
	}

	_, ok = GetAccountByActivationCode("")
	if ok {
		t.Error("Account should not exist")
	}

	_, ok = GetAccountByActivationCode("iDoNotExist")
	if ok {
		t.Error("Account should not exist")
	}
}

func TestGetAccountByResetPWCode(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const enabledUUID = "test0002-1234-6789-1234-678901234567"
	const enabledCode = "rc_a"
	const disabledCode = "rc_c"

	acc, ok := GetAccountByResetPWCode(enabledCode)
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.UUID != enabledUUID {
		t.Errorf("UUID was expected to be '%s'", enabledUUID)
	}

	_, ok = GetAccountByResetPWCode(disabledCode)
	if ok {
		t.Error("Account should not exist")
	}

	_, ok = GetAccountByResetPWCode("")
	if ok {
		t.Error("Account should not exist")
	}

	_, ok = GetAccountByResetPWCode("iDoNotExist")
	if ok {
		t.Error("Account should not exist")
	}
}

func TestGetAccountDisabled(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const disabledUUID = "test0004-1234-6789-1234-678901234567"
	const enabledUUID = "test0001-1234-6789-1234-678901234567"

	acc, ok := GetAccountDisabled(disabledUUID)
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.UUID != disabledUUID {
		t.Errorf("UUID was expected to be '%s'", disabledUUID)
	}

	_, ok = GetAccountDisabled(enabledUUID)
	if ok {
		t.Error("Account should not exist")
	}
}

func TestAccount_SetPassword(t *testing.T) {
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

func TestAccount_Create(t *testing.T) {
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

func TestAccount_SSHKeys(t *testing.T) {
	InitTestDb(t)
	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}

	keys := acc.SSHKeys()
	if len(keys) != 1 {
		t.Error("List should contain one single key")
	}
	key := keys[0]
	if key.AccountUUID != acc.UUID {
		t.Errorf("Account uuid expected to be '%s' but was '%s'", acc.UUID, key.AccountUUID)
	}
}

func TestAccount_Update(t *testing.T) {
	InitTestDb(t)

	newLogin := "alice_in_wonderland"
	newEmail := "alice_in_wonderland@example.com"
	newPw := "secret"
	newTitle := "Dr."
	newFirstName := "I am actually not Alice"
	newMiddleName := "and my last name is"
	newLastName := "Badchild"
	newActivationCode := "activation code"
	newResetPWCode := "reset password code"

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

	acc.ActivationCode = sql.NullString{String: newActivationCode, Valid: true}
	err = acc.Update()
	if err != nil {
		t.Error(err)
	}
	acc, ok = GetAccountByActivationCode(newActivationCode)
	if !ok {
		t.Error("Activation code update failed")
	}
	if acc.ActivationCode.String != newActivationCode {
		t.Error("Activation code was not updated")
	}

	acc.ResetPWCode = sql.NullString{String: newResetPWCode, Valid: true}
	err = acc.Update()
	if err != nil {
		t.Error(err)
	}
	acc, ok = GetAccountByResetPWCode(newResetPWCode)
	if !ok {
		t.Error("Password reset code update failed")
	}
	if acc.ResetPWCode.String != newResetPWCode {
		t.Error("Reset password code was not updated")
	}

	acc.IsDisabled = true
	err = acc.Update()
	if err != nil {
		t.Error(err)
	}
	acc, ok = GetAccountDisabled(uuidAlice)
	if !ok {
		t.Error("Disable account update failed")
	}
	if !acc.IsDisabled {
		t.Error("Account isDisabled was not updated")
	}
}

func TestValidate(t *testing.T) {
	InitTestDb(t)

	const errDiffUser = "Please choose a different username"

	account := &Account{}

	// Test all data missing
	valErr := account.Validate()
	if valErr.Message == "" {
		t.Error("Expected validation error")
	}

	account.FirstName = "fn"
	account.LastName = "ln"
	account.Login = "alice"
	account.Email = "bob@foo.com"
	account.Institute = "Inst"
	account.Department = "Dep"
	account.City = "cty"
	account.Country = "ctry"

	// Test existing login
	valErr = account.Validate()
	if valErr.FieldErrors["login"] != errDiffUser {
		t.Errorf("Expected invalid username error, but got: '%s'", valErr.FieldErrors["login"])
	}

	// Test missing email
	account.Login = "noone"
	account.Email = ""
	valErr = account.Validate()
	if valErr.FieldErrors["email"] != "Please add email" {
		t.Errorf("Expected missing email error, but got: '%s'", valErr.FieldErrors["email"])
	}

	// Test missing first name
	account.Email = "noone@nowhere.com"
	account.FirstName = ""
	valErr = account.Validate()
	if valErr.FieldErrors["first_name"] != "Please add first name" {
		t.Errorf("Expected missing first name error, but got: '%s'", valErr.FieldErrors["first_name"])
	}

	// Test missing last name
	account.FirstName = "fn"
	account.LastName = ""
	valErr = account.Validate()
	if valErr.FieldErrors["last_name"] != "Please add last name" {
		t.Errorf("Expected missing last name error, but got: '%s'", valErr.FieldErrors["last_name"])
	}

	// Test missing institute
	account.LastName = "ln"
	account.Institute = ""
	valErr = account.Validate()
	if valErr.FieldErrors["institute"] != "Please add institute" {
		t.Errorf("Expected missing institute error, but got: '%s'", valErr.FieldErrors["institute"])
	}

	// Test missing department
	account.Institute = "Inst"
	account.Department = ""
	valErr = account.Validate()
	if valErr.FieldErrors["department"] != "Please add department" {
		t.Errorf("Expected missing department error, but got: '%s'", valErr.FieldErrors["department"])
	}

	// Test missing city
	account.Department = "Dep"
	account.City = ""
	valErr = account.Validate()
	if valErr.FieldErrors["city"] != "Please add city" {
		t.Errorf("Expected missing city error, but got: '%s'", valErr.FieldErrors["city"])
	}

	// Test missing country
	account.City = "cty"
	account.Country = ""
	valErr = account.Validate()
	if valErr.FieldErrors["country"] != "Please add country" {
		t.Errorf("Expected missing country error, but got: '%s'", valErr.FieldErrors["country"])
	}

	// Test valid
	account.Country = "ctry"
	valErr = account.Validate()
	if valErr.Message != "" {
		t.Errorf("Expected valid registration , but got error in fields: '%s'", valErr.FieldErrors)
	}
}
