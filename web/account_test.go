// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"bytes"
	"encoding/json"
	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	accessTokenAlice      = "3N7MP7M7"
	accessTokenBob        = "LJ3W7ZFK" // is expired
	accessTokenAliceAdmin = "KDEW57D4" // has scope 'account-admin'
	keyPrintAlice         = "A3tkBXFQWkjU6rzhkofY55G7tPR_Lmna4B-WEGVFXOQ"
	keyPrintAliceNew      = "WHHqtkitF7o-EyTWFgdFKCYhU1PElLnK3U0luzyc0ko"
)

type jsonAccount struct {
	URL        string    `json:"url"`
	UUID       string    `json:"uuid"`
	Login      string    `json:"login"`
	Title      *string   `json:"title"`
	FirstName  string    `json:"first_name"`
	MiddleName *string   `json:"middle_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func TestGetAccount(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("GET", "/api/accounts/alice", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("GET", "/api/accounts/alice", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// expired token
	request, _ = http.NewRequest("GET", "/api/accounts/bob", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenBob)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// non existing account
	request, _ = http.NewRequest("GET", "/api/accounts/foo", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// not own account
	request, _ = http.NewRequest("GET", "/api/accounts/bob", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all ok (own account)
	request, _ = http.NewRequest("GET", "/api/accounts/alice", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	acc := &jsonAccount{}
	err := json.NewDecoder(response.Body).Decode(acc)
	if err != nil {
		t.Error(err)
	}
	if acc.Login != "alice" {
		t.Error("Account login expected to be 'alice'")
	}

	// all ok (token with admin scope)
	request, _ = http.NewRequest("GET", "/api/accounts/bob", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAliceAdmin)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	acc = &jsonAccount{}
	err = json.NewDecoder(response.Body).Decode(acc)
	if err != nil {
		t.Error(err)
	}
	if acc.Login != "bob" {
		t.Error("Account login expected to be 'bob'")
	}
}

func TestListAccounts(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("GET", "/api/accounts", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("GET", "/api/accounts", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// insufficient scope
	request, _ = http.NewRequest("GET", "/api/accounts", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("GET", "/api/accounts", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAliceAdmin)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	accounts := []jsonAccount{}
	err := json.NewDecoder(response.Body).Decode(&accounts)
	if err != nil {
		t.Error(err)
	}
	if len(accounts) != 2 {
		t.Error("Two accounts expected in response")
	}
	acc := accounts[0]
	if acc.Login != "alice" {
		t.Error("Account login expected to be 'alice'")
	}
}

func TestUpdateAccount(t *testing.T) {
	mkBody := func() io.Reader {
		title := "Dr"
		acc := &jsonAccount{Title: &title, FirstName: "Alix", LastName: "Bonenfant"}
		b, _ := json.Marshal(acc)
		return bytes.NewReader(b)
	}
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("PUT", "/api/accounts/alice", mkBody())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("PUT", "/api/accounts/alice", mkBody())
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong account
	request, _ = http.NewRequest("PUT", "/api/accounts/bob", mkBody())
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all ok (own account)
	request, _ = http.NewRequest("PUT", "/api/accounts/alice", mkBody())
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	acc := &jsonAccount{}
	err := json.NewDecoder(response.Body).Decode(acc)
	if err != nil {
		t.Error(err)
	}
	if acc.FirstName != "Alix" {
		t.Error("Account FirstName expected to be 'Alix'")
	}
	if *acc.Title != "Dr" {
		t.Error("Account Title expected to be 'Dr'")
	}
	if acc.LastName != "Bonenfant" {
		t.Error("Account FirstName expected to be 'Alix'")
	}
}

func TestUpdateAccountPassword(t *testing.T) {
	mkBody := func(old, new, repeat string) io.Reader {
		pw := &struct {
			PasswordOld       string `json:"password_old"`
			PasswordNew       string `json:"password_new"`
			PasswordNewRepeat string `json:"password_new_repeat"`
		}{old, new, repeat}
		b, _ := json.Marshal(pw)
		return bytes.NewReader(b)
	}
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("testtest", "TestTest", "TestTest"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("testtest", "TestTest", "TestTest"))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong account
	request, _ = http.NewRequest("PUT", "/api/accounts/bob/password", mkBody("testtest", "TestTest", "TestTest"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong password
	request, _ = http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("WRONG!", "TestTest", "TestTest"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// too short password
	request, _ = http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("testtest", "Test", "Test"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong repeated password
	request, _ = http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("testtest", "TestTest", "TestFooo"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("PUT", "/api/accounts/alice/password", mkBody("testtest", "TestTest", "TestTest"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	acc, _ := data.GetAccountByLogin("alice")
	if !acc.VerifyPassword("TestTest") {
		t.Error("Unable to verify password")
	}
}

type jsonKey struct {
	URL         string    `json:"url"`
	Fingerprint string    `json:"fingerprint"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Login       string    `json:"login"`
	AccountURL  string    `json:"account_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func TestListAccountKeys(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("GET", "/api/accounts/alice/keys", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("GET", "/api/accounts/alice/keys", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// not existing account
	request, _ = http.NewRequest("GET", "/api/accounts/doesnotexist/keys", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("GET", "/api/accounts/alice/keys", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	keys := []jsonKey{}
	dec := json.NewDecoder(response.Body)
	err := dec.Decode(&keys)
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 1 {
		t.Error("Expected list with one key")
	}
	key := keys[0]
	if key.Login != "alice" {
		t.Errorf("Login expected to be 'alice' but was %s", key.Login)
	}
	if key.Fingerprint != keyPrintAlice {
		t.Errorf("Fingerprint expected to be '%s' but was '%s'", keyPrintAlice, key.Fingerprint)
	}
}

func TestGetKey(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("GET", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("GET", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// not existing key
	request, _ = http.NewRequest("GET", "/api/keys/doesnotexist", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("GET", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	key := jsonKey{}
	dec := json.NewDecoder(response.Body)
	err := dec.Decode(&key)
	if err != nil {
		t.Error(err)
	}
	if key.Login != "alice" {
		t.Errorf("Login expected to be 'alice' but was %s", key.Login)
	}
	if key.Fingerprint != keyPrintAlice {
		t.Errorf("Fingerprint expected to be '%s' but was '%s'", keyPrintAlice, key.Fingerprint)
	}
}

func TestCreateKey(t *testing.T) {
	mkBody := func(key, description string) io.Reader {
		pw := &struct {
			Key         string `json:"key"`
			Description string `json:"description"`
		}{key, description}
		b, _ := json.Marshal(pw)
		return bytes.NewReader(b)
	}
	keyBytes, _ := ioutil.ReadFile(conf.GetResourceFile("fixtures", "id_rsa.pub"))
	keyStr := string(keyBytes)

	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("POST", "/api/accounts/alice/keys", mkBody(keyStr, "desc"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("POST", "/api/accounts/alice/keys", mkBody(keyStr, "desc"))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong account
	request, _ = http.NewRequest("POST", "/api/accounts/doesnotexist/keys", mkBody(keyStr, "desc"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// no authorization header
	request, _ = http.NewRequest("POST", "/api/accounts/alice/keys", mkBody(keyStr, "desc"))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
	key := jsonKey{}
	dec := json.NewDecoder(response.Body)
	err := dec.Decode(&key)
	if err != nil {
		t.Error(err)
	}
	if key.Login != "alice" {
		t.Errorf("Login expected to be 'alice' but was %s", key.Login)
	}
	if key.Key != keyStr {
		t.Error("Key does not match")
	}
	if key.Fingerprint != keyPrintAliceNew {
		t.Errorf("Fingerprint expected to be '%s' but was '%s'", keyPrintAliceNew, key.Fingerprint)
	}
}

func TestDeleteKey(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no authorization header
	request, _ := http.NewRequest("DELETE", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong token
	request, _ = http.NewRequest("DELETE", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// not existing key
	request, _ = http.NewRequest("DELETE", "/api/keys/doesnotexist", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("DELETE", "/api/keys/"+keyPrintAlice, strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+accessTokenAlice)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}
