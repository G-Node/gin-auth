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
	"reflect"
	"strings"
	"testing"

	"github.com/G-Node/gin-auth/util"
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

func TestGetAccountByCredential(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const validLogin = "bob"
	const validEmail = "aclic@foo.com"

	acc, ok := GetAccountByCredential(validLogin)
	if !ok {
		t.Errorf("Account for login '%s' was not found.\n", validLogin)
	}
	if acc.Login != validLogin {
		t.Errorf("Retrieved account for '%s' but got %v.\n", validLogin, acc)
	}

	acc, ok = GetAccountByCredential(validEmail)
	if !ok {
		t.Errorf("Account for email '%s' was not found.\n", validEmail)
	}
	if acc.Email != validEmail {
		t.Errorf("Retrieved account for '%s' but got '%v'.\n", validEmail, acc)
	}

	_, ok = GetAccountByCredential("doesNotExist")
	if ok {
		t.Error("Account should not exist.")
	}

	// Test whole barrage of inactive accounts
	inactiveLogin := []string{"inact_log1", "inact_log2", "inact_log3", "inact_log4",
		"inact_log5", "inact_log6"}
	for _, v := range inactiveLogin {
		_, ok = GetAccountByCredential(v)
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

func TestSetPasswordReset(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	const disabledLogin = "inact_log4"
	const disabledEmail = "email4@example.com"
	const enabledLogin = "inact_log1"
	const enabledEmail = "email1@example.com"

	// Test empty credential
	_, ok := SetPasswordReset("")
	if ok {
		t.Error("Account should not have been updated using an empty credential")
	}

	// Test non existing credential
	_, ok = SetPasswordReset("iDoNotExist")
	if ok {
		t.Error("Account should not have been updated using non existing credential")
	}

	// Test valid login of disabled account
	_, ok = SetPasswordReset(disabledLogin)
	if ok {
		t.Error("Account should not have been updated using disabled account login")
	}

	// Test valid email of disabled account
	_, ok = SetPasswordReset(disabledEmail)
	if ok {
		t.Error("Account should not have been updated using disabled account email")
	}

	// Test valid update using login
	account, ok := SetPasswordReset(enabledLogin)
	if !ok {
		t.Errorf("Account should have been updated using valid account login '%s'", enabledLogin)
	}
	if account.ResetPWCode.String == "" {
		t.Errorf("Account should have reset pw code, but was empty (using login '%s')", enabledLogin)
	}

	// Test valid update using email
	old := account.ResetPWCode.String
	account, ok = SetPasswordReset(enabledEmail)
	if !ok {
		t.Errorf("Account should have been updated using valid account email '%s'", enabledEmail)
	}
	if account.ResetPWCode.String == "" {
		t.Errorf("Account should have reset pw code, but was empty (using email '%s')", enabledEmail)
	}
	if account.ResetPWCode.String == old {
		t.Errorf("Account should have new reset pw code, but was unchanged (using email '%s')", enabledEmail)
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

func TestAccount_UpdatePassword(t *testing.T) {
	InitTestDb(t)
	const pw = "supersecret"

	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}
	err := acc.UpdatePassword(pw)
	if err != nil {
		t.Errorf("Error updating password: '%s'", err.Error())
	}

	if acc.PWHash == pw {
		t.Error("PWHash equals plain text password")
	}
	if len(acc.PWHash) < 60 {
		t.Error("PWHash is too short")
	}
	if !acc.VerifyPassword(pw) {
		t.Error("Unable to verify password")
	}

	checkDb, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}
	if !checkDb.VerifyPassword(pw) {
		t.Error("Password update failed")
	}
}

func TestAccount_UpdateEmail(t *testing.T) {
	InitTestDb(t)
	const short = "a"
	const missing = "aaaa"
	const valid = "testaddress12@example.com"

	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}

	err := acc.UpdateEmail(short)
	if reflect.TypeOf(err).String() != "*util.ValidationError" {
		t.Errorf("Expected valid e-mail address error but got: '%s', '%s'",
			reflect.TypeOf(err).String(), err.Error())
	}
	if !strings.Contains(err.(*util.ValidationError).FieldErrors["email"], "Please use a valid e-mail address") {
		t.Errorf("Expected valid e-mail address error but got: '%s'", err.Error())
	}

	err = acc.UpdateEmail(missing)
	if reflect.TypeOf(err).String() != "*util.ValidationError" {
		t.Errorf("Expected valid e-mail address error but got: '%s', '%s'",
			reflect.TypeOf(err).String(), err.Error())
	}
	if !strings.Contains(err.(*util.ValidationError).FieldErrors["email"], "Please use a valid e-mail address") {
		t.Errorf("Expected valid e-mail address error but got: '%s'", err.Error())
	}

	// Test maximal length error
	s := []string{}
	s = append(s, "@")
	for i := 0; i < 513; i++ {
		s = append(s, "s")
	}
	js := strings.Join(s, "")

	err = acc.UpdateEmail(js)
	if reflect.TypeOf(err).String() != "*util.ValidationError" {
		t.Errorf("Expected e-mail address too long error but got: '%s', '%s'",
			reflect.TypeOf(err).String(), err.Error())
	}
	if !strings.Contains(err.(*util.ValidationError).FieldErrors["email"], "Address too long") {
		t.Errorf("Expected e-mail address too long error but got: '%s'", err.Error())
	}

	err = acc.UpdateEmail(acc.Email)
	if reflect.TypeOf(err).String() != "*util.ValidationError" {
		t.Errorf("Expected choose different e-mail address error but got: '%s', '%s'",
			reflect.TypeOf(err).String(), err.Error())
	}
	if !strings.Contains(err.(*util.ValidationError).FieldErrors["email"],
		"Please choose a different e-mail address") {
		t.Errorf("Expected choose different e-mail address error but got: '%s'", err.Error())
	}

	err = acc.UpdateEmail(valid)
	if err != nil {
		t.Errorf("Encountered unexpected error: '%s'", err.Error())
	}
	acc, ok = GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}
	if acc.Email != valid {
		t.Errorf("Expected e-mail address to be '%s', but was '%s'", valid, acc.Email)
	}
}

func TestAccount_Create(t *testing.T) {
	InitTestDb(t)

	fresh := &Account{Login: "theo", Email: "theo@example.com", FirstName: "Theo", LastName: "Test"}
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
	if len(keys) != 2 {
		t.Error("List should contain two single keys")
	}
	key := keys[0]
	if key.AccountUUID != acc.UUID {
		t.Errorf("Account uuid expected to be '%s' but was '%s'", acc.UUID, key.AccountUUID)
	}
}

func TestAccount_Update(t *testing.T) {
	InitTestDb(t)

	newLogin := "alice_in_wonderland"
	newPw := "secret"
	newTitle := "Dr."
	newFirstName := "I am actually not Alice"
	newMiddleName := "and my last name is"
	newLastName := "Badchild"
	newResetPWCode := "reset password code"
	newInstitute := "institute"
	newDepartment := "department"
	newCity := "Kierling"
	newCountry := "Iceland"
	newEmailPublic := true
	newAffiliationPublic := true

	acc, ok := GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}

	acc.SetPassword(newPw)
	acc.Login = newLogin
	acc.Title = sql.NullString{String: newTitle, Valid: true}
	acc.FirstName = newFirstName
	acc.MiddleName = sql.NullString{String: newMiddleName, Valid: true}
	acc.LastName = newLastName
	acc.Institute = newInstitute
	acc.Department = newDepartment
	acc.City = newCity
	acc.Country = newCountry
	acc.IsEmailPublic = newEmailPublic
	acc.IsAffiliationPublic = newAffiliationPublic

	err := acc.Update()
	if err != nil {
		t.Error(err)
	}

	acc, ok = GetAccount(uuidAlice)
	if !ok {
		t.Error("Account does not exist")
	}

	if acc.VerifyPassword(newPw) {
		t.Error("PWHash was updated though this should not have happened via update")
	}
	if acc.Login == newLogin {
		t.Error("Login was updated although this should never happen")
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
	if acc.Institute != newInstitute {
		t.Error("Institute was not updated")
	}
	if acc.Department != newDepartment {
		t.Error("Department was not updated")
	}
	if acc.City != newCity {
		t.Error("City was not updated")
	}
	if acc.Country != newCountry {
		t.Error("Country was not updated")
	}
	if acc.IsEmailPublic != newEmailPublic {
		t.Error("IsEmailPublic was not updated")
	}
	if acc.IsAffiliationPublic != newAffiliationPublic {
		t.Error("IsAffiliationPublic was not updated")
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

func TestAccount_RemoveActivationCode(t *testing.T) {
	InitTestDb(t)

	const login = "inact_log1"
	const activationCode = "ac_a"

	acc, ok := GetAccountByLogin(login)
	if ok {
		t.Error("Account should not be active")
	}
	acc, ok = GetAccountByActivationCode(activationCode)
	if !ok {
		t.Error("Account does not exist")
	}

	err := acc.RemoveActivationCode()
	if err != nil {
		t.Errorf("An error occurred trying to remove an activation code: '%s'", err.Error())
	}
	acc, ok = GetAccountByLogin(login)
	if !ok {
		t.Error("Account should be active")
	}
	if acc.ActivationCode.Valid {
		t.Error("Activation code should be empty")
	}
}

func TestValidate(t *testing.T) {
	InitTestDb(t)

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
	if valErr.FieldErrors["login"] != "Please choose a different login" {
		t.Errorf("Expected existing login error, but got: '%s'", valErr.FieldErrors["login"])
	}

	// Test missing login
	account.Login = ""
	valErr = account.Validate()
	if valErr.FieldErrors["login"] != "Please add login" {
		t.Errorf("Expected missing login error, but got: '%s'", valErr.FieldErrors["login"])
	}

	// Test login with invalid characters
	account.Login = "alice/"
	valErr = account.Validate()
	if !strings.Contains(valErr.FieldErrors["login"], "Please use only the following characters: ") {
		t.Errorf("Expected invalid characters error, but got: '%s'\n", valErr.FieldErrors["login"])
	}

	// Test existing email
	account.Login = "no-one_1234567890_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	valErr = account.Validate()
	if valErr.FieldErrors["email"] != "Please choose a different email address" {
		t.Errorf("Expected existing email error, but got: '%s'", valErr.FieldErrors["email"])
	}

	// Test invalid email
	account.Email = "typoemail"
	valErr = account.Validate()
	if valErr.FieldErrors["email"] != "Please add a valid e-mail address" {
		t.Errorf("Expected invalid email error, but got: '%s'", valErr.FieldErrors["email"])
	}
	account.Email = "t@"
	valErr = account.Validate()
	if valErr.FieldErrors["email"] != "Please add a valid e-mail address" {
		t.Errorf("Expected invalid email error, but got: '%s'", valErr.FieldErrors["email"])
	}

	// Test missing email
	account.Email = ""
	valErr = account.Validate()
	if valErr.FieldErrors["email"] != "Please add a valid e-mail address" {
		t.Errorf("Expected missing email error, but got: '%s'", valErr.FieldErrors["email"])
	}

	// Test valid
	account.Email = "noone@example.com"
	valErr = account.Validate()
	if valErr.Message != "" {
		t.Errorf("Expected valid registration , but got error in fields: '%s'", valErr.FieldErrors)
	}

	// Test maximal length error
	s := []string{}
	for i := 0; i < 513; i++ {
		s = append(s, "s")
	}
	js := strings.Join(s, "")

	account.Title.String = js
	account.FirstName = js
	account.MiddleName.String = js
	account.LastName = js
	account.Login = js
	account.Email = js
	account.Institute = js
	account.Department = js
	account.City = js
	account.Country = js

	valErr = account.Validate()

	if valErr.FieldErrors["title"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected title length error, but got: '%s'", valErr.FieldErrors["title"])
	}
	if valErr.FieldErrors["first_name"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected first name length error, but got: '%s'", valErr.FieldErrors["first_name"])
	}
	if valErr.FieldErrors["middle_name"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected middle name length error, but got: '%s'", valErr.FieldErrors["middle_name"])
	}
	if valErr.FieldErrors["last_name"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected last name length error, but got: '%s'", valErr.FieldErrors["last_name"])
	}
	if valErr.FieldErrors["login"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected login length error, but got: '%s'", valErr.FieldErrors["login"])
	}
	if valErr.FieldErrors["email"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected e-mail length error, but got: '%s'", valErr.FieldErrors["email"])
	}
	if valErr.FieldErrors["institute"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected institute length error, but got: '%s'", valErr.FieldErrors["institute"])
	}
	if valErr.FieldErrors["department"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected department length error, but got: '%s'", valErr.FieldErrors["department"])
	}
	if valErr.FieldErrors["city"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected city length error, but got: '%s'", valErr.FieldErrors["city"])
	}
	if valErr.FieldErrors["country"] != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected title length error, but got: '%s'", valErr.FieldErrors["country"])
	}
}
