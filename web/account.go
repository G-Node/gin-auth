// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"encoding/json"
	"net/http"

	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
	"github.com/gorilla/mux"
)

// ListAccounts is a handler which returns a list of existing accounts as JSON
func ListAccounts(w http.ResponseWriter, r *http.Request) {
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	if !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to list accounts forbidden", http.StatusUnauthorized)
		return
	}

	accounts := data.ListAccounts()
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(accounts)
}

// GetAccount is a handler which returns a requested account as JSON
func GetAccount(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	account, ok := data.GetAccountByLogin(login)
	if !ok {
		PrintErrorJSON(w, r, "The requested account does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID != account.UUID || !oauth.Match.Contains("account-read") && !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(account)
}

// UpdateAccount is a handler which updated all updatable fields of an account (Title, FirstName,
// MiddleName and LastName) and returns the updated account as JSON
func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	account, ok := data.GetAccountByLogin(login)
	if !ok {
		PrintErrorJSON(w, r, "The requested account does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID != account.UUID || !oauth.Match.Contains("account-write") && !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(account)
	if err != nil {
		PrintErrorJSON(w, r, "Error while processing account", http.StatusBadRequest)
		return
	}

	err = account.Update()
	if err != nil {
		PrintErrorJSON(w, r, "Error while processing account", http.StatusBadRequest)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(account)
}

// UpdateAccountPassword is a handler which parses the old and new password from the request body and
// updates the accounts password. Returns StatusOK and an empty body on success.
func UpdateAccountPassword(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	account, ok := data.GetAccountByLogin(login)
	if !ok {
		PrintErrorJSON(w, r, "The requested account does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID != account.UUID || !oauth.Match.Contains("account-write") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	pwData := &struct {
		PasswordOld       string `json:"password_old"`
		PasswordNew       string `json:"password_new"`
		PasswordNewRepeat string `json:"password_new_repeat"`
	}{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(pwData)

	if !account.VerifyPassword(pwData.PasswordOld) {
		err := &util.ValidationError{
			Message:     "Unable to set password",
			FieldErrors: map[string]string{"password_old": "Wrong password"}}
		PrintErrorJSON(w, r, err, http.StatusBadRequest)
		return
	}

	if len(pwData.PasswordNew) < 8 {
		err := &util.ValidationError{
			Message:     "Unable to set password",
			FieldErrors: map[string]string{"password_new": "Password must be at least 6 characters long"}}
		PrintErrorJSON(w, r, err, http.StatusBadRequest)
		return
	}

	if pwData.PasswordNew != pwData.PasswordNewRepeat {
		err := &util.ValidationError{
			Message:     "Unable to set password",
			FieldErrors: map[string]string{"password_new_repeat": "Repeated password does not match"}}
		PrintErrorJSON(w, r, err, http.StatusBadRequest)
		return
	}

	account.SetPassword(pwData.PasswordNew)
	err := account.Update()

	if err != nil {
		PrintErrorJSON(w, r, err, http.StatusInternalServerError)
		return
	}
}
