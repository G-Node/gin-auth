// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import "github.com/gorilla/mux"

// RegisterRoutes adds all registered routes for this app to the
// main router. This should make it easier to get a quick overview
// over all routes.
func RegisterRoutes(r *mux.Router) {
	// all for /oauth
	oauth := r.PathPrefix("/oauth").Subrouter()
	oauth.HandleFunc("/authorize", Authorize).Methods("GET")
	oauth.HandleFunc("/login_page", LoginPage).Methods("GET")
	oauth.HandleFunc("/login", Login).Methods("POST")
	oauth.HandleFunc("/approve_page", ApprovePage).Methods("GET")
	oauth.HandleFunc("/approve", Approve).Methods("POST")
	oauth.HandleFunc("/token", Token).Methods("POST")
	oauth.HandleFunc("/validate/{token}", Validate).Methods("GET")
}
