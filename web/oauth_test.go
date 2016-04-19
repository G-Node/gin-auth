package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"encoding/json"
	"github.com/G-Node/gin-auth/data"
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

func newAuthQuery() url.Values {
	query := url.Values{}
	query.Add("response_type", "code")
	query.Add("client_id", "gin")
	query.Add("redirect_uri", "https://localhost:8081/login")
	query.Add("scope", "repo-read,repo-write")
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

	// wrong response type
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
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
	}

	// wrong password
	body = newLoginBody()
	body.Set("password", "notapassword")
	request, _ = http.NewRequest("POST", "/oauth/login", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Response code '%d' expected but was '%d'", http.StatusUnauthorized, response.Code)
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

func newApproveBody() *url.Values {
	body := &url.Values{}
	body.Add("request_id", "U7JIKKYI")
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
	data := &responseTokenData{}
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
