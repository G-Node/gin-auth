// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
)

// RegisterRoutes adds all registered routes for this app to the
// main router. This should make it easier to get a quick overview
// over all routes.
func RegisterRoutes(r *mux.Router) {
	// all for /oauth
	oauth := r.PathPrefix("/oauth").Subrouter()
	oauth.HandleFunc("/authorize", Authorize).
		Methods("GET")
	oauth.HandleFunc("/login_page", LoginPage).
		Methods("GET")
	oauth.HandleFunc("/login", LoginWithCredentials).
		Methods("POST")
	oauth.HandleFunc("/login", LoginWithSession).
		Methods("GET")
	oauth.HandleFunc("/approve_page", ApprovePage).
		Methods("GET")
	oauth.HandleFunc("/approve", Approve).
		Methods("POST")
	oauth.HandleFunc("/logout/{token}", Logout).
		Methods("GET")
	oauth.HandleFunc("/registration_page", RegistrationPage).Methods("GET")
	oauth.Handle("/registration", RegistrationHandler(captcha.VerifyString)).Methods("POST")
	oauth.HandleFunc("/registered_page", RegisteredPage).Methods("GET")
	oauth.HandleFunc("/activation", Activation).Methods("GET")
	oauth.HandleFunc("/reset_init_page", ResetInitPage).Methods("GET")
	oauth.HandleFunc("/reset_init", ResetInit).Methods("POST")
	oauth.HandleFunc("/reset_page", ResetPage).Methods("GET")
	oauth.HandleFunc("/reset", Reset).Methods("POST")
	oauth.HandleFunc("/token", Token).
		Methods("POST")
	oauth.HandleFunc("/validate/{token}", Validate).
		Methods("GET")
	// all for /api
	api := r.PathPrefix("/api").Subrouter()
	api.Handle("/accounts", OAuthHandlerPermissive()(http.HandlerFunc(ListAccounts))).
		Methods("GET")
	api.Handle("/accounts/{login}", OAuthHandlerPermissive()(http.HandlerFunc(GetAccount))).
		Methods("GET")
	api.Handle("/accounts/{login}", OAuthHandler("account-write", "account-admin")(http.HandlerFunc(UpdateAccount))).
		Methods("PUT")
	api.Handle("/accounts/{login}/password", OAuthHandler("account-write")(http.HandlerFunc(UpdateAccountPassword))).
		Methods("PUT")
	api.Handle("/accounts/{login}/keys", OAuthHandler("account-read", "account-admin")(http.HandlerFunc(ListAccountKeys))).
		Methods("GET")
	api.Handle("/accounts/{login}/keys", OAuthHandler("account-write")(http.HandlerFunc(CreateKey))).
		Methods("POST")
	api.Handle("/keys/{fingerprint}", OAuthHandler("account-read", "account-admin")(http.HandlerFunc(GetKey))).
		Methods("GET")
	api.Handle("/keys/{fingerprint}", OAuthHandler("account-write")(http.HandlerFunc(DeleteKey))).
		Methods("DELETE")

	// captcha service
	cpt := r.PathPrefix("/captcha").Subrouter()
	cpt.Handle("/{id}", captcha.Server(captcha.StdWidth, captcha.StdHeight)).Methods("GET")
	cpt.HandleFunc("/reload/{id}", CaptchaReload).Methods("PUT")
}
