// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
	"github.com/gorilla/mux"
)

const (
	cookiePath = "/"
	cookieName = "session"
)

// OAuthInfo provides information about an authorized access token
type OAuthInfo struct {
	Match util.StringSet
	Token *data.AccessToken
}

// OAuthToken gets an access token registered by an OAuthHandler.
func OAuthToken(r *http.Request) (*OAuthInfo, bool) {
	tokens.Lock()
	tok, ok := tokens.store[r]
	tokens.Unlock()

	return tok, ok
}

// Synchronized store for access tokens.
var tokens = struct {
	sync.Mutex
	store map[*http.Request]*OAuthInfo
}{store: make(map[*http.Request]*OAuthInfo)}

// OAuthHandler processes a request and extracts a bearer token from the authorization
// header. If the bearer token is valid and has a matching scope the respective AccessToken
// data can later be obtained using the OAuthToken function.
func OAuthHandler(scope ...string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return oauth{Permissive: false, scope: util.NewStringSet(scope...), handler: handler}
	}
}

// OAuthHandlerPermissive processes a request and extracts a bearer token from the authorization
// header. If the bearer token is valid and has a matching scope the respective AccessToken
// data can later be obtained using the OAuthToken function.
// A permissive handler does not strictly require the presence of a bearer token. In this case
// the request is handled normally but no OAuth information is present in subsequent handlers.
func OAuthHandlerPermissive() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return oauth{Permissive: true, scope: util.NewStringSet(), handler: handler}
	}
}

// The actual OAuth handler
type oauth struct {
	Permissive bool
	scope      util.StringSet
	handler    http.Handler
}

// ServeHTTP implements http.Handler for oauth.
func (o oauth) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if tokenStr := r.Header.Get("Authorization"); tokenStr != "" && strings.HasPrefix(tokenStr, "Bearer ") {
		tokenStr = strings.Trim(tokenStr[6:], " ")

		if token, ok := data.GetAccessToken(tokenStr); ok {
			match := token.Scope

			if !o.Permissive {
				match = match.Intersect(o.scope)
				if match.Len() < 1 {
					PrintErrorJSON(w, r, "Insufficient scope", http.StatusUnauthorized)
					return
				}
			}

			tokens.Lock()
			tokens.store[r] = &OAuthInfo{Match: match, Token: token}
			tokens.Unlock()

			defer func() {
				tokens.Lock()
				delete(tokens.store, r)
				tokens.Unlock()
			}()
		} else if !o.Permissive {
			PrintErrorJSON(w, r, "Invalid bearer token", http.StatusUnauthorized)
			return
		}

	} else if !o.Permissive {
		PrintErrorJSON(w, r, "No bearer token", http.StatusUnauthorized)
		return
	}

	o.handler.ServeHTTP(w, r)
}

// Authorize handles the beginning of an OAuth grant request following the schema
// of 'implicit' or 'code' grant types.
func Authorize(w http.ResponseWriter, r *http.Request) {
	param := &struct {
		ResponseType string
		ClientId     string
		RedirectURI  string
		State        string
		Scope        string
	}{}

	err := util.ReadQueryIntoStruct(r, param, false)
	if err != nil {
		PrintErrorHTML(w, r, err, 400)
		return
	}

	client, ok := data.GetClientByName(param.ClientId)
	if !ok {
		PrintErrorHTML(w, r, fmt.Sprintf("Client '%s' does not exist", param.ClientId), http.StatusBadRequest)
		return
	}

	scope := util.NewStringSet(strings.Split(param.Scope, " ")...)
	request, err := client.CreateGrantRequest(param.ResponseType, param.RedirectURI, param.State, scope)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusBadRequest)
		return
	}

	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, "/oauth/login_page?request_id="+request.Token, http.StatusFound)
}

type loginData struct {
	Login     string
	Password  string
	RequestID string
}

// LoginPage shows a page where the user can enter his credentials.
func LoginPage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query == nil || len(query) == 0 {
		PrintErrorHTML(w, r, "Query parameter 'request_id' was missing", http.StatusBadRequest)
		return
	}

	token := query.Get("request_id")
	request, ok := data.GetGrantRequest(token)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusNotFound)
		return
	}

	// if there is a session cookie redirect to Login
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		_, ok := data.GetSession(cookie.Value)
		if ok {
			w.Header().Add("Cache-Control", "no-store")
			http.Redirect(w, r, "/oauth/login?request_id="+request.Token, http.StatusFound)
			return
		}
	}

	// show login page
	tmpl := conf.MakeTemplate("login.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "layout", &loginData{RequestID: token})
	if err != nil {
		panic(err)
	}
}

// LoginWithCredentials validates user credentials.
func LoginWithCredentials(w http.ResponseWriter, r *http.Request) {
	param := &loginData{}
	err := util.ReadFormIntoStruct(r, param, false)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusBadRequest)
		return
	}

	// look for existing grant request
	request, ok := data.GetGrantRequest(param.RequestID)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusNotFound)
		return
	}

	// verify login data
	account, ok := data.GetAccountByLogin(param.Login)
	if !ok {
		w.Header().Add("Cache-Control", "no-store")
		http.Redirect(w, r, "/oauth/login_page?request_id="+request.Token, http.StatusFound)
		return
	}

	ok = account.VerifyPassword(param.Password)
	if !ok {
		w.Header().Add("Cache-Control", "no-store")
		http.Redirect(w, r, "/oauth/login_page?request_id="+request.Token, http.StatusFound)
		return
	}

	// associate grant request with account
	request.AccountUUID = sql.NullString{String: account.UUID, Valid: true}
	err = request.Update()
	if err != nil {
		panic(err)
	}

	// create session
	session := &data.Session{AccountUUID: account.UUID}
	err = session.Create()
	if err != nil {
		panic(err)
	}

	cookie := &http.Cookie{
		Name:    cookieName,
		Value:   session.Token,
		Path:    cookiePath,
		Expires: session.Expires,
	}
	http.SetCookie(w, cookie)

	// if approved finish the grant request, otherwise redirect to approve page
	if request.IsApproved() {
		if request.GrantType == "code" {
			finishCodeRequest(w, r, request)
		} else {
			finishImplicitRequest(w, r, request)
		}
	} else {
		w.Header().Add("Cache-Control", "no-store")
		http.Redirect(w, r, "/oauth/approve_page?request_id="+request.Token, http.StatusFound)
	}
}

// LoginWithSession validates session cookie.
func LoginWithSession(w http.ResponseWriter, r *http.Request) {
	requestId := r.URL.Query().Get("request_id")
	if requestId == "" {
		PrintErrorHTML(w, r, "Query parameter 'request_id' was missing", http.StatusBadRequest)
		return
	}

	// look for existing grant request
	request, ok := data.GetGrantRequest(requestId)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusNotFound)
		return
	}

	// get session cookie
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		PrintErrorHTML(w, r, "No session cookie provided", http.StatusBadRequest)
		return
	}

	// validate cookie
	session, ok := data.GetSession(cookie.Value)
	if !ok {
		PrintErrorHTML(w, r, "Invalid session cookie", http.StatusNotFound)
		return
	}
	err = session.UpdateExpirationTime()
	if err != nil {
		panic(err)
	}

	account, ok := data.GetAccount(session.AccountUUID)
	if !ok {
		panic("Session has not account")
	}

	// associate grant request with account
	request.AccountUUID = sql.NullString{String: account.UUID, Valid: true}
	err = request.Update()
	if err != nil {
		panic(err)
	}

	cookie = &http.Cookie{
		Name:    cookieName,
		Value:   session.Token,
		Path:    cookiePath,
		Expires: session.Expires,
	}
	http.SetCookie(w, cookie)

	// if approved finish the grant request, otherwise redirect to approve page
	if request.IsApproved() {
		if request.GrantType == "code" {
			finishCodeRequest(w, r, request)
		} else {
			finishImplicitRequest(w, r, request)
		}
	} else {
		w.Header().Add("Cache-Control", "no-store")
		http.Redirect(w, r, "/oauth/approve_page?request_id="+request.Token, http.StatusFound)
	}
}

func finishCodeRequest(w http.ResponseWriter, r *http.Request, request *data.GrantRequest) {
	request.Code = sql.NullString{String: util.RandomToken(), Valid: true}
	err := request.Update()
	if err != nil {
		panic(err)
	}

	scope := url.QueryEscape(strings.Join(request.ScopeRequested.Strings(), " "))
	state := url.QueryEscape(request.State)
	url := fmt.Sprintf("%s?scope=%s&state=%s&code=%s", request.RedirectURI, scope, state, request.Code.String)

	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, url, http.StatusFound)
}

func finishImplicitRequest(w http.ResponseWriter, r *http.Request, request *data.GrantRequest) {
	err := request.Delete()
	if err != nil {
		panic(err)
	}

	token := &data.AccessToken{
		Token:       util.RandomToken(),
		ClientUUID:  request.ClientUUID,
		AccountUUID: request.AccountUUID,
		Scope:       request.ScopeRequested,
	}

	err = token.Create()
	if err != nil {
		panic(err)
	}

	scope := url.QueryEscape(strings.Join(token.Scope.Strings(), " "))
	state := url.QueryEscape(request.State)
	url := fmt.Sprintf("%s?token_type=bearer&scope=%s&state=%s&access_token=%s", request.RedirectURI, scope, state, token.Token)

	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, url, http.StatusFound)
}

// Logout remove a valid token (and if present the session cookie too) so it can't be used any more.
func Logout(w http.ResponseWriter, r *http.Request) {
	tokenStr := mux.Vars(r)["token"]
	if token, ok := data.GetAccessToken(tokenStr); ok {
		if err := token.Delete(); err != nil {
			panic(err)
		}
	} else {
		PrintErrorHTML(w, r, "Access token does not exist", http.StatusNotFound)
		return
	}

	cookie, err := r.Cookie(cookieName)
	if err == nil {
		if session, ok := data.GetSession(cookie.Value); ok {
			if err := session.Delete(); err != nil {
				panic(err)
			}
		}
	}

	uri := r.URL.Query().Get("redirect_uri")
	if uri != "" {
		w.Header().Add("Cache-Control", "no-store")
		http.Redirect(w, r, uri, http.StatusFound)
	} else {
		pageData := struct {
			Header  string
			Message string
		}{"You successfully signed out!", ""}

		tmpl := conf.MakeTemplate("success.html")
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Content-Type", "text/html")
		err := tmpl.ExecuteTemplate(w, "layout", pageData)
		if err != nil {
			panic(err)
		}
	}
}

// ApprovePage shows a page where the user can approve client access.
func ApprovePage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query == nil || len(query) == 0 {
		PrintErrorHTML(w, r, "Query parameter 'request_id' was missing", http.StatusBadRequest)
		return
	}
	token := query.Get("request_id")

	request, ok := data.GetGrantRequest(token)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusNotFound)
		return
	}
	if !request.AccountUUID.Valid {
		PrintErrorHTML(w, r, "Grant request is not authenticated", http.StatusUnauthorized)
		return
	}

	client := request.Client()
	scope := request.ScopeRequested.Difference(client.ScopeWhitelist)
	var existScope, addScope map[string]string
	if approval, ok := client.ApprovalForAccount(request.AccountUUID.String); ok && approval.Scope.Len() > 0 {
		existScope, ok = data.DescribeScope(approval.Scope)
		if !ok {
			panic("Invalid scope")
		}
		addScope, ok = data.DescribeScope(scope.Difference(approval.Scope))
		if !ok {
			panic("Invalid scope")
		}
	} else {
		addScope, ok = data.DescribeScope(scope)
		if !ok {
			panic("Invalid scope")
		}
	}

	pageData := struct {
		Client        string
		AddScope      map[string]string
		ExistingScope map[string]string
		RequestID     string
	}{client.Name, addScope, existScope, request.Token}

	tmpl := conf.MakeTemplate("approve.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", pageData)
	if err != nil {
		panic(err)
	}
}

// Approve evaluates an access approval given to a certain client.
func Approve(w http.ResponseWriter, r *http.Request) {
	param := &struct {
		Client    string
		RequestID string
		Scope     []string
	}{}
	util.ReadFormIntoStruct(r, param, true)

	request, ok := data.GetGrantRequest(param.RequestID)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusNotFound)
		return
	}

	if !request.AccountUUID.Valid {
		PrintErrorHTML(w, r, "Grant request is not authenticated", http.StatusUnauthorized)
		return
	}

	client := request.Client()

	scopeApproved := util.NewStringSet(param.Scope...)
	scopeRequired := request.ScopeRequested.Difference(client.ScopeWhitelist)
	if !scopeApproved.IsSuperset(scopeRequired) {
		PrintErrorHTML(w, r, "Requested scope was not approved", http.StatusUnauthorized)
		return
	}

	// create approval
	err := client.Approve(request.AccountUUID.String, request.ScopeRequested)
	if err != nil {
		panic(err)
	}

	// if approved finish the grant request
	if !request.IsApproved() {
		panic("Requested scope should be approved but was not")
	}

	if request.GrantType == "code" {
		finishCodeRequest(w, r, request)
	} else {
		finishImplicitRequest(w, r, request)
	}
}

type tokenResponse struct {
	TokenType    string  `json:"token_type"`
	Scope        string  `json:"scope"`
	AccessToken  string  `json:"access_token"`
	RefreshToken *string `json:"refresh_token"`
}

// Token exchanges a grant code for an access and refresh token
func Token(w http.ResponseWriter, r *http.Request) {
	// Read authorization header
	clientId, clientSecret, authorizeOk := r.BasicAuth()

	// Parse request body
	body := &struct {
		GrantType    string
		ClientId     string
		ClientSecret string
		Scope        string
		Code         string
		RefreshToken string
		Username     string
		Password     string
	}{}
	err := util.ReadFormIntoStruct(r, body, true)
	if err != nil {
		PrintErrorJSON(w, r, err, 400)
		return
	}

	// Take clientId and clientSecret from body if they are not in the header
	if !authorizeOk {
		clientId = body.ClientId
		clientSecret = body.ClientSecret
	}

	// Check client
	client, ok := data.GetClientByName(clientId)
	if !ok {
		PrintErrorJSON(w, r, "Wrong client id or client secret", http.StatusUnauthorized)
		return
	}
	if clientSecret != client.Secret {
		PrintErrorJSON(w, r, "Wrong client id or client secret", http.StatusUnauthorized)
		return
	}

	// Prepare a response depending on the grant type
	var response *tokenResponse
	switch body.GrantType {

	case "authorization_code":
		request, ok := data.GetGrantRequestByCode(body.Code)
		if !ok {
			PrintErrorJSON(w, r, "Invalid grant code", http.StatusUnauthorized)
			return
		}
		if request.ClientUUID != client.UUID {
			request.Delete()
			PrintErrorJSON(w, r, "Invalid grant code", http.StatusUnauthorized)
			return
		}

		access, refresh, err := request.ExchangeCodeForTokens()
		if err != nil {
			PrintErrorJSON(w, r, "Invalid grant code", http.StatusUnauthorized)
			return
		}

		response = &tokenResponse{
			TokenType:    "Bearer",
			Scope:        strings.Join(request.ScopeRequested.Strings(), " "),
			AccessToken:  access,
			RefreshToken: &refresh,
		}

	case "refresh_token":
		refresh, ok := data.GetRefreshToken(body.RefreshToken)
		if !ok {
			PrintErrorJSON(w, r, "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		if refresh.ClientUUID != client.UUID {
			refresh.Delete()
			PrintErrorJSON(w, r, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		access := data.AccessToken{
			Token:       util.RandomToken(),
			AccountUUID: sql.NullString{String: refresh.AccountUUID, Valid: true},
			ClientUUID:  refresh.ClientUUID,
			Scope:       refresh.Scope,
		}
		err := access.Create()
		if err != nil {
			PrintErrorJSON(w, r, err, http.StatusInternalServerError)
			return
		}

		response = &tokenResponse{
			TokenType:   "Bearer",
			Scope:       strings.Join(refresh.Scope.Strings(), " "),
			AccessToken: access.Token,
		}

	case "password":
		account, ok := data.GetAccountByLogin(body.Username)
		if !ok {
			PrintErrorJSON(w, r, "Wrong username or password", http.StatusUnauthorized)
			return
		}
		if !account.VerifyPassword(body.Password) {
			PrintErrorJSON(w, r, "Wrong username or password", http.StatusUnauthorized)
			return
		}

		scope := util.NewStringSet(strings.Split(body.Scope, " ")...)
		if scope.Len() == 0 || !client.ScopeWhitelist.IsSuperset(scope) {
			PrintErrorJSON(w, r, "Invalid scope", http.StatusUnauthorized)
			return
		}

		access := data.AccessToken{
			Token:       util.RandomToken(),
			AccountUUID: sql.NullString{String: account.UUID, Valid: true},
			ClientUUID:  client.UUID,
			Scope:       scope,
		}
		err := access.Create()
		if err != nil {
			PrintErrorJSON(w, r, err, http.StatusInternalServerError)
			return
		}

		response = &tokenResponse{
			TokenType:   "Bearer",
			Scope:       strings.Join(scope.Strings(), " "),
			AccessToken: access.Token,
		}

	case "client_credentials":
		scope := util.NewStringSet(strings.Split(body.Scope, " ")...)
		if scope.Len() == 0 || !client.ScopeWhitelist.IsSuperset(scope) {
			PrintErrorJSON(w, r, "Invalid scope", http.StatusUnauthorized)
			return
		}

		access := data.AccessToken{
			Token:      util.RandomToken(),
			ClientUUID: client.UUID,
			Scope:      scope,
		}
		err := access.Create()
		if err != nil {
			PrintErrorJSON(w, r, err, http.StatusInternalServerError)
			return
		}

		response = &tokenResponse{
			TokenType:   "Bearer",
			Scope:       strings.Join(scope.Strings(), " "),
			AccessToken: access.Token,
		}

	default:
		PrintErrorJSON(w, r, fmt.Sprintf("Unsupported grant type %s", body.GrantType), http.StatusBadRequest)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(response)
}

// Validate validates a token and returns information about it as JSON
func Validate(w http.ResponseWriter, r *http.Request) {
	tokenStr := mux.Vars(r)["token"]
	token, ok := data.GetAccessToken(tokenStr)
	if !ok {
		PrintErrorJSON(w, r, "The requested token does not exist", http.StatusNotFound)
		return
	}

	var login, accountUrl *string
	if token.AccountUUID.Valid {
		if account, ok := data.GetAccount(token.AccountUUID.String); ok {
			login = &account.Login
			accountUrl = new(string)
			(*accountUrl) = conf.MakeUrl("/api/accounts/%s", account.Login)
		} else {
			PrintErrorJSON(w, r, "Unable to find account associated with the request", http.StatusInternalServerError)
			return
		}
	}

	scope := strings.Join(token.Scope.Strings(), " ")
	response := &struct {
		URL        string    `json:"url"`
		JTI        string    `json:"jti"`
		EXP        time.Time `json:"exp"`
		ISS        string    `json:"iss"`
		Login      *string   `json:"login"`
		AccountURL *string   `json:"account_url"`
		Scope      string    `json:"scope"`
	}{
		URL:        conf.MakeUrl("/oauth/validate/%s", token.Token),
		JTI:        token.Token,
		EXP:        token.Expires,
		ISS:        "gin-auth",
		Login:      login,
		AccountURL: accountUrl,
		Scope:      scope,
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(response)
}

type validateAccount struct {
	*data.Account
	*util.ValidationError
}

// RegistrationPage displays entry fields required for the creation of a new gin account
func RegistrationPage(w http.ResponseWriter, r *http.Request) {
	valAccount := &validateAccount{}
	valAccount.Account = &data.Account{}
	valAccount.ValidationError = &util.ValidationError{}

	tmpl := conf.MakeTemplate("registration.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", valAccount)
	if err != nil {
		panic(err)
	}
}

type passwordData struct {
	Password        string
	PasswordControl string
}

// Registration parses user entries for a new account. It will redirect back to the
// entry form, if input is invalid. If the input is correct, it will create a new account,
// send an e-mail with an activation link and redirect to the the registered page.
func Registration(w http.ResponseWriter, r *http.Request) {
	tmpl := conf.MakeTemplate("registration.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	account := &data.Account{}
	pw := &passwordData{}

	err := util.ReadFormIntoStruct(r, account, true)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusInternalServerError)
		return
	}

	err = util.ReadFormIntoStruct(r, pw, true)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusInternalServerError)
		return
	}

	valAccount := &validateAccount{}
	valAccount.ValidationError = &util.ValidationError{}
	valAccount.Account = account

	if r.Form.Encode() == "" {
		valAccount.Message = "Please add all required fields (*)"
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	valAccount.ValidationError = valAccount.Account.Validate()

	if pw.Password != pw.PasswordControl {
		valAccount.FieldErrors["password"] = "Provided password did not match password control"
		if valAccount.Message == "" {
			valAccount.Message = "Provided password did not match password control"
		}
	}
	if pw.Password == "" || pw.PasswordControl == "" {
		valAccount.FieldErrors["password"] = "Please enter password and password control"
		if valAccount.Message == "" {
			valAccount.Message = "Please enter password and password control"
		}
	}
	if len(pw.Password) > 512 || len(pw.PasswordControl) > 512 {
		valAccount.FieldErrors["password"] =
			fmt.Sprintf("Entry too long, please shorten to %d characters", 512)
	}

	if valAccount.Message != "" {
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	valAccount.Account.SetPassword(pw.Password)
	valAccount.Account.ActivationCode = sql.NullString{String: util.RandomToken(), Valid: true}

	err = account.Create()
	if err != nil {
		fmt.Printf("Error: Registration failed due to error: '%s'\n", err)

		valAccount.Message = "An error occurred during registration."
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	fmt.Printf("Registration of user '%s' (%s) successful\n", valAccount.Account.Login, valAccount.Account.Email)
	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, "/oauth/registered_page", http.StatusFound)
}

// RegisteredPage displays information about how a newly created gin account can be activated.
func RegisteredPage(w http.ResponseWriter, r *http.Request) {
	head := "Account registered"
	message := "Your account activation is pending. "
	message += "An e-mail with an activation code has been sent to your e-mail address."

	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", info)
	if err != nil {
		panic(err)
	}
}

// Activation removes an existing activation code from an account, thus rendering the account active.
func Activation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PrintErrorHTML(w, r, "Activation request was malformed", http.StatusBadRequest)
		return
	}

	getCode := r.Form.Get("activation_code")
	if getCode == "" {
		PrintErrorHTML(w, r, "Account activation code was absent", http.StatusBadRequest)
		return
	}

	account, exists := data.GetAccountByActivationCode(getCode)
	if !exists {
		PrintErrorHTML(w, r, "Requested account does not exist", http.StatusNotFound)
		return
	}

	account.ActivationCode.Valid = false
	err = account.Update()
	if err != nil {
		panic(err)
	}

	head := "Account activation"
	message := fmt.Sprintf("Congratulation %s %s! The account for %s has been activated and can now be used.",
		account.FirstName, account.LastName, account.Login)
	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	err = tmpl.ExecuteTemplate(w, "layout", info)
	if err != nil {
		panic(err)
	}
}

type credentialData struct {
	Credential string
	ErrMessage string
}

// ResetInitPage provides an input form for resetting an account password
func ResetInitPage(w http.ResponseWriter, r *http.Request) {

	credData := &credentialData{}

	tmpl := conf.MakeTemplate("resetinit.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", credData)
	if err != nil {
		panic(err)
	}
}

// ResetInit checks whether a provided login or e-mail address
// belongs to a non-disabled account. If this is the case, the corresponding
// account is updated with a password reset code and an email containing
// the code is sent to the e-mail address of the account.
func ResetInit(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	credData := &credentialData{}

	err := util.ReadFormIntoStruct(r, credData, true)
	if err != nil {
		panic(err)
	}

	if credData.Credential == "" {
		credData.ErrMessage = "Please enter your login or e-mail address"
		tmpl := conf.MakeTemplate("resetinit.html")
		w.Header().Add("Warning", credData.ErrMessage)
		err = tmpl.ExecuteTemplate(w, "layout", credData)
		if err != nil {
			panic(err)
		}
		return
	}

	account, ok := data.SetPasswordReset(credData.Credential)
	if !ok {
		credData.ErrMessage = "Invalid login or e-mail address"
		tmpl := conf.MakeTemplate("resetinit.html")
		w.Header().Add("Warning", credData.ErrMessage)
		err = tmpl.ExecuteTemplate(w, "layout", credData)
		if err != nil {
			panic(err)
		}
		return
	}

	fmt.Printf("Update pw code '%s' of account with login '%s' and email '%s'\n",
		account.ResetPWCode.String, account.Login, account.Email)

	head := "Success!"
	message := "An e-mail with a password reset token has been sent to your e-mail address. "
	message += "Please follow the contained link to reset your password. "
	message += "Please note that your account will stay deactivated until your password reset has been completed."
	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	err = tmpl.ExecuteTemplate(w, "layout", info)
}
