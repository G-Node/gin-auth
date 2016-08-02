// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
)

func TestResetInitPage(t *testing.T) {
	handler := InitTestHttpHandler(t)
	const resetURL = "/oauth/reset_init_page"

	request, _ := http.NewRequest("GET", resetURL, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
}

func TestResetInit(t *testing.T) {
	handler := InitTestHttpHandler(t)

	const resetInitURL = "/oauth/reset_init"
	const disabledLogin = "inact_log4"
	const disabledEmail = "email4@example.com"
	const enabledLogin = "inact_log1"

	// Test post empty body
	request, _ := http.NewRequest("POST", resetInitURL, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Please enter your login or e-mail address" {
		t.Errorf("Expected empty id field message but got '%s'", response.Header().Get("Warning"))
	}

	mkBody := &url.Values{}
	mkBody.Add("Credential", "iDoNotExist")

	// Test invalid login
	request, _ = http.NewRequest("POST", resetInitURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Invalid login or e-mail address" {
		t.Errorf("Expected invalid login message but got '%s'", response.Header().Get("Warning"))
	}

	// Test login of disabled account
	mkBody = &url.Values{}
	mkBody.Add("Credential", disabledLogin)

	request, _ = http.NewRequest("POST", resetInitURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Invalid login or e-mail address" {
		t.Errorf("Expected invalid login message but got '%s'", response.Header().Get("Warning"))
	}

	// Test e-mail of disabled account
	mkBody = &url.Values{}
	mkBody.Add("Credential", disabledEmail)

	request, _ = http.NewRequest("POST", resetInitURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Invalid login or e-mail address" {
		t.Errorf("Expected invalid login message but got '%s'", response.Header().Get("Warning"))
	}

	// Test valid update using login
	mkBody = &url.Values{}
	mkBody.Add("Credential", enabledLogin)

	request, _ = http.NewRequest("POST", resetInitURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "" {
		t.Errorf("Expected empty warning message but got '%s'", response.Header().Get("Warning"))
	}

	// Test error when sending e-mail
	mode := conf.GetSmtpCredentials().Mode
	host := conf.GetSmtpCredentials().Host
	conf.GetSmtpCredentials().Mode = ""
	conf.GetSmtpCredentials().Host = "iDoNotExist.com"
	request, _ = http.NewRequest("POST", resetInitURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusInternalServerError but got '%d'", response.Code)
	}

	// reset smtp configuration
	conf.GetSmtpCredentials().Mode = mode
	conf.GetSmtpCredentials().Host = host
}

func TestResetPage(t *testing.T) {
	handler := InitTestHttpHandler(t)
	const resetURL = "/oauth/reset_page"
	const codeKey = "reset_code"
	const codeInvalid = "iDoNotExist"
	const codeDisabled = "rc_c"
	const codeValid = "rc_a"

	// Test missing password reset code
	request, _ := http.NewRequest("GET", resetURL, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected StatusBadRequest on missing reset code but got '%d'", response.Code)
	}

	// Test invalid password reset code
	request, _ = http.NewRequest("GET", resetURL, strings.NewReader(""))
	q := request.URL.Query()
	q.Add(codeKey, codeInvalid)
	request.URL.RawQuery = q.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on invalid reset code but got '%d'", response.Code)
	}

	// Test valid password reset code of disabled account
	request, _ = http.NewRequest("GET", resetURL, strings.NewReader(""))
	q = request.URL.Query()
	q.Add(codeKey, codeDisabled)
	request.URL.RawQuery = q.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on disabled reset code but got '%d'", response.Code)
	}

	// Test valid password reset code
	request, _ = http.NewRequest("GET", resetURL, strings.NewReader(""))
	q = request.URL.Query()
	q.Add(codeKey, codeValid)
	request.URL.RawQuery = q.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}
}

func TestReset(t *testing.T) {
	handler := InitTestHttpHandler(t)
	const resetURL = "/oauth/reset"
	const codeKey = "ResetCode"
	const codeInvalid = "iDoNotExist"
	const codeDisabled = "rc_c"
	const codeValid = "rc_a"
	const codeValidInactive = "rc_b"

	// Test empty body
	request, _ := http.NewRequest("POST", resetURL, strings.NewReader(""))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on missing reset code but got '%d'", response.Code)
	}

	// Test invalid password reset code
	mkBody := &url.Values{}
	mkBody.Add(codeKey, codeInvalid)

	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on invalid reset code but got '%d'", response.Code)
	}

	// Test valid password reset code of disabled account
	mkBody.Set(codeKey, codeDisabled)

	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on disabled reset code but got '%d'", response.Code)
	}

	// Test valid password reset code, missing password
	mkBody.Set(codeKey, codeValid)
	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Please enter password and password control" {
		t.Errorf("Expected empty password warning but got '%s'", response.Header().Get("Warning"))
	}

	// Test valid password reset code, password and password control missmatch
	mkBody.Add("Password", "pw")
	mkBody.Add("PasswordControl", "pwcontrol")
	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Provided password did not match password control" {
		t.Errorf("Expected empty password warning but got '%s'", response.Header().Get("Warning"))
	}

	// Test valid password reset code, password too long
	s := []string{}
	for i := 0; i < 513; i++ {
		s = append(s, "s")
	}
	js := strings.Join(s, "")

	mkBody.Set("Password", js)
	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}
	if response.Header().Get("Warning") != "Entry too long, please shorten to 512 characters" {
		t.Errorf("Expected empty password warning but got '%s'", response.Header().Get("Warning"))
	}

	// Test valid password reset code, reset of pw code
	account, exists := data.GetAccountByResetPWCode(codeValid)
	if !exists {
		t.Errorf("Account with reset code '%s' does not exist", codeValid)
	}
	id := account.UUID
	_, exists = data.GetAccount(id)
	if exists {
		t.Errorf("Account with id '%s' reset code '%s' should not be active", id, codeValid)
	}
	pwHash := account.PWHash

	mkBody.Set("Password", "pw")
	mkBody.Set("PasswordControl", "pw")
	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}

	account, exists = data.GetAccount(id)
	if !exists {
		t.Errorf("Account with id '%s' should be active", id)
	}
	if account.PWHash == pwHash {
		t.Errorf("Password of Account with id '%s' has not been updated", id)
	}
	if account.ResetPWCode.String != "" {
		t.Errorf("Password reset code of Account with id '%s' has not been deleted", id)
	}

	// Test valid password reset code, reset of pw code, reset of activation code
	account, exists = data.GetAccountByResetPWCode(codeValidInactive)
	if !exists {
		t.Errorf("Account with reset code '%s' does not exist", codeValidInactive)
	}
	id = account.UUID
	_, exists = data.GetAccount(id)
	if exists {
		t.Errorf("Account with id '%s' reset code '%s' should not be active", id, codeValidInactive)
	}
	pwHash = account.PWHash

	mkBody.Set(codeKey, codeValidInactive)
	request, _ = http.NewRequest("POST", resetURL, strings.NewReader(mkBody.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid reset code but got '%d'", response.Code)
	}

	account, exists = data.GetAccount(id)
	if !exists {
		t.Errorf("Account with id '%s' should be active", id)
	}
	if account.PWHash == pwHash {
		t.Errorf("Password of Account with id '%s' has not been updated", id)
	}
	if account.ResetPWCode.String != "" {
		t.Errorf("Password reset code of Account with id '%s' has not been removed", id)
	}
	if account.ActivationCode.String != "" {
		t.Errorf("Activation code of Account with id '%s' has not been removed", id)
	}
}
