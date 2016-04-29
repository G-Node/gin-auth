// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetAccount returns the requested account as JSON
func GetAccount(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	oauthInfo, ok := OAuthToken(r)
	if ok {
		fmt.Fprintf(w, "login: %s\n", login)
		fmt.Fprintf(w, "match: %s\n", oauthInfo.Match.Strings())
	} else {
		fmt.Fprintf(w, "fail")
	}
}
