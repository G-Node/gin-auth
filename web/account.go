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

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
	"github.com/gorilla/mux"
)

// ListAccounts is a handler which returns a list of existing accounts as JSON
func ListAccounts(w http.ResponseWriter, r *http.Request) {
	isAdmin := false
	if oauth, ok := OAuthToken(r); ok {
		isAdmin = oauth.Match.Contains("account-admin")
	}

	var accounts []data.Account
	search := r.URL.Query().Get("q")
	if search != "" {
		accounts = data.SearchAccounts(search)
	} else {
		accounts = data.ListAccounts()
	}

	marshal := make([]data.AccountMarshaler, 0, len(accounts))
	for i := 0; i < len(accounts); i++ {
		acc := &accounts[i]
		marshal = append(marshal, data.AccountMarshaler{
			WithMail:        isAdmin || acc.IsEmailPublic,
			WithAffiliation: isAdmin || acc.IsAffiliationPublic,
			Account:         acc,
		})
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(marshal)
}

// GetAccount is a handler which returns a requested account as JSON
func GetAccount(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]

	account, ok := data.GetAccountByLogin(login)
	if !ok {
		PrintErrorJSON(w, r, "The requested account does not exist", http.StatusNotFound)
		return
	}

	isAdmin := false
	isOwner := false
	if oauth, ok := OAuthToken(r); ok {
		isAdmin = oauth.Match.Contains("account-admin")
		isOwner = oauth.Token.AccountUUID.String == account.UUID && oauth.Match.Contains("account-write")
	}

	marshal := &data.AccountMarshaler{
		WithMail:        account.IsEmailPublic || isOwner || isAdmin,
		WithAffiliation: account.IsAffiliationPublic || isOwner || isAdmin,
		Account:         account,
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(marshal)
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

	if oauth.Token.AccountUUID.String != account.UUID || !oauth.Match.Contains("account-write") && !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	marshal := &data.AccountMarshaler{WithMail: true, WithAffiliation: true, Account: account}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(marshal)
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
	enc.Encode(marshal)
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

	if oauth.Token.AccountUUID.String != account.UUID || !oauth.Match.Contains("account-write") {
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

	if len(pwData.PasswordNew) < 6 {
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

	err := account.UpdatePassword(pwData.PasswordNew)
	if err != nil {
		PrintErrorJSON(w, r, err, http.StatusInternalServerError)
		return
	}
}

// UpdateAccountEmail parses an e-mail address and the account password
// from a JSON request body and updates the e-mail address of the authorized account.
func UpdateAccountEmail(w http.ResponseWriter, r *http.Request) {

	login := mux.Vars(r)["login"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Missing OAuth token")
	}

	acc, ok := data.GetAccountByLogin(login)
	if !ok {
		PrintErrorJSON(w, r, "The requested account does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID.String != acc.UUID || !oauth.Match.Contains("account-write") {
		PrintErrorJSON(w, r, "Unauthorized account access", http.StatusUnauthorized)
		return
	}

	cred := &struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	dec := json.NewDecoder(r.Body)
	dec.Decode(cred)

	if !acc.VerifyPassword(cred.Password) {
		valErr := &util.ValidationError{
			Message:     "Invalid password",
			FieldErrors: map[string]string{"password": "Invalid password"}}
		PrintErrorJSON(w, r, valErr, http.StatusBadRequest)
		return
	}

	err := acc.UpdateEmail(cred.Email)
	if err != nil {
		PrintErrorJSON(w, r, err, http.StatusBadRequest)
		return
	}

	tmplFields := &struct {
		From    string
		To      string
		Subject string
		Body    string
	}{}
	tmplFields.From = conf.GetSmtpCredentials().From
	tmplFields.To = cred.Email
	tmplFields.Subject = "GIN account confirmation"
	tmplFields.Body = "The e-mail address of your GIN account has been successfully changed."

	content := util.MakeEmailTemplate("emailplain.txt", tmplFields)
	email := &data.Email{}
	err = email.Create(util.NewStringSet(cred.Email), content.Bytes())
	if err != nil {
		msg := "An error occurred trying to create change e-mail address confirmation."
		PrintErrorJSON(w, r, msg, http.StatusInternalServerError)
		return
	}
}

// ListAccountKeys is a handler which returns all ssh keys belonging to a given
// account as JSON.
func ListAccountKeys(w http.ResponseWriter, r *http.Request) {
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

	if oauth.Token.AccountUUID.String != account.UUID || !oauth.Match.Contains("account-read") && !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to requested key forbidden", http.StatusUnauthorized)
		return
	}

	keys := account.SSHKeys()
	marshal := make([]data.SSHKeyMarshaler, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		marshal = append(marshal, data.SSHKeyMarshaler{SSHKey: &keys[i], Account: account})
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(marshal)
}

// GetKey returns a single ssh key identified by its fingerprint as JSON.
func GetKey(w http.ResponseWriter, r *http.Request) {
	fingerprint := mux.Vars(r)["fingerprint"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	key, ok := data.GetSSHKey(fingerprint)
	if !ok {
		PrintErrorJSON(w, r, "The requested key does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID.String != key.AccountUUID || !oauth.Match.Contains("account-read") && !oauth.Match.Contains("account-admin") {
		PrintErrorJSON(w, r, "Access to requested key forbidden", http.StatusUnauthorized)
		return
	}

	// account is only needed for the output (maybe this can be avoided)
	account, _ := data.GetAccount(key.AccountUUID)

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(&data.SSHKeyMarshaler{SSHKey: key, Account: account})
}

// CreateKey stores a new key for a given account.
func CreateKey(w http.ResponseWriter, r *http.Request) {
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

	if oauth.Token.AccountUUID.String != account.UUID || !oauth.Match.Contains("account-write") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	key := &data.SSHKey{AccountUUID: account.UUID}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(key)
	if err != nil {
		PrintErrorJSON(w, r, err, http.StatusBadRequest)
		return
	}

	err = key.Create()
	if err != nil {
		panic(err)
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(&data.SSHKeyMarshaler{SSHKey: key, Account: account})
}

// DeleteKey removes a single ssh key identified by its fingerprint and returns
// the deleted key as JSON.
func DeleteKey(w http.ResponseWriter, r *http.Request) {
	fingerprint := mux.Vars(r)["fingerprint"]
	oauth, ok := OAuthToken(r)
	if !ok {
		panic("Request was authorized but no OAuth token is available!") // this should never happen
	}

	key, ok := data.GetSSHKey(fingerprint)
	if !ok {
		PrintErrorJSON(w, r, "The requested key does not exist", http.StatusNotFound)
		return
	}

	if oauth.Token.AccountUUID.String != key.AccountUUID || !oauth.Match.Contains("account-write") {
		PrintErrorJSON(w, r, "Access to requested account forbidden", http.StatusUnauthorized)
		return
	}

	err := key.Delete()
	if err != nil {
		panic(err)
	}

	// account is only needed for the output (maybe this can be avoided)
	account, _ := data.GetAccount(key.AccountUUID)

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(&data.SSHKeyMarshaler{SSHKey: key, Account: account})
}
