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
	oauth.HandleFunc("/login", Login).
		Methods("POST")
	oauth.HandleFunc("/approve_page", ApprovePage).
		Methods("GET")
	oauth.HandleFunc("/approve", Approve).
		Methods("POST")
	oauth.HandleFunc("/token", Token).
		Methods("POST")
	oauth.HandleFunc("/validate/{token}", Validate).
		Methods("GET")
	// all for /api
	api := r.PathPrefix("/api").Subrouter()
	api.Handle("/accounts", OAuthHandler("account-admin")(http.HandlerFunc(ListAccounts))).
		Methods("GET")
	api.Handle("/accounts/{login}", OAuthHandler("account-read", "account-admin")(http.HandlerFunc(GetAccount))).
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
}
