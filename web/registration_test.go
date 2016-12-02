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

	"github.com/G-Node/gin-auth/data"
)

func TestRegistrationInit(t *testing.T) {
	const uri = "/oauth/registration_init"
	const forwardURI = "/oauth/registration_page"

	handler := InitTestHttpHandler(t)

	// Test fail on invalid response type
	queryVals := &url.Values{}
	queryVals.Add("response_type", "code")

	request, _ := http.NewRequest("GET", uri+"?"+queryVals.Encode(), strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test correct redirect
	queryVals.Set("response_type", "client")
	queryVals.Add("client_id", "gin")
	queryVals.Add("redirect_uri", "http://localhost:8080/notice")
	queryVals.Add("state", "someClientState")
	queryVals.Add("scope", "account-create")

	request, _ = http.NewRequest("GET", uri+"?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}
	loc, err := response.Result().Location()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(loc.String(), forwardURI) {
		t.Errorf("Forward to unexpected URI: %q\n", loc.String())
	}
}

func TestRegistrationPage(t *testing.T) {
	const uri = "/oauth/registration_page"
	const invalidID = "invalidID"
	const invalidCodeID = "B4LIMIMB"
	const validID = "QPJ64HK0"

	handler := InitTestHttpHandler(t)

	// Test fail with missing requestID
	request, _ := http.NewRequest("GET", uri, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test fail with invalid grant request ID
	queryVals := &url.Values{}
	queryVals.Add("request_id", invalidID)

	request, _ = http.NewRequest("GET", uri+"?"+queryVals.Encode(), strings.NewReader(""))
	request.URL.Query().Add("request_id", invalidID)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test fail with invalid grant request scope
	queryVals.Set("request_id", invalidCodeID)
	request, _ = http.NewRequest("GET", uri+"?"+queryVals.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test success with valid grant request
	queryVals.Set("request_id", validID)
	request, _ = http.NewRequest("GET", uri+"?"+queryVals.Encode(), strings.NewReader(""))
	request.URL.Query().Add("request_id", validID)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}

func TestRegistrationHandler(t *testing.T) {
	data.InitTestDb(t)
	f := func(id string, resolve string) bool {
		return id != "" && id == resolve
	}
	handler := RegistrationHandler(f)

	const registrationURL = "/oauth/registration"
	const registeredPageURL = "/oauth/registered_page"

	body := &url.Values{}
	body.Add("Title", "Title")
	body.Add("FirstName", "First Name")
	body.Add("MiddleName", "Middle Name")
	body.Add("LastName", "Last Name")
	body.Add("Login", "tl")
	body.Add("Email", "testemail@example.com")
	body.Add("IsEmailPublic", "true")
	body.Add("Institute", "Institute")
	body.Add("Department", "Department")
	body.Add("City", "City")
	body.Add("Country", "Country")
	body.Add("IsAffiliationPublic", "true")
	body.Add("Password", "pw")
	body.Add("PasswordControl", "pw")

	emails, _ := data.GetQueuedEmails()
	num := len(emails)
	// test that a request without a posted form does not redirect to another page
	request, _ := http.NewRequest("POST", registrationURL, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.String() != "" {
		t.Errorf("Expected empty location header, but was '%s'", redirect.String())
	}

	// test that a request without valid request id returns a BadRequest
	request, _ = http.NewRequest("POST", registrationURL, strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d but got %d\n", http.StatusBadRequest, response.Code)
	}

	// test that a request with correct form content but missing captcha stays on the same page
	body.Add("request_id", "QPJ64HK0")
	request, _ = http.NewRequest("POST", registrationURL, strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	redirect, err = url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.String() != "" {
		t.Errorf("Expected empty location header, but was '%s'", redirect.String())
	}

	// test that a request with correct form content but incorrect captcha stays on the same page
	body.Add("captcha_id", "test")
	request, _ = http.NewRequest("POST", registrationURL, strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	redirect, err = url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.String() != "" {
		t.Errorf("Expected empty location header, but was '%s'", redirect.String())
	}

	emails, _ = data.GetQueuedEmails()
	if len(emails) != num {
		t.Errorf("Expected e-mail queue to contain '%d' entries but had '%d'", num, len(emails))
	}

	// test that a request with correct form content redirects to registered_page and contains a request token
	body.Add("captcha_resolve", "test")
	request, _ = http.NewRequest("POST", registrationURL, strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Errorf("Expected status %d but got status %d\n", http.StatusFound, response.Code)
	}
	redirect, err = url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(redirect.String(), registeredPageURL) {
		t.Errorf("Expected to be redirected to '%s', but was '%s'", registeredPageURL, redirect.String())
	}
	_, exists := data.GetGrantRequest(redirect.Query().Get("request_id"))
	if !exists {
		t.Errorf("Missing or invalid grant request token in redirect query: %q\n", redirect.RawQuery)
	}

	emails, _ = data.GetQueuedEmails()
	if len(emails) == num {
		t.Error("E-Mail entry was not created")
	}
}

func TestRegisteredPage(t *testing.T) {
	const uri = "/oauth/registered_page"
	const invalidToken = "iDoNotExist"
	const validToken = "QPJ64HK0"

	handler := InitTestHttpHandler(t)

	// Test fail on missing URI query
	request, _ := http.NewRequest("GET", uri, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test fail on invalid grant request token
	urlValue := &url.Values{}
	urlValue.Add("request_id", invalidToken)

	request, _ = http.NewRequest("GET", uri+"?"+urlValue.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// Test redirect with valid grant request token
	urlValue.Set("request_id", validToken)
	request, _ = http.NewRequest("GET", uri+"?"+urlValue.Encode(), strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	grantRequest, exists := data.GetGrantRequest(validToken)
	if !exists {
		t.Error("Grant request does not exist")
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(redirect.String(), grantRequest.RedirectURI) {
		t.Errorf("Expected to be redirected to '%s', but was '%s'", grantRequest.RedirectURI, redirect.String())
	}
	if !strings.Contains(redirect.Query().Get("state"), grantRequest.State) {
		t.Errorf("Missing or invalid state in reponse query: '%s'", redirect.RawQuery)
	}
}

func TestActivation(t *testing.T) {
	handler := InitTestHttpHandler(t)
	const activationURL = "/oauth/activation"
	const activationCodeDisabled = "ac_b"
	const activationCode = "ac_a"

	// Test missing query
	request, _ := http.NewRequest("GET", activationURL, strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected StatusBadRequest on empty activationCode but got '%d'", response.Code)
	}

	// Test invalid activation code
	request, _ = http.NewRequest("GET", activationURL, strings.NewReader(""))
	q := request.URL.Query()
	q.Add("activation_code", "iDoNotExist")
	request.URL.RawQuery = q.Encode()

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on invalid activationCode but got '%d'", response.Code)
	}

	// Test activation code of disabled account
	request, _ = http.NewRequest("GET", activationURL, strings.NewReader(""))
	q = request.URL.Query()
	q.Add("activation_code", activationCodeDisabled)
	request.URL.RawQuery = q.Encode()

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound on disabled account but got '%d'", response.Code)
	}

	// Test valid activation
	account, exists := data.GetAccountByActivationCode(activationCode)
	if !exists {
		t.Errorf("Error on fetching account by activation code '%s'", activationCode)
	}
	if account.ActivationCode.String != activationCode {
		t.Errorf("Expected activation code to be '%s' but got '%s'",
			activationCode, account.ActivationCode.String)
	}
	accountLogin := account.Login

	request, _ = http.NewRequest("GET", activationURL, strings.NewReader(""))
	q = request.URL.Query()
	q.Add("activation_code", activationCode)
	request.URL.RawQuery = q.Encode()

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK on valid activationCode but got '%d'", response.Code)
	}

	account, exists = data.GetAccountByLogin(accountLogin)
	if !exists {
		t.Errorf("Error on fetching account by login '%s'", accountLogin)
	}
	if account.ActivationCode.String != "" || account.ActivationCode.Valid {
		t.Errorf("Activation code should be cleared after activation but was '%s', '%t'",
			account.ActivationCode.String, account.ActivationCode.Valid)
	}
}
