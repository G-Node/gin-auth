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
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/G-Node/gin-auth/data"
	"github.com/gorilla/mux"
)

// InitTestHttpHandler initializes a handler with all registered routes and returns it
// along with a response recorder.
func InitTestHttpHandler(t *testing.T) http.Handler {
	data.InitTestDb(t)
	router := mux.NewRouter()
	router.NotFoundHandler = &NotFoundHandler{}
	RegisterRoutes(router)
	return router
}

func TestOAuthHandler(t *testing.T) {
	data.InitTestDb(t)

	r := mux.NewRouter()
	r.NotFoundHandler = &NotFoundHandler{}

	var called, authorized bool
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, authorized = OAuthToken(r)
	})

	handler := OAuthHandler("account-admin")(protected)

	// missing authorization header
	called, authorized = false, false
	request, _ := http.NewRequest("GET", "/", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if called || authorized || response.Code != http.StatusUnauthorized {
		t.Error("Request should not be authorized")
	}

	// wrong authorization header
	called, authorized = false, false
	request, _ = http.NewRequest("GET", "/", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer doesnotexist")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if called || authorized || response.Code != http.StatusUnauthorized {
		t.Error("Request should not be authorized")
	}

	// insufficient scope
	called, authorized = false, false
	request, _ = http.NewRequest("GET", "/", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer 3N7MP7M7")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if called || authorized || response.Code != http.StatusUnauthorized {
		t.Error("Request should not be authorized")
	}

	handler = OAuthHandler("account-read")(protected)

	// all OK
	called, authorized = false, false
	request, _ = http.NewRequest("GET", "/", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer 3N7MP7M7")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if !called || !authorized || response.Code != http.StatusOK {
		t.Error("Request should be authorized")
	}

	// is oauth info deleted
	_, ok := OAuthToken(request)
	if ok {
		t.Error("OAuth info should be removed")
	}
}

func newAuthQuery() url.Values {
	query := url.Values{}
	query.Add("response_type", "code")
	query.Add("client_id", "gin")
	query.Add("redirect_uri", "https://localhost:8081/login")
	query.Add("scope", "repo-read repo-write")
	query.Add("state", "testcode")
	return query
}

func TestAuthorize(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// missing query param
	request, _ := http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong response type
	query := newAuthQuery()
	query.Set("response_type", "wrong")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong scope
	query = newAuthQuery()
	query.Set("scope", "foo,bar")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong client id
	query = newAuthQuery()
	query.Set("client_id", "doesnotexist")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong redirect
	query = newAuthQuery()
	query.Set("redirect_uri", "https://example.com/invalid")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// all OK
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = newAuthQuery().Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.Path != "/oauth/login_page" {
		t.Errorf("Wrong redirect")
	}
	if redirect.Query().Get("request_id") == "" {
		t.Errorf("Request id not found")
	}
}

func TestLoginPage(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// missing query param
	request, _ := http.NewRequest("GET", "/oauth/login_page", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong request_id
	request, _ = http.NewRequest("GET", "/oauth/login_page?request_id=doesnotexist", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// valid request_id
	request, _ = http.NewRequest("GET", "/oauth/login_page?request_id=U7JIKKYI", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}

func newLoginBody() *url.Values {
	body := &url.Values{}
	body.Add("request_id", "U7JIKKYI")
	body.Add("login", "bob")
	body.Add("password", "testtest")
	return body
}

func TestLogin(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// form param missing
	request, _ := http.NewRequest("POST", "/oauth/login", strings.NewReader(""))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong request id
	body := newLoginBody()
	body.Set("request_id", "doesnotexist")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// wrong login
	body = newLoginBody()
	body.Set("login", "doesnotexist")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}

	// wrong password
	body = newLoginBody()
	body.Set("password", "notapassword")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}

	// all OK
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(newLoginBody().Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.Path != "/oauth/approve_page" {
		t.Error("Wrong redirect")
	}
	if redirect.Query().Get("request_id") == "" {
		t.Error("Request id not found")
	}
}

func TestApprovePage(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// missing query param
	request, _ := http.NewRequest("GET", "/oauth/approve_page", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong request_id
	request, _ = http.NewRequest("GET", "/oauth/approve_page?request_id=doesnotexist", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// valid request_id
	request, _ = http.NewRequest("GET", "/oauth/approve_page?request_id=B4LIMIMB", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}

func newApproveBody() *url.Values {
	body := &url.Values{}
	body.Add("request_id", "B4LIMIMB")
	body.Add("scope", "repo-read")
	body.Add("scope", "repo-write")
	return body
}

func TestApprove(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// wrong request id
	body := newApproveBody()
	body.Set("request_id", "doesnotexist")
	request, _ := http.NewRequest("POST", "/oauth/approve", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all OK
	request, _ = http.NewRequest("POST", "/oauth/approve", strings.NewReader(newApproveBody().Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.Query().Get("code") == "" {
		t.Error("Code not found")
	}
}

func newTokenBody() *url.Values {
	body := &url.Values{}
	body.Add("redirect_uri", "https://localhost:8081/login")
	body.Add("code", "HGZQP6WE")
	body.Add("grant_type", "authorization_code")
	return body
}

func TestToken(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// wrong client id
	body := newTokenBody()
	request, _ := http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("doesnotexist", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong client secret
	body = newTokenBody()
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "notsosecret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong code
	body = newTokenBody()
	body.Set("code", "reallywrongcode")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK
	body = newTokenBody()
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	data := &struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}{}
	json.Unmarshal(response.Body.Bytes(), data)
	if data.AccessToken == "" {
		t.Error("No token recieved")
	}
	if data.RefreshToken == "" {
		t.Error("No token recieved")
	}
	if data.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}
}

type testValidateResponse struct {
	URL        string    `json:"url"`
	JTI        string    `json:"jti"`
	EXP        time.Time `json:"exp"`
	ISS        string    `json:"iss"`
	Login      string    `json:"login"`
	AccountURL string    `json:"account_url"`
	Scope      []string  `json:"scope"`
}

func TestValidate(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// wrong token
	request, _ := http.NewRequest("GET", "/oauth/validate/doesnotexist", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// expired token
	request, _ = http.NewRequest("GET", "/oauth/validate/LJ3W7ZFK", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// valid token
	request, _ = http.NewRequest("GET", "/oauth/validate/3N7MP7M7", strings.NewReader(""))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	result := &struct {
		URL        string    `json:"url"`
		JTI        string    `json:"jti"`
		EXP        time.Time `json:"exp"`
		ISS        string    `json:"iss"`
		Login      string    `json:"login"`
		AccountURL string    `json:"account_url"`
		Scope      []string  `json:"scope"`
	}{}
	json.Unmarshal(response.Body.Bytes(), result)
	if result.JTI != "3N7MP7M7" {
		t.Errorf("JTI expected to be '3N7MP7M7' but was '%s'", result.JTI)
	}
	if result.ISS != "gin-auth" {
		t.Errorf("ISS expected to be 'gin-auth' but was '%s'", result.ISS)
	}
	if result.Login != "alice" {
		t.Errorf("Login expected to be 'alice' but was '%s'", result.Login)
	}
}
