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

const (
	sessionCookieBob     = "4KDNO8T0"
	sessionCookieExpired = "2MFZZUKI"
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

func TestAuthorize(t *testing.T) {
	handler := InitTestHttpHandler(t)

	mkQuery := func() url.Values {
		query := url.Values{}
		query.Add("response_type", "code")
		query.Add("client_id", "gin")
		query.Add("redirect_uri", "https://localhost:8081/login")
		query.Add("scope", "repo-read repo-write")
		query.Add("state", "testcode")
		return query
	}

	// missing query param
	request, _ := http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong response type
	query := mkQuery()
	query.Set("response_type", "wrong")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong scope
	query = mkQuery()
	query.Set("scope", "foo,bar")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong client id
	query = mkQuery()
	query.Set("client_id", "doesnotexist")
	request, _ = http.NewRequest("GET", "/oauth/authorize", strings.NewReader(""))
	request.URL.RawQuery = query.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong redirect
	query = mkQuery()
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
	request.URL.RawQuery = mkQuery().Encode()
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

func TestLoginWithSession(t *testing.T) {
	handler := InitTestHttpHandler(t)

	// no request id
	request, _ := http.NewRequest("GET", "/oauth/login", strings.NewReader(""))
	request.AddCookie(&http.Cookie{Name: cookieName, Value: sessionCookieBob})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong request id
	request, _ = http.NewRequest("GET", "/oauth/login", strings.NewReader(""))
	request.URL.RawQuery = url.Values{"request_id": []string{"doesnotexist"}}.Encode()
	request.AddCookie(&http.Cookie{Name: cookieName, Value: sessionCookieBob})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// no session
	request, _ = http.NewRequest("GET", "/oauth/login", strings.NewReader(""))
	request.URL.RawQuery = url.Values{"request_id": []string{"U7JIKKYI"}}.Encode()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// expired session
	request, _ = http.NewRequest("GET", "/oauth/login", strings.NewReader(""))
	request.URL.RawQuery = url.Values{"request_id": []string{"U7JIKKYI"}}.Encode()
	request.AddCookie(&http.Cookie{Name: cookieName, Value: sessionCookieExpired})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all ok
	request, _ = http.NewRequest("GET", "/oauth/login", strings.NewReader(""))
	request.URL.RawQuery = url.Values{"request_id": []string{"U7JIKKYI"}}.Encode()
	request.AddCookie(&http.Cookie{Name: cookieName, Value: sessionCookieBob})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Error(response)
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}
	redirect, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.Path != "/oauth/approve_page" {
		t.Errorf("Wrong redirect %s", redirect.Path)
	}
	if redirect.Query().Get("request_id") == "" {
		t.Error("Request id not found")
	}
}

func TestLoginWithCredentials(t *testing.T) {
	handler := InitTestHttpHandler(t)

	mkBody := func() *url.Values {
		body := &url.Values{}
		body.Add("request_id", "U7JIKKYI")
		body.Add("login", "bob")
		body.Add("password", "testtest")
		return body
	}

	// form param missing
	request, _ := http.NewRequest("POST", "/oauth/login", strings.NewReader(""))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusBadRequest, response.Code)
	}

	// wrong request id
	body := mkBody()
	body.Set("request_id", "doesnotexist")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// wrong login
	body = mkBody()
	body.Set("login", "doesnotexist")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}

	// wrong password
	body = mkBody()
	body.Set("password", "notapassword")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusFound, response.Code)
	}

	// all OK
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(mkBody().Encode()))
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

func TestApprove(t *testing.T) {
	handler := InitTestHttpHandler(t)

	mkBody := func() *url.Values {
		body := &url.Values{}
		body.Add("request_id", "B4LIMIMB")
		body.Add("scope", "repo-read")
		body.Add("scope", "repo-write")
		return body
	}

	// wrong request id
	body := mkBody()
	body.Set("request_id", "doesnotexist")
	request, _ := http.NewRequest("POST", "/oauth/approve", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusNotFound, response.Code)
	}

	// all OK
	request, _ = http.NewRequest("POST", "/oauth/approve", strings.NewReader(mkBody().Encode()))
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

func TestTokenAuthorizationCode(t *testing.T) {
	const codeAlice = "HGZQP6WE"

	mkBody := func(code string) *url.Values {
		body := &url.Values{}
		body.Add("code", code)
		body.Add("grant_type", "authorization_code")
		return body
	}

	handler := InitTestHttpHandler(t)

	// wrong client id
	body := mkBody(codeAlice)
	request, _ := http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("doesnotexist", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong client secret
	body = mkBody(codeAlice)
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "wrongsecret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong code
	body = mkBody("reallywrongcode")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK (with authorization header)
	body = mkBody(codeAlice)
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	responseBody := &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.RefreshToken == nil {
		t.Error("No refresh token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}

	// try to read the same code again
	body = mkBody(codeAlice)
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK (with client credentials in body)
	data.InitTestDb(t)
	body = mkBody(codeAlice)
	body.Add("client_id", "gin")
	body.Add("client_secret", "secret")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	responseBody = &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.RefreshToken == nil {
		t.Error("No refresh token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}
}

func TestTokenRefreshToken(t *testing.T) {
	const refreshTokenAlice = "YYPTDSVZ"

	mkBody := func(refreshToken string) *url.Values {
		body := &url.Values{}
		body.Add("refresh_token", refreshToken)
		body.Add("grant_type", "refresh_token")
		return body
	}

	handler := InitTestHttpHandler(t)

	// wrong client id
	body := mkBody(refreshTokenAlice)
	request, _ := http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("doesnotexist", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong client secret
	body = mkBody(refreshTokenAlice)
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "wrongsecret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong refresh token
	body = mkBody("wrongtoken")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK (with authorization header)
	body = mkBody(refreshTokenAlice)
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("gin", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	responseBody := &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}

	// all OK (with client credentials in body)
	body = mkBody(refreshTokenAlice)
	body.Add("client_id", "gin")
	body.Add("client_secret", "secret")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}

	responseBody = &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}
}

func TestTokenPassword(t *testing.T) {
	mkBody := func(username, password, scope string) *url.Values {
		body := &url.Values{}
		body.Add("password", password)
		body.Add("username", username)
		body.Add("scope", scope)
		body.Add("grant_type", "password")
		return body
	}

	handler := InitTestHttpHandler(t)

	// wrong client id
	body := mkBody("alice", "testtest", "account-read repo-read")
	request, _ := http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("doesnotexist", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong secret
	body = mkBody("alice", "testtest", "account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "wrongsecret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong username
	body = mkBody("doesnotexist", "testtest", "account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong password
	body = mkBody("alice", "wrongpassword", "account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong scope
	body = mkBody("alice", "testtest", "account-read account-write")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK (with authorization header)
	body = mkBody("alice", "testtest", "account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
	responseBody := &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}

	// all OK (with client credentials in body)
	body = mkBody("alice", "testtest", "account-read repo-read")
	body.Add("client_id", "wb")
	body.Add("client_secret", "secret")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
	responseBody = &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}
}

func TestTokenClientCredentials(t *testing.T) {
	mkBody := func(scope string) *url.Values {
		body := &url.Values{}
		body.Add("scope", scope)
		body.Add("grant_type", "client_credentials")
		return body
	}

	handler := InitTestHttpHandler(t)

	// wrong client id
	body := mkBody("account-read repo-read")
	request, _ := http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("doesnotexist", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong secret
	body = mkBody("account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "wrongsecret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong scope
	body = mkBody("account-read account-write")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// all OK (with authorization header)
	body = mkBody("account-read repo-read")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("wb", "secret")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
	responseBody := &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
		t.Error("Token type is supposed to be 'Bearer'")
	}

	// all OK (with client credentials in body)
	body = mkBody("account-read repo-read")
	body.Add("client_id", "wb")
	body.Add("client_secret", "secret")
	request, _ = http.NewRequest("POST", "/oauth/token", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
	responseBody = &tokenResponse{}
	json.Unmarshal(response.Body.Bytes(), responseBody)
	if responseBody.AccessToken == "" {
		t.Error("No access token recieved")
	}
	if responseBody.TokenType != "Bearer" {
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

func TestRegistrationPage(t *testing.T) {
	handler := InitTestHttpHandler(t)

	request, _ := http.NewRequest("GET", "/oauth/registration_page", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}

func TestRegistration(t *testing.T) {
	handler := InitTestHttpHandler(t)

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

	// test that a request without a posted form redirects back to registration page
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

	// test that a request with correct form content redirects to registered_page
	request, _ = http.NewRequest("POST", registrationURL, strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	redirect, err = url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	if redirect.String() != registeredPageURL {
		t.Errorf("Expected to be redirected to '%s', but was '%s'", registeredPageURL, redirect.String())
	}
}

func TestRegisteredPage(t *testing.T) {
	handler := InitTestHttpHandler(t)

	request, _ := http.NewRequest("GET", "/oauth/registered_page", strings.NewReader(""))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusOK, response.Code)
	}
}
